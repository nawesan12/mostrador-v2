/**
 * Importar desde Empretienda. No tiene API: leemos la tienda pública como la ve un comprador. El
 * sitemap lista las fichas, y cada ficha trae su producto en variables de JavaScript (`stock`,
 * `imagenes`, `p_nombre`, `categorias_flatten`, `cdn_products`), además de la descripción en el HTML.
 */
import type { ProductoEditable } from './datos/tipos.ts';
import { leer, origenDe as origenComun, texto } from './importar/comun.ts';

/** Stock "ilimitado" de Empretienda: en Mostrador queda con este número y se ajusta a mano. */
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

type Stock = {
	s_cantidad: number;
	s_ilimitado: number;
	s_precio: number;
	s_oferta: number;
	s_precio_oferta: number;
	s_mostrar_precio: number;
	valoratributo?: { valor?: { vat_valor?: string } }[];
};
type Categoria = { c_link_full: string; c_nombre: string };

/** Una ficha de Empretienda → producto de Mostrador. Null si la página no es una ficha. */
export function productoDeFicha(html: string, url: string): ProductoEditable | null {
	const nombre = variable<string>(html, 'p_nombre')?.replace(/\s+/g, ' ').trim();
	const stock = variable<Stock[]>(html, 'stock');
	if (!nombre || nombre.length < 2 || !stock?.length) return null;
	// Con variantes (talle, color), el precio es el más bajo y el stock la suma.
	const precios = stock.map((s) =>
		s.s_oferta && s.s_precio_oferta > 0 ? s.s_precio_oferta : s.s_precio
	);
	const ilimitado = stock.some((s) => s.s_ilimitado);
	const cantidad = stock.reduce((t, s) => t + Math.max(0, s.s_cantidad || 0), 0);

	const cdn = variable<string>(html, 'cdn_products') ?? '';
	const imagenes = variable<{ i_link: string }[]>(html, 'imagenes') ?? [];
	const foto = imagenes[0]?.i_link ? cdn + imagenes[0].i_link : null;

	// La categoría es la más larga cuyo link es prefijo del path de la ficha.
	const path = new URL(url).pathname;
	const categorias = variable<Categoria[]>(html, 'categorias_flatten') ?? [];
	const categoria =
		categorias
			.filter((c) => path.startsWith(c.c_link_full + '/'))
			.sort((a, b) => b.c_link_full.length - a.c_link_full.length)[0]?.c_nombre ?? null;

	// Mostrador todavía no tiene variantes: quedan anotadas en la descripción para no perderlas.
	const variantes = stock
		.map((s) => {
			const v = (s.valoratributo ?? []).map((a) => a.valor?.vat_valor?.trim()).filter(Boolean);
			return v.length ? `${v.join(' / ')} (${s.s_ilimitado ? 'sin límite' : s.s_cantidad})` : '';
		})
		.filter(Boolean);
	const desc = html.match(/class="product-vip__description[^"]*"[^>]*>([\s\S]*?)<\/div>/)?.[1];

	return {
		nombre: nombre.slice(0, 120),
		descripcion:
			[desc ? texto(desc) : '', variantes.length ? `Variantes: ${variantes.join(', ')}.` : '']
				.filter(Boolean)
				.join('\n\n')
				.slice(0, 4000) || undefined,
		categoria: categoria?.slice(0, 60) ?? null,
		precio: Math.round(Math.min(...precios.filter((p) => p >= 0), Infinity) || 0),
		stock: ilimitado ? STOCK_ILIMITADO : cantidad,
		foto,
		canales: { web: true }
	};
}

export { leer };
