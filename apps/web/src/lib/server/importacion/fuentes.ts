/**
 * Las fuentes reales de la demo, como `importacion.Fuentes()` en Go (dominio/importacion/fuentes.go),
 * pero a la medida de un pedido del Worker: en la demo cada paso lo pide el navegador y tiene que
 * entrar en un pedido (50 subpedidos a la red, 10 ms de CPU en el plan gratis).
 *
 * Tamaños de página:
 *   - Productos, de a `POR_PAGINA` (5), lo mismo que ya calculó "Importar tienda": cada producto se
 *     guarda con la misma función que el editor (varias consultas a D1 y armar el catálogo).
 *   - Clientes y pedidos (sólo Tiendanube), de a 50: se guardan con un `json_each` por página, así que
 *     cuestan lo mismo con 5 que con 50.
 *   - A la red sale una página por paso (WooCommerce, más sus variaciones; Tiendanube, más hasta dos
 *     reintentos ante un 429), muy por debajo de los 50.
 *
 * En la demo se importa hasta `MAX_PRODUCTOS` productos y `MAX_HISTORIAL` clientes y pedidos: es para
 * probar, no para mudar una tienda entera (eso lo hace la API, sin topes).
 */
import { MAX_PRODUCTOS, POR_PAGINA, origenDe, type Importado } from '../importar/comun.ts';
import { deShopify, type ProductoShopify } from '../importar/shopify.ts';
import { leerPaginaWoo } from '../importar/woocommerce.ts';
import { leerJson } from '../importar/comun.ts';
import {
	clienteTN,
	ErrorLimite,
	ErrorReconectar,
	leer,
	pedidoTN,
	primera,
	productoTN,
	type ApiTN,
	type ClienteTN,
	type PedidoTN,
	type ProductoTN
} from '../importar/tiendanube.ts';
import {
	ErrorDefinitivo,
	type Etapa,
	type Fuente,
	type Fuentes,
	type Origen,
	type Pagina,
	type ProductoViejo
} from './fuente.ts';

/** Clientes y pedidos por paso (Tiendanube). */
export const POR_PAGINA_HISTORIAL = 50;
/** Hasta cuántos clientes y cuántos pedidos se importan en la demo. */
export const MAX_HISTORIAL = 1000;
/** Páginas de productos que entran en el tope. */
const MAX_PAGINAS = Math.ceil(MAX_PRODUCTOS / POR_PAGINA);

export type Opciones = {
	fetch?: typeof fetch;
	/** Espera ante un 429 de Tiendanube; los tests la reemplazan. */
	dormir?: (ms: number) => Promise<void>;
	/** Un Tiendanube falso en localhost (sólo las pruebas de punta a punta). */
	tiendanubeLocal?: string | null;
};

const viejos = (items: Importado[]): ProductoViejo[] =>
	items.map((i) => ({ externo: i.externo ?? '', url: i.url ?? '', producto: i.producto }));

const paginaDe = (cursor: unknown) => {
	const n = (cursor as { pagina?: unknown } | null)?.pagina;
	return typeof n === 'number' && Number.isInteger(n) && n >= 1 ? n : 1;
};

/** El origen https de una tienda pública (sin IPs ni puertos: el Worker sólo sale a dominios). */
function origen(o: Origen) {
	const r = origenDe(o.url);
	if (!r) throw new ErrorDefinitivo('La dirección de la tienda no es válida.');
	return r;
}

/** Shopify: `/products.json` de a 5. Sin cuenta no hay clientes ni pedidos. */
function shopify(base: string, f: typeof fetch): Fuente {
	return {
		async pagina(etapa, cursor) {
			if (etapa !== 'productos') return { siguiente: null };
			const n = paginaDe(cursor);
			const r = await leerJson<{ products?: ProductoShopify[] }>(
				`${base}/products.json?limit=${POR_PAGINA}&page=${n}`,
				f
			);
			const crudos = r?.datos?.products;
			if (!Array.isArray(crudos)) {
				if (n === 1) throw new ErrorDefinitivo('La tienda no respondió como Shopify.');
				throw new Error('Shopify no respondió.');
			}
			const items = crudos.map(deShopify).filter((x): x is Importado => !!x);
			const sigue = crudos.length === POR_PAGINA && n < MAX_PAGINAS;
			return { productos: viejos(items), siguiente: sigue ? { pagina: n + 1 } : null };
		}
	};
}

/** WooCommerce: la Store API de a 5, con las variaciones de los variables de la página. */
function woo(base: string, f: typeof fetch): Fuente {
	return {
		async pagina(etapa, cursor) {
			if (etapa !== 'productos') return { siguiente: null };
			const n = paginaDe(cursor);
			const p = await leerPaginaWoo(base, n, f);
			if (!p) {
				if (n === 1) throw new ErrorDefinitivo('La tienda no respondió como WooCommerce.');
				// Pasada la última página, la Store API responde 400: no hay más.
				return { siguiente: null };
			}
			const sigue = p.leidos === POR_PAGINA && n < MAX_PAGINAS;
			return { productos: viejos(p.items), siguiente: sigue ? { pagina: n + 1 } : null };
		}
	};
}

/** Lo que se guarda (cifrado) al conectar la tienda de Tiendanube, como `CredencialTiendanube`. */
export type CredencialTN = { token: string; tienda: string };
export const credencialTiendanube = (token: string, tienda: string) =>
	JSON.stringify({ token, tienda } satisfies CredencialTN);

type CursorTN = { url: string; idioma: string; leidos: number };
const RECURSO: Partial<Record<Etapa, string>> = {
	productos: 'products',
	clientes: 'customers',
	pedidos: 'orders'
};

/** Tiendanube: productos, clientes y pedidos por la API. El cursor es la URL que manda el Link. */
function tiendanube(api: ApiTN): Fuente {
	return {
		async pagina(etapa, cursor): Promise<Pagina> {
			const recurso = RECURSO[etapa];
			if (!recurso) return { siguiente: null };
			const porPagina = etapa === 'productos' ? POR_PAGINA : POR_PAGINA_HISTORIAL;
			const tope = etapa === 'productos' ? MAX_PRODUCTOS : MAX_HISTORIAL;
			const previo = (cursor ?? {}) as Partial<CursorTN>;
			const c: CursorTN = {
				url:
					typeof previo.url === 'string' && previo.url
						? previo.url
						: primera(api, recurso, porPagina),
				idioma: typeof previo.idioma === 'string' ? previo.idioma : '',
				leidos: typeof previo.leidos === 'number' ? previo.leidos : 0
			};
			let pag;
			try {
				pag = await leer(api, c.url);
			} catch (e) {
				if (e instanceof ErrorReconectar) throw new ErrorDefinitivo(e.message);
				// Tiendanube pidió esperar más de lo que conviene en un pedido: el próximo paso reintenta
				// la misma página (sin outbox, el navegador hace de espera).
				if (e instanceof ErrorLimite) return { siguiente: c };
				throw e;
			}
			const idioma = pag.idioma || c.idioma;
			const p: Pagina = { siguiente: null };
			const regs = pag.registros.filter(
				(r): r is Record<string, unknown> => !!r && typeof r === 'object'
			);
			if (etapa === 'productos')
				p.productos = regs.flatMap((r) => {
					const i = productoTN(r as unknown as ProductoTN, idioma);
					return i ? [{ externo: i.externo, url: i.urlVieja, producto: i.producto }] : [];
				});
			if (etapa === 'clientes')
				p.clientes = regs.flatMap((r) => {
					const x = clienteTN(r as unknown as ClienteTN);
					return x ? [x] : [];
				});
			if (etapa === 'pedidos')
				p.pedidos = regs.flatMap((r) => {
					const x = pedidoTN(r as unknown as PedidoTN);
					return x
						? [
								{
									externo: x.externo,
									numero: x.numeroExterno,
									fecha: Date.parse(x.fecha),
									email: x.clienteEmail,
									nombre: x.clienteNombre,
									total: x.total,
									items: x.items,
									estado: x.estado
								}
							]
						: [];
				});
			const leidos = c.leidos + pag.registros.length;
			if (pag.siguiente && leidos < tope) p.siguiente = { url: pag.siguiente, idioma, leidos };
			return p;
		}
	};
}

/** Las fuentes de la demo, por plataforma. Mientras falte una, crear responde 'sin-fuente'. */
export function fuentes(o: Opciones = {}): Fuentes {
	const f = o.fetch ?? ((...a: Parameters<typeof fetch>) => fetch(...a));
	return {
		shopify: (x) => shopify(origen(x), f),
		woocommerce: (x) => woo(origen(x), f),
		tiendanube: (x) => {
			let c: Partial<CredencialTN> = {};
			try {
				c = JSON.parse(x.credencial ?? '{}') as Partial<CredencialTN>;
			} catch {
				// Queda vacía: falta conectar.
			}
			if (typeof c.token !== 'string' || !c.token || typeof c.tienda !== 'string' || !c.tienda)
				throw new ErrorDefinitivo('Falta conectar la tienda de Tiendanube.');
			// En un pedido del Worker no conviene esperar mucho: hasta 3 s por vez y dos reintentos.
			return tiendanube({
				token: c.token,
				tiendaId: c.tienda,
				fetch: f,
				dormir: o.dormir,
				local: o.tiendanubeLocal,
				esperaMax: 3000,
				reintentos: 2
			});
		}
	};
}

export const FUENTES: Fuentes = fuentes();
