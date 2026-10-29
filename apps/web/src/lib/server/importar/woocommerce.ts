/**
 * Importar desde WooCommerce con la Store API (`/wp-json/wc/store/v1/products`), que es pública. Los
 * precios vienen en unidades menores (`currency_minor_unit`). Las variaciones de los productos
 * variables (con su precio y stock) se piden aparte, de una vez para toda la página, con
 * `type=variation&parent=<ids>`. Gemelo en Go: `internal/importar/woocommerce.go`.
 *
 * Stock: `low_stock_remaining` si viene (sólo cuando la tienda lleva la cuenta y queda poco); si no,
 * en stock → 1 y se avisa que es estimado; sin stock → 0.
 */
import {
	ajustar,
	leerJson,
	MAX_PRODUCTOS,
	normal,
	pesos,
	POR_PAGINA,
	resumir,
	textoRico,
	type Crudo,
	type Importado,
	type Resultado
} from './comun.ts';

type Precios = { price?: string; regular_price?: string; currency_minor_unit?: number };
export type ProductoWoo = {
	id: number;
	name?: string;
	type?: string;
	sku?: string;
	description?: string;
	short_description?: string;
	prices?: Precios;
	images?: { src?: string }[];
	categories?: { name?: string }[];
	attributes?: {
		name?: string;
		has_variations?: boolean;
		terms?: { name?: string; slug?: string }[];
	}[];
	variations?: { id: number; attributes?: { name?: string; value?: string }[] }[];
	is_in_stock?: boolean;
	low_stock_remaining?: number | null;
	permalink?: string;
};

const API = '/wp-json/wc/store/v1/products';

const precio = (p?: Precios) => pesos(p?.price, p?.currency_minor_unit ?? 0);
/** El stock y si es estimado. */
const stockDe = (p?: ProductoWoo): [number, boolean] =>
	typeof p?.low_stock_remaining === 'number'
		? [Math.max(0, p.low_stock_remaining), false]
		: p?.is_in_stock
			? [1, true]
			: [0, false];

/**
 * Un producto de la Store API → producto de Tu Local Web. `variaciones` son las que se pidieron aparte
 * (las que falten quedan con el precio del producto y el stock que diga el producto).
 */
export function deWoo(p: ProductoWoo, variaciones: ProductoWoo[] = []): Importado | null {
	const atributos = (p.attributes ?? []).filter((a) => a.has_variations && a.terms?.length);
	const variable = p.type === 'variable' && atributos.length > 0;
	const [stock, estimado] = stockDe(p);
	let algunoEstimado = estimado;
	const fotos = (p.images ?? []).map((i) => i.src ?? '');
	const opciones = variable
		? atributos.map((a) => ({
				nombre: textoRico(a.name ?? ''),
				valores: (a.terms ?? []).map((t) => textoRico(t.name ?? ''))
			}))
		: [];

	const variantes: Crudo['variantes'] = [];
	if (variable) {
		const porId = new Map(variaciones.map((v) => [v.id, v]));
		for (const v of p.variations ?? []) {
			// Cada atributo: el valor (slug o nombre del término); vacío es "cualquiera".
			const posibles = atributos.map((a, i) => {
				const dado = v.attributes?.find(
					(x) => normal(x.name ?? '') === normal(a.name ?? '')
				)?.value;
				if (!dado) return opciones[i].valores;
				const t = (a.terms ?? []).find(
					(t) => normal(t.slug ?? '') === normal(dado) || normal(t.name ?? '') === normal(dado)
				);
				return [t ? textoRico(t.name ?? '') : dado];
			});
			const d = porId.get(v.id);
			const [s, e] = d ? stockDe(d) : [stock, estimado];
			algunoEstimado ||= e;
			const foto = d?.images?.[0]?.src ?? null;
			// La foto de la variación se suma a la galería si todavía no está.
			if (foto && !fotos.includes(foto)) fotos.push(foto);
			let combos: string[][] = [[]];
			for (const vals of posibles) combos = combos.flatMap((c) => vals.map((x) => [...c, x]));
			for (const valores of combos)
				variantes.push({
					valores,
					sku: d?.sku ?? '',
					precio: (d && precio(d.prices)) ?? precio(p.prices) ?? 0,
					stock: s,
					foto
				});
		}
	}

	const crudo: Crudo = {
		nombre: textoRico(p.name ?? ''),
		descripcion: textoRico(p.description || p.short_description || ''),
		categoria: p.categories?.[0]?.name ? textoRico(p.categories[0].name) : null,
		precio: precio(p.prices) ?? 0,
		sku: p.sku ?? '',
		stock,
		estimado: algunoEstimado,
		fotos,
		opciones,
		variantes
	};
	const i = ajustar(crudo);
	let url = '';
	try {
		const path = new URL(p.permalink ?? '').pathname;
		if (path !== '/') url = path;
	} catch {
		// Sin permalink no hay redirección.
	}
	return i && { ...i, externo: String(p.id), url };
}

/** Una página de productos ya leída, con las variaciones de sus productos variables. */
export function paginaWoo(
	productos: ProductoWoo[],
	variaciones: ProductoWoo[]
): { items: Importado[]; leidos: number; omitidos: number } {
	const lista = Array.isArray(productos) ? productos : [];
	const items = lista.map((p) => deWoo(p, variaciones)).filter((x): x is Importado => !!x);
	return { items, leidos: lista.length, omitidos: lista.length - items.length };
}

/** Para los casos compartidos con Go. */
export function resultadoWoo(json: {
	productos: ProductoWoo[];
	variaciones: ProductoWoo[];
}): Resultado {
	const { items, omitidos } = paginaWoo(json.productos, json.variaciones);
	return resumir(items, omitidos);
}

/** ¿Es una tienda WooCommerce? La cantidad de productos sale de `X-WP-Total`. */
export async function descubrirWoo(
	origen: string,
	f: typeof fetch = fetch
): Promise<{ total: number | null } | null> {
	const r = await leerJson<unknown>(`${origen}${API}?per_page=1&page=1`, f);
	if (!r || !Array.isArray(r.datos)) return null;
	const n = Number(r.cabeceras.get('x-wp-total'));
	return { total: Number.isFinite(n) && n >= 0 ? Math.min(MAX_PRODUCTOS, n) : null };
}

/** Hasta cuántas páginas de variaciones se piden por página de productos. */
const MAX_PAGINAS_VARIACIONES = 5;

/** Lee la página `pagina` (desde 1) de a `POR_PAGINA`, con sus variaciones. */
export async function leerPaginaWoo(origen: string, pagina: number, f: typeof fetch = fetch) {
	const r = await leerJson<ProductoWoo[]>(
		`${origen}${API}?per_page=${POR_PAGINA}&page=${pagina}`,
		f
	);
	if (!r || !Array.isArray(r.datos)) return null;
	const padres = r.datos.filter((p) => p.type === 'variable').map((p) => p.id);
	const variaciones: ProductoWoo[] = [];
	for (let i = 1; padres.length && i <= MAX_PAGINAS_VARIACIONES; i++) {
		const v = await leerJson<ProductoWoo[]>(
			`${origen}${API}?type=variation&parent=${padres.join(',')}&per_page=100&page=${i}`,
			f
		);
		if (!v || !Array.isArray(v.datos)) break;
		variaciones.push(...v.datos);
		if (i >= Number(v.cabeceras.get('x-wp-totalpages') || 1)) break;
	}
	return paginaWoo(r.datos, variaciones);
}
