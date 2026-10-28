/**
 * Lo común a los importadores de catálogo (Empretienda, Shopify, WooCommerce): leer la tienda con
 * topes, pasar HTML a texto, redondear precios y ajustar cada producto a los topes de Mostrador.
 *
 * `ajustar`, `textoRico` y `pesos` tienen su gemelo en Go (`apps/api/internal/importar/comun.go`):
 * los dos pasan los casos de `packages/contrato/casos/importar/`. Si cambia una regla, cambia en los
 * dos lados y en los `.esperado.json`.
 */
import type { Datos, ProductoEditable } from '../datos/tipos.ts';

/** Productos que se leen y guardan por pedido: cada uno son 2 subpedidos (repetido y guardar). */
export const POR_PAGINA = 5;
/** Hasta cuántos productos se importan de una tienda. */
export const MAX_PRODUCTOS = 500;

/** Topes de ProductoEditable (los mismos que validan el sandbox y Go). */
const TOPE = {
	nombre: 120,
	descripcion: 4000,
	categoria: 60,
	sku: 40,
	fotos: 8,
	largoFoto: 600,
	opciones: 3,
	nombreOpcion: 20,
	valores: 20,
	valor: 20,
	variantes: 100,
	stock: 1_000_000
};

/**
 * Acepta "mitienda", "mitienda.com.ar" o el link completo; devuelve el origen https. Sin IPs, sin
 * localhost y sin puerto: el servidor sólo sale a buscar dominios públicos. Con `sufijo`, un nombre
 * suelto se completa ("mitienda" → mitienda.empretienda.com.ar).
 */
export function origenDe(entrada: string, sufijo?: string): string | null {
	let t = entrada.trim().toLowerCase();
	if (!t) return null;
	if (sufijo && /^[a-z0-9-]{2,63}$/.test(t)) t += sufijo;
	try {
		const u = new URL(t.includes('://') ? t : `https://${t}`);
		if (!/^([a-z0-9-]+\.)+[a-z]{2,}$/.test(u.hostname) || u.port) return null;
		return `https://${u.hostname}`;
	} catch {
		return null;
	}
}

/** HTML a texto: saltos en `<br>` y `</p>`, sin etiquetas y con las entidades de siempre. */
export const texto = (html: string) =>
	html
		.replace(/<br\s*\/?>/gi, '\n')
		.replace(/<\/p>/gi, '\n')
		.replace(/<[^>]+>/g, '')
		.replace(/&nbsp;/g, ' ')
		.replace(/&amp;/g, '&')
		.replace(/&lt;/g, '<')
		.replace(/&gt;/g, '>')
		.replace(/&quot;/g, '"')
		.replace(/&#0?39;/g, "'")
		.replace(/\n{3,}/g, '\n\n')
		.trim();

/** Entidades con nombre que aparecen en los catálogos (el resto queda como vino). */
const ENTIDADES: Record<string, string> = {
	ndash: '–',
	mdash: '—',
	lsquo: '‘',
	rsquo: '’',
	ldquo: '“',
	rdquo: '”',
	laquo: '«',
	raquo: '»',
	hellip: '…',
	deg: '°',
	ordm: 'º',
	ordf: 'ª',
	iexcl: '¡',
	iquest: '¿',
	aacute: 'á',
	eacute: 'é',
	iacute: 'í',
	oacute: 'ó',
	uacute: 'ú',
	Aacute: 'Á',
	Eacute: 'É',
	Iacute: 'Í',
	Oacute: 'Ó',
	Uacute: 'Ú',
	ntilde: 'ñ',
	Ntilde: 'Ñ',
	uuml: 'ü',
	Uuml: 'Ü',
	times: '×',
	trade: '™',
	reg: '®',
	copy: '©'
};

/** Un código numérico de entidad a su carácter (null si no es válido). */
function caracter(n: number): string | null {
	if (!Number.isFinite(n) || n <= 0 || n > 0x10ffff || (n >= 0xd800 && n <= 0xdfff)) return null;
	return n === 160 ? ' ' : String.fromCodePoint(n);
}

/**
 * HTML de un catálogo (Shopify, WooCommerce) a texto. Más completo que `texto`: los espacios del
 * código fuente no cuentan, los bloques (`li`, `div`, títulos) cortan renglón y se entienden las
 * entidades numéricas (`&#8217;`). Cada renglón sin espacios en las puntas.
 */
export function textoRico(html: string): string {
	const t = texto(
		html.replace(/[ \t\n\r\f]+/g, ' ').replace(/<\/(li|div|h[1-6]|tr|ul|ol)>/gi, '\n')
	)
		.replace(/&#(\d{1,7});/g, (e, n: string) => caracter(Number(n)) ?? e)
		.replace(/&#[xX]([0-9a-fA-F]{1,6});/g, (e, n: string) => caracter(parseInt(n, 16)) ?? e)
		.replace(/&([a-zA-Z]+);/g, (e, n: string) => ENTIDADES[n] ?? e);
	return t
		.split('\n')
		.map((r) => r.replace(/^[ \t]+|[ \t]+$/g, ''))
		.join('\n')
		.replace(/\n{3,}/g, '\n\n')
		.replace(/^\n+|\n+$/g, '');
}

/**
 * Un precio en texto ("50.00", o "1250050" con 2 decimales en WooCommerce) a pesos enteros,
 * redondeando la mitad para arriba. Null si no es un número.
 */
export function pesos(valor: unknown, decimales = 0): number | null {
	const m = String(valor ?? '')
		.trim()
		.match(/^(\d+)(?:\.(\d+))?$/);
	if (!m) return null;
	const frac = m[2] ?? '';
	const digitos = (m[1] + frac).replace(/^0+(?=\d)/, '');
	const d = decimales + frac.length;
	if (digitos.length > 15 || d > 15) return null;
	const n = Number(digitos);
	const div = 10 ** d;
	const q = Math.floor(n / div);
	return (n - q * div) * 2 >= div ? q + 1 : q;
}

/** Sin mayúsculas ni tildes: así compara Mostrador nombres de opciones y valores repetidos. */
export const normal = (s: string) =>
	s
		.normalize('NFD')
		.replace(/\p{Mn}/gu, '')
		.normalize('NFC')
		.toLowerCase();

const colapsar = (s: string) => s.replace(/[ \t\n\r\f]+/g, ' ').replace(/^ | $/g, '');
const largo = (s: string) => [...s].length;
const bytes = new TextEncoder();
const cortar = (s: string, n: number) => [...s].slice(0, n).join('').replace(/ +$/, '');

/** Un producto como lo leyó el importador, antes de los topes. Las variantes van por posición. */
export type Crudo = {
	nombre: string;
	descripcion: string;
	categoria: string | null;
	/** En pesos enteros. Con variantes, el producto toma el más bajo. */
	precio: number;
	sku: string;
	stock: number;
	/** La plataforma no publica la cantidad: el stock es 1 (hay) o 0 (no hay). */
	estimado: boolean;
	fotos: string[];
	opciones: { nombre: string; valores: string[] }[];
	/** `valores[i]` es el valor de `opciones[i]`. */
	variantes: {
		valores: string[];
		sku: string;
		precio: number;
		stock: number;
		foto: string | null;
	}[];
};

/** Un producto listo para guardar, con lo que hay que avisar. */
export type Importado = { producto: ProductoEditable; recortado: boolean; estimado: boolean };

/** Lo que devuelven los parsers por página (y lo que comparan los casos compartidos con Go). */
export type Resultado = {
	productos: ProductoEditable[];
	/** Algún producto tiene el stock estimado: la interfaz pide revisarlo. */
	stockEstimado: boolean;
	/** Cuántos productos se recortaron para entrar en los topes. */
	recortes: number;
	/** Cuántos se descartaron (sin nombre). */
	omitidos: number;
};

export const resumir = (items: Importado[], omitidos: number): Resultado => ({
	productos: items.map((i) => i.producto),
	stockEstimado: items.some((i) => i.estimado),
	recortes: items.filter((i) => i.recortado).length,
	omitidos
});

/**
 * Ajusta un producto a los topes de Mostrador. Lo que no entra se recorta y se marca: nombre y
 * textos se cortan; fotos (sólo https), opciones y valores se limitan; si las combinaciones pasan
 * de 100, se sacan valores de la opción más larga. Las variantes que quedan sin valor se descartan.
 * Null si el nombre no llega a 2 letras.
 */
export function ajustar(c: Crudo): Importado | null {
	let rec = false;
	const corta = (s: string, n: number) => {
		if (largo(s) <= n) return s;
		rec = true;
		return cortar(s, n);
	};
	const nombreLimpio = colapsar(c.nombre);
	if (largo(nombreLimpio) < 2) return null;
	const nombre = corta(nombreLimpio, TOPE.nombre);
	const descripcion = corta(c.descripcion, TOPE.descripcion);
	const cat = colapsar(c.categoria ?? '');
	const categoria = cat ? corta(cat, TOPE.categoria) : null;
	const sku = corta(c.sku.trim(), TOPE.sku);

	const fotos: string[] = [];
	for (const f of c.fotos.map((x) => x.trim())) {
		// El largo en bytes, como lo valida Go.
		if (!f.startsWith('https://') || bytes.encode(f).length > TOPE.largoFoto) rec = true;
		else if (fotos.includes(f)) continue;
		else if (fotos.length === TOPE.fotos) rec = true;
		else fotos.push(f);
	}

	// Cada opción de la plataforma tiene su mapa valor (sin mayúsculas ni tildes) → valor final; null:
	// la opción se descartó.
	const opciones: { nombre: string; valores: string[] }[] = [];
	const mapas: (Map<string, string> | null)[] = [];
	for (const o of c.opciones) {
		const n = corta(colapsar(o.nombre), TOPE.nombreOpcion);
		if (
			!n ||
			opciones.length === TOPE.opciones ||
			opciones.some((x) => normal(x.nombre) === normal(n))
		) {
			rec = true;
			mapas.push(null);
			continue;
		}
		const valores: string[] = [];
		const mapa = new Map<string, string>();
		for (const crudo of o.valores) {
			const v = corta(colapsar(crudo), TOPE.valor);
			if (!v) continue;
			const igual = valores.find((x) => normal(x) === normal(v));
			if (igual) mapa.set(normal(v), igual);
			else if (valores.length === TOPE.valores) rec = true;
			else {
				valores.push(v);
				mapa.set(normal(v), v);
			}
		}
		if (!valores.length) {
			rec = true;
			mapas.push(null);
			continue;
		}
		opciones.push({ nombre: n, valores });
		mapas.push(mapa);
	}
	// Hasta 100 combinaciones: se saca el último valor de la opción con más (la última si empatan).
	const combinaciones = () => opciones.reduce((t, o) => t * o.valores.length, 1);
	while (opciones.length && combinaciones() > TOPE.variantes) {
		let j = 0;
		opciones.forEach((o, i) => o.valores.length >= opciones[j].valores.length && (j = i));
		const fuera = opciones[j].valores.pop();
		const mapa = mapas.filter((m) => m)[j]!;
		for (const [k, v] of mapa) if (v === fuera) mapa.delete(k);
		rec = true;
	}

	const variantes: {
		valores: Record<string, string>;
		sku: string;
		precio: number;
		stock: number;
		foto: string | null;
	}[] = [];
	const vistas = new Set<string>();
	if (opciones.length)
		for (const v of c.variantes) {
			const valores: Record<string, string> = {};
			let ok = true;
			let k = 0;
			mapas.forEach((mapa, i) => {
				if (!mapa) return;
				const final = mapa.get(normal(cortar(colapsar(v.valores[i] ?? ''), TOPE.valor)));
				if (final === undefined) ok = false;
				else valores[opciones[k].nombre] = final;
				k++;
			});
			const clave = opciones.map((o) => valores[o.nombre]).join('\u0000');
			if (!ok || vistas.has(clave)) {
				rec = true;
				continue;
			}
			vistas.add(clave);
			variantes.push({
				valores,
				sku: corta(v.sku.trim(), TOPE.sku),
				precio: v.precio,
				stock: Math.min(TOPE.stock, Math.max(0, Math.trunc(v.stock))),
				foto: v.foto && fotos.includes(v.foto) ? v.foto : null
			});
		}

	const precio = variantes.length ? Math.min(...variantes.map((v) => v.precio)) : c.precio;
	const producto: ProductoEditable = {
		nombre,
		descripcion,
		categoria,
		precio,
		sku: variantes.length ? '' : sku,
		stock: variantes.length
			? variantes.reduce((t, v) => t + v.stock, 0)
			: Math.min(TOPE.stock, Math.max(0, Math.trunc(c.stock))),
		fotos,
		opciones,
		variantes: variantes.map((v) => ({ ...v, precio: v.precio === precio ? null : v.precio })),
		canales: { web: true }
	};
	return { producto, recortado: rec, estimado: c.estimado };
}

/** Pide una página de la tienda, con tope de tamaño y de tiempo. Null si no responde bien. */
export async function pedir(
	url: string,
	f: typeof fetch = fetch
): Promise<{ texto: string; cabeceras: Headers } | null> {
	const r = await f(url, {
		// En ASCII: con tildes, workerd lo manda como UTF-8 y hay tiendas que no responden.
		headers: { 'user-agent': 'Mostrador (importacion de la tienda del dueno)' },
		redirect: 'follow',
		signal: AbortSignal.timeout(8000)
	}).catch(() => null);
	if (!r?.ok) return null;
	const t = await r.text().catch(() => null);
	return t === null || t.length > 2_000_000 ? null : { texto: t, cabeceras: r.headers };
}

/** Como `pedir`, sólo el cuerpo. */
export const leer = async (url: string, f: typeof fetch = fetch) =>
	(await pedir(url, f))?.texto ?? null;

/** Lee un JSON de la tienda (null si no responde o no es JSON). */
export async function leerJson<T>(
	url: string,
	f: typeof fetch = fetch
): Promise<{ datos: T; cabeceras: Headers } | null> {
	const r = await pedir(url, f);
	if (!r) return null;
	try {
		return { datos: JSON.parse(r.texto) as T, cabeceras: r.cabeceras };
	} catch {
		return null;
	}
}

/** La regla de no repetir: un producto con el mismo nombre (sin mayúsculas) ya está en la tienda. */
export async function yaEsta(datos: Datos, nombre: string): Promise<boolean> {
	const ya = await datos.listar('productos', { q: nombre, cuantos: 5 });
	return ya.filas.some((f) => String(f.nombre ?? '').toLowerCase() === nombre.toLowerCase());
}
