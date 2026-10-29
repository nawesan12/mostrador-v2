/**
 * Importar desde Empretienda. No tiene API: leemos la tienda pública como la ve un comprador. El
 * sitemap lista las fichas, y cada ficha trae su producto en variables de JavaScript (`stock`,
 * `imagenes`, `p_nombre`, `categorias_flatten`, `cdn_products`), además de la descripción en el HTML.
 */
import type { ProductoEditable } from './datos/tipos.ts';
import {
	ajustar,
	leer,
	normal,
	origenDe as origenComun,
	resumir,
	textoRico,
	type Crudo,
	type Importado,
	type Resultado
} from './importar/comun.ts';

/** Stock "ilimitado" de Empretienda: en Tu Local Web queda con este número y se ajusta a mano. */
export const STOCK_ILIMITADO = 999;

/** Acepta "mitienda", "mitienda.empretienda.com.ar" o el link completo; devuelve el origen https. */
export const origenDe = (entrada: string) => origenComun(entrada, '.empretienda.com.ar');

/** Las fichas del sitemap: en Empretienda son las URL con prioridad 0.8. */
export function fichasDelSitemap(xml: string, origen: string): string[] {
	const fichas = new Set<string>();
	for (const [, bloque] of xml.matchAll(/<url>([\s\S]*?)<\/url>/g)) {
		const loc = bloque.match(/<loc>([^<]+)<\/loc>/)?.[1]?.trim();
		const prioridad = bloque.match(/<priority>([^<]+)<\/priority>/)?.[1];
		if (!loc || prioridad !== '0.8') continue;
		try {
			if (`https://${new URL(loc).hostname}` === origen) fichas.add(loc);
		} catch {
			/* una URL rota no frena el resto */
		}
	}
	return [...fichas];
}

/** Lee `var nombre = <json>;` del HTML. */
function variable<T>(html: string, nombre: string): T | undefined {
	const i = html.indexOf(`var ${nombre} = `);
	if (i < 0) return undefined;
	const desde = i + nombre.length + 7;
	const fin = html.indexOf(';\n', desde);
	const crudo = html
		.slice(desde, fin < 0 ? undefined : fin)
		.trim()
		.replace(/;$/, '');
	try {
		return JSON.parse(crudo) as T;
	} catch {
		return undefined;
	}
}

/** Una combinación de la ficha: `valoratributo` dice el atributo (`at_nombre`) y su valor. */
type Stock = {
	s_cantidad: number;
	s_ilimitado: number;
	s_precio: number;
	s_oferta: number;
	s_precio_oferta: number;
	s_imagen?: number;
	valoratributo?: { at_nombre?: string; valor?: { vat_valor?: string } }[];
};

/** Lo que se usa de una ficha: sus variables y la descripción (la forma del caso compartido). */
export type FichaEmpretienda = {
	url: string;
	p_nombre: string;
	cdn_products?: string;
	imagenes?: { idImagenes: number; i_link: string }[];
	categorias_flatten?: { c_link_full: string; c_nombre: string }[];
	stock: Stock[];
	descripcion: string;
};

/** La ficha de la página. Null si la página no es una ficha. */
export function fichaDeHtml(html: string, url: string): FichaEmpretienda | null {
	const p_nombre = variable<string>(html, 'p_nombre');
	const stock = variable<Stock[]>(html, 'stock');
	if (typeof p_nombre !== 'string' || !stock?.length) return null;
	return {
		url,
		p_nombre,
		cdn_products: variable<string>(html, 'cdn_products') ?? '',
		imagenes: variable<FichaEmpretienda['imagenes']>(html, 'imagenes') ?? [],
		categorias_flatten:
			variable<FichaEmpretienda['categorias_flatten']>(html, 'categorias_flatten') ?? [],
		stock,
		descripcion:
			html.match(/class="product-vip__description[^"]*"[^>]*>([\s\S]*?)<\/div>/)?.[1]?.trim() ?? ''
	};
}

/**
 * Una ficha → producto de Tu Local Web, con sus variantes reales: cada stock es una combinación, con su
 * precio (el de oferta si la tiene), su stock (ilimitado → STOCK_ILIMITADO) y su foto (`s_imagen`).
 * Gemelo en Go: `importar.DeEmpretienda`.
 */
export function deEmpretienda(f: FichaEmpretienda): Importado | null {
	let camino = '';
	try {
		camino = new URL(f.url).pathname;
	} catch {
		/* sin path no hay categoría */
	}
	// La categoría es la más larga cuyo link es prefijo del path de la ficha.
	const categoria =
		(f.categorias_flatten ?? [])
			.filter((c) => camino.startsWith(c.c_link_full + '/'))
			.sort((a, b) => b.c_link_full.length - a.c_link_full.length)[0]
			?.c_nombre.trim() ?? '';
	const cdn = f.cdn_products ?? '';
	const fotoDe = new Map<number, string>();
	const fotos: string[] = [];
	for (const im of f.imagenes ?? []) {
		if (!im.i_link) continue;
		fotoDe.set(im.idImagenes, cdn + im.i_link);
		fotos.push(cdn + im.i_link);
	}
	const variantes = f.stock.map((s) => ({
		valores: (s.valoratributo ?? []).map((a) => a.valor?.vat_valor ?? ''),
		sku: '',
		precio: s.s_oferta && s.s_precio_oferta > 0 ? s.s_precio_oferta : s.s_precio,
		stock: s.s_ilimitado ? STOCK_ILIMITADO : Math.max(0, s.s_cantidad || 0),
		foto: fotoDe.get(s.s_imagen ?? 0) ?? null
	}));
	const c: Crudo = {
		nombre: f.p_nombre,
		descripcion: textoRico(f.descripcion),
		categoria,
		precio: Math.min(...variantes.map((v) => v.precio)),
		sku: '',
		stock: variantes[0].stock,
		estimado: false,
		fotos,
		opciones: [],
		variantes: []
	};
	// Las opciones salen de la primera combinación (todas traen los mismos atributos, en orden).
	const attrs = f.stock[0].valoratributo ?? [];
	if (attrs.length) {
		c.opciones = attrs.map((a, i) => {
			const valores: string[] = [];
			const vistos = new Set<string>();
			for (const v of variantes) {
				const x = (v.valores[i] ?? '').trim();
				if (x && !vistos.has(normal(x))) {
					vistos.add(normal(x));
					valores.push(x);
				}
			}
			return { nombre: a.at_nombre?.trim() || `Opción ${i + 1}`, valores };
		});
		c.variantes = variantes;
	}
	const i = ajustar(c);
	// Como en Go, el path de la ficha es su id y la dirección vieja (para las redirecciones).
	return i && { ...i, externo: camino, url: camino };
}

/** Las fichas de una página (lo que comparan los casos con Go). */
export function resultadoEmpretienda(fichas: FichaEmpretienda[]): Resultado {
	const items = fichas.map(deEmpretienda);
	return resumir(
		items.filter((i): i is Importado => !!i),
		items.filter((i) => !i).length
	);
}

/** Una ficha de Empretienda → producto de Tu Local Web. Null si la página no es una ficha. */
export function productoDeFicha(html: string, url: string): ProductoEditable | null {
	const f = fichaDeHtml(html, url);
	return (f && deEmpretienda(f)?.producto) ?? null;
}

export { leer };
