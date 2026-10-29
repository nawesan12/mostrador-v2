/**
 * Importar desde Tiendanube con su API oficial (la app de socio de Mostrador, con OAuth). Gemelo en
 * Go: `apps/api/internal/integraciones/tiendanube/` (api.go, convertir.go y clientes.go). Los dos
 * pasan el caso de `packages/contrato/casos/importar/tiendanube.json`; si cambia una regla, cambia en
 * los dos lados y se regenera lo esperado con `go test -run TestCaso -actualizar`.
 *
 *   - La tienda autoriza la app (OAuth sin PKCE; el state lo guarda quien llama). El token no vence.
 *   - Se lee de a páginas siguiendo el header Link (sólo URLs de la API, para no mandar el token a
 *     otro lado). Ante 429 se espera lo que dice `x-rate-limit-reset`; un 401 es "conectala de nuevo".
 *   - Los textos vienen por idioma ({"es": …}) y las fechas como `+0000`.
 */
import { ajustar, textoRico, type Crudo } from './comun.ts';
import type { ProductoEditable } from '../datos/tipos.ts';

const VERSION = '2025-03';
export const BASE = `https://api.tiendanube.com/${VERSION}`;
/**
 * En las pruebas de punta a punta, la API y el canje van a un Tiendanube falso en localhost (la
 * variable TIENDANUBE_LOCAL del preview). Sólo se acepta http://localhost o 127.0.0.1 con puerto.
 */
export const localDe = (v: unknown) =>
	typeof v === 'string' && /^http:\/\/(localhost|127\.0\.0\.1):\d{2,5}$/.test(v) ? v : null;
const AUTORIZAR = 'https://www.tiendanube.com/apps/%s/authorize';
const TOKEN = 'https://www.tiendanube.com/apps/authorize/token';
const USER_AGENT = 'Mostrador (soporte@mostradorapp.com.ar)';
/** Lo mismo que el importador de Empretienda cuando la tienda no lleva el stock. */
export const STOCK_ILIMITADO = 999;
/** Tope de lo que se lee de una página (en el Worker la memoria es poca). */
const TAMANIO_MAX = 5_000_000;

// ---------------------------------------------------------------- Texto como Go

// Los espacios de `unicode.IsSpace` de Go (no los de `\s` de JS, que suma U+FEFF y no U+0085).
const ESP =
	'\\t\\n\\v\\f\\r \\u0085\\u00a0\\u1680\\u2000-\\u200a\\u2028\\u2029\\u202f\\u205f\\u3000';
const PUNTAS = new RegExp(`^[${ESP}]+|[${ESP}]+$`, 'g');
const TRAMOS = new RegExp(`[${ESP}]+`, 'g');
/** strings.TrimSpace. */
const recortar = (s: string) => s.replace(PUNTAS, '');
/** strings.Join(strings.Fields(s), " "). */
const espacios = (s: string) => recortar(s).replace(TRAMOS, ' ');
/** A lo sumo n letras (runas), sin espacios de más. */
function corto(s: string, n: number) {
	const t = espacios(s);
	const r = [...t];
	return r.length > n ? recortar(r.slice(0, n).join('')) : t;
}
const valor = (s: unknown) => (typeof s === 'string' ? recortar(s) : '');

/** Un campo por idioma. Tiendanube los manda así aunque la tienda tenga un solo idioma. */
export type Texto = Record<string, string> | null | undefined;

/** El texto en el idioma pedido; si no, en español; si no, el primero que haya (por clave). */
export function textoEn(t: Texto, idioma: string): string {
	if (!t || typeof t !== 'object') return '';
	const hay = (k: string) => typeof t[k] === 'string' && recortar(t[k]) !== '';
	if (hay(idioma)) return t[idioma];
	if (hay('es')) return t.es;
	const claves = Object.keys(t).filter(hay).sort();
	return claves.length ? t[claves[0]] : '';
}

/** "18900.50" a pesos enteros, redondeando. Vacío o inválido es 0 (como `pesos` de Go). */
function pesosTN(s: unknown): number {
	if (typeof s !== 'string' && typeof s !== 'number') return 0;
	const t = recortar(String(s));
	if (!/^[+-]?(\d+\.?\d*|\.\d+)([eE][+-]?\d+)?$/.test(t)) return 0;
	const f = Number(t);
	return Number.isFinite(f) && f >= 0 ? Math.round(f) : 0;
}

// ---------------------------------------------------------------- Productos

type Variante = {
	image_id?: number | null;
	price?: string | null;
	promotional_price?: string | null;
	stock_management?: boolean;
	stock?: number | null;
	sku?: string | null;
	values?: Texto[];
};
export type ProductoTN = {
	id: number;
	name?: Texto;
	description?: Texto;
	handle?: Texto;
	attributes?: Texto[];
	categories?: { id: number; name?: Texto; parent?: number | null }[];
	images?: { id: number; src?: string; position?: number }[];
	variants?: Variante[];
};

/** Un producto listo para guardar, con su id en Tiendanube y el path de su ficha vieja (301). */
export type ImportadoTN = {
	externo: string;
	urlVieja: string;
	producto: ProductoEditable;
	recortado: boolean;
};

/** Un producto de la API → producto de Mostrador (null sin nombre o sin variantes). */
export function productoTN(p: ProductoTN, idioma: string): ImportadoTN | null {
	const variantes = p.variants ?? [];
	if (!variantes.length) return null;
	// La categoría más específica: una que no sea madre de otra de las que trae el producto.
	const cats = p.categories ?? [];
	const madres = new Set(cats.flatMap((c) => (c.parent == null ? [] : [c.parent])));
	const cat = cats.find((c) => !madres.has(c.id));
	// Fotos en el orden de la tienda, en https (el CDN también las sirve así).
	const imgs = [...(p.images ?? [])].sort((a, b) => (a.position ?? 0) - (b.position ?? 0));
	const fotoDe = new Map<number, string>();
	const fotos = imgs.map((im) => {
		let src = recortar(im.src ?? '');
		if (src.startsWith('http://')) src = 'https://' + src.slice(7);
		fotoDe.set(im.id, src);
		return src;
	});
	// Precio (el de oferta si es menor) y stock (sin stock_management, ilimitado).
	const vars = variantes.map((v) => {
		let precio = pesosTN(v.price);
		const oferta = pesosTN(v.promotional_price);
		if (oferta > 0 && (precio === 0 || oferta < precio)) precio = oferta;
		const stock =
			v.stock_management && typeof v.stock === 'number'
				? Math.max(0, Math.trunc(v.stock))
				: STOCK_ILIMITADO;
		return {
			precio,
			stock,
			sku: typeof v.sku === 'string' ? recortar(v.sku) : '',
			foto: (v.image_id != null && fotoDe.get(v.image_id)) || null,
			valores: (v.values ?? []).map((x) => textoEn(x, idioma))
		};
	});
	const opciones: Crudo['opciones'] = [];
	// Sin atributos (o con una sola variante sin valores) es un producto simple.
	const atributos = p.attributes ?? [];
	if (atributos.length && vars[0].valores.length)
		atributos.forEach((a, i) => {
			const valores: string[] = [];
			const vistos = new Set<string>();
			for (const v of vars) {
				const x = v.valores[i];
				if (x && !vistos.has(x.toLowerCase())) {
					vistos.add(x.toLowerCase());
					valores.push(x);
				}
			}
			opciones.push({ nombre: recortar(textoEn(a, idioma)) || `Opción ${i + 1}`, valores });
		});
	const crudo: Crudo = {
		nombre: textoEn(p.name, idioma),
		descripcion: textoRico(textoEn(p.description, idioma)),
		categoria: cat ? recortar(textoEn(cat.name, idioma)) : null,
		precio: Math.min(...vars.map((v) => v.precio)),
		sku: vars[0].sku,
		stock: vars[0].stock,
		estimado: false,
		fotos,
		opciones,
		variantes: opciones.length ? vars : []
	};
	const i = ajustar(crudo);
	if (!i) return null;
	const h = textoEn(p.handle, idioma).replace(/^[/ ]+|[/ ]+$/g, '');
	return {
		externo: String(p.id),
		urlVieja: h ? `/productos/${h}/` : '',
		producto: i.producto,
		recortado: i.recortado
	};
}

// ---------------------------------------------------------------- Clientes y pedidos

type Direccion = { city?: string | null; locality?: string | null; province?: string | null };
export type ClienteTN = {
	id: number;
	name?: string;
	email?: string | null;
	phone?: string | null;
	note?: string | null;
	default_address?: Direccion | null;
	billing_city?: string | null;
	billing_province?: string | null;
};
export type Cliente = {
	externo: string;
	nombre: string;
	email: string;
	telefono: string;
	ciudad: string;
	provincia: string;
	notas: string;
};

/** Un cliente de la API → la ficha de Mostrador. Sin nombre ni email no hay a quién guardar. */
export function clienteTN(c: ClienteTN): Cliente | null {
	const r: Cliente = {
		externo: String(c.id),
		nombre: corto(c.name ?? '', 80),
		email: corto(valor(c.email), 120).toLowerCase(),
		telefono: corto(valor(c.phone), 30),
		ciudad: '',
		provincia: '',
		notas: corto(valor(c.note), 500)
	};
	const d = c.default_address;
	if (d) {
		r.ciudad = corto(valor(d.city) || valor(d.locality), 60);
		r.provincia = corto(valor(d.province), 60);
	}
	r.ciudad ||= corto(valor(c.billing_city), 60);
	r.provincia ||= corto(valor(c.billing_province), 60);
	r.nombre ||= r.email;
	if (!r.nombre || (r.email && !r.email.includes('@'))) return null;
	return r;
}

export type PedidoTN = {
	id: number;
	number?: number;
	created_at?: string;
	total?: string | null;
	currency?: string;
	status?: string;
	payment_status?: string;
	contact_email?: string | null;
	contact_name?: string | null;
	customer?: { name?: string; email?: string | null } | null;
	products?: {
		name?: string;
		price?: string | null;
		quantity?: number | string;
		variant_values?: unknown[];
	}[];
};
export type Pedido = {
	externo: string;
	numeroExterno: string;
	/** ISO como lo escribe Go (sin milésimas en cero). */
	fecha: string;
	clienteEmail: string;
	clienteNombre: string;
	total: number;
	estado: string;
	items: { nombre: string; variante: string; cantidad: number; precio: number }[];
};

const PAGOS = new Set(['paid', 'partially_paid', 'partially_refunded', 'authorized']);

/**
 * Las fechas de la API: RFC 3339 o con la zona sin dos puntos ("2022-11-15T19:36:59+0000"). En ms
 * desde 1970, o null si no es una fecha.
 */
export function fechaTN(s: string | undefined): number | null {
	const m = recortar(s ?? '').match(
		/^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2})(?:[.,](\d{1,9}))?(Z|[+-]\d{2}:\d{2}|[+-]\d{4})$/
	);
	if (!m) return null;
	const [a, me, d, h, mi, se] = m.slice(1, 7).map(Number);
	if (me < 1 || me > 12 || h > 23 || mi > 59 || se > 59) return null;
	const local = Date.UTC(a, me - 1, d, h, mi, se, Number((m[7] ?? '0').padEnd(3, '0').slice(0, 3)));
	if (new Date(local).getUTCDate() !== d) return null;
	if (m[8] === 'Z') return local;
	const z = m[8].replace(':', '');
	const off = (Number(z.slice(1, 3)) * 60 + Number(z.slice(3, 5))) * (z[0] === '-' ? -1 : 1);
	return local - off * 60_000;
}

/** Como `time.Time` de Go en JSON: "2026-03-02T17:10:00Z" (las milésimas sólo si hay). */
const isoGo = (ms: number) => new Date(ms).toISOString().replace(/\.?0+Z$/, 'Z');

/** Cantidad: número (se trunca) o texto entero ("1"). */
function entero(v: unknown): number {
	if (typeof v === 'number') return Number.isFinite(v) ? Math.trunc(v) : 0;
	if (typeof v === 'string' && /^[+-]?\d+$/.test(recortar(v))) return Number(recortar(v));
	return 0;
}

/**
 * Un pedido de la API → el historial del cliente. Sólo los pagos en pesos: los cancelados y los que
 * no se cobraron no cuentan.
 */
export function pedidoTN(p: PedidoTN): Pedido | null {
	if (p.currency && p.currency !== 'ARS') return null;
	if (!PAGOS.has(p.payment_status ?? '')) return null;
	const fecha = fechaTN(p.created_at);
	if (p.status === 'cancelled' || fecha === null) return null;
	const r = {
		externo: String(p.id),
		numeroExterno: String(p.number ?? 0),
		fecha: isoGo(fecha),
		clienteEmail: valor(p.contact_email).toLowerCase(),
		clienteNombre: corto(valor(p.contact_name), 80),
		total: pesosTN(p.total),
		estado: p.status ?? '',
		items: (p.products ?? []).map((it) => ({
			nombre: corto(it.name ?? '', 120),
			variante: corto(
				(it.variant_values ?? [])
					.flatMap((v) => (typeof v === 'string' && recortar(v) ? [recortar(v)] : []))
					.join(' / '),
				60
			),
			cantidad: Math.max(1, entero(it.quantity)),
			precio: pesosTN(it.price)
		}))
	};
	if (p.customer) {
		r.clienteEmail ||= valor(p.customer.email).toLowerCase();
		r.clienteNombre ||= corto(p.customer.name ?? '', 80);
	}
	if (!r.clienteEmail && !r.clienteNombre) return null;
	return r;
}

/** Para el caso compartido con Go: lo mismo que arma `TestCaso`. */
export function resultadoTiendanube(json: {
	productos: ProductoTN[];
	clientes: ClienteTN[];
	pedidos: PedidoTN[];
}) {
	return {
		productos: json.productos.map((p) => productoTN(p, 'es')).filter((x) => !!x),
		clientes: json.clientes.map(clienteTN).filter((x) => !!x),
		pedidos: json.pedidos.map(pedidoTN).filter((x) => !!x)
	};
}

// ---------------------------------------------------------------- OAuth y API

/** La credencial no sirve más (desinstalaron la app o la revocaron): hay que conectar de nuevo. */
export class ErrorReconectar extends Error {
	constructor() {
		super('Tiendanube ya no deja leer la tienda: conectala de nuevo.');
	}
}
/** Tiendanube sigue pidiendo esperar después de los reintentos: se prueba en el próximo paso. */
export class ErrorLimite extends Error {
	constructor() {
		super('Tiendanube pidió esperar.');
	}
}

/** La app de socio (TIENDANUBE_CLIENT_ID y TIENDANUBE_SECRETO). */
export type AppTN = {
	id: string;
	secreto: string;
	fetch?: typeof fetch;
	/** Un Tiendanube falso para las pruebas (`localDe`). */
	local?: string | null;
};
export type TokenTN = { acceso: string; tiendaId: string; alcance: string[] };

/** Adónde se manda a la dueña para que autorice la app. */
export const urlAutorizar = (app: AppTN, state: string) =>
	AUTORIZAR.replace('%s', encodeURIComponent(app.id)) + '?state=' + encodeURIComponent(state);

/** Cambia el código de la vuelta (dura 5 minutos) por el token de la tienda. Null si no se pudo. */
export async function canjear(app: AppTN, codigo: string): Promise<TokenTN | null> {
	const f = app.fetch ?? fetch;
	const r = await f(app.local ? `${app.local}/apps/authorize/token` : TOKEN, {
		method: 'POST',
		headers: { 'content-type': 'application/json', 'user-agent': USER_AGENT },
		body: JSON.stringify({
			client_id: app.id,
			client_secret: app.secreto,
			grant_type: 'authorization_code',
			code: codigo
		}),
		signal: AbortSignal.timeout(10_000)
	}).catch(() => null);
	if (!r) return null;
	const b = (await r.json().catch(() => null)) as {
		access_token?: string;
		scope?: string;
		user_id?: number | string;
	} | null;
	// El id de la tienda llega como número o como texto, según la respuesta.
	const id = String(b?.user_id ?? '')
		.trim()
		.replace(/\.0$/, '');
	if (!r.ok || !b?.access_token || !id) return null;
	return { acceso: b.access_token, tiendaId: id, alcance: (b.scope ?? '').split(',') };
}

export type PaginaTN = {
	registros: unknown[];
	siguiente: string;
	/** Idioma principal de la tienda (X-Main-Language), para los campos por idioma. */
	idioma: string;
	total: number | null;
};

export type ApiTN = {
	/** Un Tiendanube falso para las pruebas (`localDe`). */
	local?: string | null;
	token: string;
	tiendaId: string;
	fetch?: typeof fetch;
	/** Espera entre reintentos; los tests la reemplazan. */
	dormir?: (ms: number) => Promise<void>;
	/** Tope de cada espera ante un 429 (en el Worker no conviene esperar mucho). */
	esperaMax?: number;
	reintentos?: number;
};

/** La URL de la primera página de un recurso (products, customers, orders). */
export const primera = (api: ApiTN, recurso: string, porPagina: number) =>
	`${baseDe(api)}/${encodeURIComponent(api.tiendaId)}/${recurso}?page=1&per_page=${porPagina}`;

const baseDe = (api: ApiTN) => (api.local ? `${api.local}/${VERSION}` : BASE);

const dormir = (ms: number) => new Promise<void>((r) => setTimeout(r, ms));

/** Cuánto esperar después de un 429: lo que falta para que se vacíe el balde (ms), con tope. */
function espera(h: Headers, max: number) {
	const ms = Number(h.get('x-rate-limit-reset'));
	return Number.isFinite(ms) && ms > 0 ? Math.min(ms, max) : Math.min(1000, max);
}

const reLink = /<([^>]+)>;\s*rel="next"/;

/**
 * Lee una página. Con 429 espera y reintenta; si sigue, ErrorLimite. Con 401 o 403, ErrorReconectar.
 * Sólo sigue URLs de la API de esa tienda: el header Link no puede mandar el token a otro lado.
 */
export async function leer(api: ApiTN, url: string): Promise<PaginaTN> {
	const prefijo = `${baseDe(api)}/${encodeURIComponent(api.tiendaId)}/`;
	if (!url.startsWith(prefijo)) throw new Error('Tiendanube: dirección fuera de la API.');
	const f = api.fetch ?? fetch;
	const reintentos = api.reintentos ?? 3;
	for (let intento = 0; ; intento++) {
		const r = await f(url, {
			headers: { authorization: `Bearer ${api.token}`, 'user-agent': USER_AGENT },
			redirect: 'manual',
			signal: AbortSignal.timeout(15_000)
		});
		if (r.status === 429) {
			if (intento >= reintentos) throw new ErrorLimite();
			await (api.dormir ?? dormir)(espera(r.headers, api.esperaMax ?? 30_000));
			continue;
		}
		if (r.status === 401 || r.status === 403) throw new ErrorReconectar();
		// Una página después de la última responde 404 ("Last page is 0"): no hay más.
		if (r.status === 404) return { registros: [], siguiente: '', idioma: '', total: null };
		if (r.status !== 200) throw new Error(`Tiendanube respondió HTTP ${r.status}.`);
		const texto = await r.text();
		if (texto.length > TAMANIO_MAX)
			throw new Error('Tiendanube mandó una página demasiado grande.');
		let registros: unknown;
		try {
			registros = JSON.parse(texto);
		} catch {
			registros = null;
		}
		if (!Array.isArray(registros)) throw new Error('Tiendanube respondió algo que no entendemos.');
		const sig = r.headers.get('link')?.match(reLink)?.[1] ?? '';
		const total = Number(r.headers.get('x-total-count'));
		return {
			registros,
			siguiente: sig.startsWith(prefijo) ? sig : '',
			idioma: r.headers.get('x-main-language') ?? '',
			total: r.headers.has('x-total-count') && Number.isFinite(total) ? total : null
		};
	}
}
