/**
 * Importar desde Shopify. Toda tienda publica su catálogo sin clave en `/products.json` (de a
 * páginas) y la cantidad de productos en `/meta.json`. Gemelo en Go: `internal/importar/shopify.go`.
 *
 * Shopify no publica cuánto stock hay, sólo si la variante está disponible: disponible → 1, no
 * disponible → 0, y el resultado avisa que el stock es estimado.
 */
import {
	ajustar,
	leerJson,
	MAX_PRODUCTOS,
	pesos,
	POR_PAGINA,
	resumir,
	textoRico,
	type Crudo,
	type Importado,
	type Resultado
} from './comun.ts';

type Variante = {
	id: number;
	option1?: string | null;
	option2?: string | null;
	option3?: string | null;
	sku?: string | null;
	price?: string | null;
	available?: boolean;
};
export type ProductoShopify = {
	id: number;
	title?: string;
	handle?: string;
	body_html?: string | null;
	product_type?: string | null;
	options?: { name: string; position?: number; values?: string[] }[];
	variants?: Variante[];
	images?: { src?: string; variant_ids?: number[] }[];
};

/** "Title" con "Default Title": el producto no tiene variantes. */
const sinVariantes = (p: ProductoShopify) =>
	!p.options?.length ||
	(p.options.length === 1 &&
		p.options[0].name === 'Title' &&
		(p.options[0].values ?? []).every((v) => v === 'Default Title'));

/** Un producto de `/products.json` → producto de Mostrador (null si no tiene nombre). */
export function deShopify(p: ProductoShopify): Importado | null {
	const variantes = p.variants ?? [];
	const imagenes = p.images ?? [];
	const fotoDe = (id: number) => imagenes.find((i) => i.variant_ids?.includes(id))?.src ?? null;
	const stock = (v?: Variante) => (v && v.available !== false ? 1 : 0);
	const opciones = sinVariantes(p)
		? []
		: [...(p.options ?? [])]
				.sort((a, b) => (a.position ?? 0) - (b.position ?? 0))
				.map((o) => ({ nombre: o.name ?? '', valores: o.values ?? [] }));
	const primera = variantes[0];
	const crudo: Crudo = {
		nombre: p.title ?? '',
		descripcion: textoRico(p.body_html ?? ''),
		categoria: p.product_type || null,
		precio: pesos(primera?.price) ?? 0,
		sku: primera?.sku ?? '',
		stock: stock(primera),
		estimado: true,
		fotos: imagenes.map((i) => i.src ?? ''),
		opciones,
		variantes: opciones.length
			? variantes.map((v) => ({
					valores: [v.option1, v.option2, v.option3].slice(0, opciones.length).map((x) => x ?? ''),
					sku: v.sku ?? '',
					precio: pesos(v.price) ?? 0,
					stock: stock(v),
					foto: fotoDe(v.id)
				}))
			: []
	};
	return ajustar(crudo);
}

/** Una página de `/products.json` ya leída. */
export function paginaShopify(json: { products?: ProductoShopify[] }): {
	items: Importado[];
	leidos: number;
	omitidos: number;
} {
	const crudos = Array.isArray(json?.products) ? json.products : [];
	const items = crudos.map(deShopify).filter((x): x is Importado => !!x);
	return { items, leidos: crudos.length, omitidos: crudos.length - items.length };
}

/** Para los casos compartidos con Go. */
export function resultadoShopify(json: { products?: ProductoShopify[] }): Resultado {
	const { items, omitidos } = paginaShopify(json);
	return resumir(items, omitidos);
}

/**
 * ¿Es una tienda Shopify? `/meta.json` trae la cantidad publicada; si no está, se prueba la primera
 * página. Devuelve cuántos productos hay (null si no se sabe) o null si no es Shopify.
 */
export async function descubrirShopify(
	origen: string,
	f: typeof fetch = fetch
): Promise<{ total: number | null } | null> {
	const meta = await leerJson<{ published_products_count?: number; myshopify_domain?: string }>(
		`${origen}/meta.json`,
		f
	);
	const n = meta?.datos?.published_products_count;
	if (meta?.datos?.myshopify_domain && typeof n === 'number')
		return { total: Math.min(MAX_PRODUCTOS, Math.max(0, n)) };
	const prueba = await leerJson<{ products?: unknown }>(`${origen}/products.json?limit=1`, f);
	return Array.isArray(prueba?.datos?.products) ? { total: null } : null;
}

/** Lee la página `pagina` (desde 1) de a `POR_PAGINA`. Null si la tienda no respondió. */
export async function leerPaginaShopify(origen: string, pagina: number, f: typeof fetch = fetch) {
	const r = await leerJson<{ products?: ProductoShopify[] }>(
		`${origen}/products.json?limit=${POR_PAGINA}&page=${pagina}`,
		f
	);
	return r && Array.isArray(r.datos?.products) ? paginaShopify(r.datos) : null;
}
