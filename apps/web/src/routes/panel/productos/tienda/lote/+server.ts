import { error, json } from '@sveltejs/kit';
import * as v from 'valibot';
import { datosDe } from '#lib/server/datos/index.ts';
import { productoDeFicha } from '#lib/server/empretienda.ts';
import {
	leer,
	MAX_PRODUCTOS,
	POR_PAGINA,
	yaEsta,
	type Importado
} from '#lib/server/importar/comun.ts';
import { FUENTES, origenPara } from '#lib/server/importar/fuentes.ts';
import { leerPaginaShopify } from '#lib/server/importar/shopify.ts';
import { leerPaginaWoo } from '#lib/server/importar/woocommerce.ts';
import type { RequestHandler } from './$types';

/** Empretienda: cada ficha leída, como importado (sin topes nuevos: queda como era). */
async function fichas(origen: string, urls: string[]) {
	// Sólo páginas de esa tienda: el servidor no sale a buscar cualquier URL.
	const propias = urls.filter((u) => u.startsWith(origen + '/'));
	const leidas = await Promise.all(
		propias.map(async (url) => {
			const html = await leer(url);
			return html ? productoDeFicha(html, url) : null;
		})
	);
	const items = leidas
		.filter((p) => !!p)
		.map((producto): Importado => ({ producto, recortado: false, estimado: false }));
	return { items, leidos: urls.length, omitidos: urls.length - items.length };
}

/**
 * Importar tienda, paso 2: una página de `POR_PAGINA` productos por pedido. Cada producto son 2
 * subpedidos (¿ya está? y guardar) más la página y, en WooCommerce, sus variaciones: muy por
 * debajo de los 50 del Worker. Los que ya están, por nombre, no se repiten: importar dos veces no
 * duplica. Se guardan de a uno para que dos con el mismo nombre en la página tampoco se dupliquen.
 */
export const POST: RequestHandler = async ({ locals, request }) => {
	const b = v.safeParse(
		v.object({
			fuente: v.picklist(FUENTES),
			origen: v.pipe(v.string(), v.maxLength(200)),
			pagina: v.optional(
				v.pipe(v.number(), v.integer(), v.minValue(1), v.maxValue(MAX_PRODUCTOS / POR_PAGINA))
			),
			fichas: v.optional(
				v.pipe(
					v.array(v.pipe(v.string(), v.maxLength(500))),
					v.minLength(1),
					v.maxLength(POR_PAGINA)
				)
			)
		}),
		await request.json().catch(() => null)
	);
	const origen = b.success ? origenPara(b.output.fuente, b.output.origen) : null;
	if (!b.success || !origen) error(400, 'Pedido inválido.');
	const { fuente, pagina = 1 } = b.output;
	if (fuente === 'empretienda' && !b.output.fichas) error(400, 'Pedido inválido.');

	const leida =
		fuente === 'empretienda'
			? await fichas(origen, b.output.fichas ?? [])
			: fuente === 'shopify'
				? await leerPaginaShopify(origen, pagina)
				: await leerPaginaWoo(origen, pagina);
	if (!leida) error(502, 'La tienda no respondió. Probá de nuevo en un rato.');

	const datos = datosDe(locals);
	let creados = 0,
		repetidos = 0,
		recortados = 0,
		conVariantes = 0,
		stockEstimado = false;
	let fallidos = leida.omitidos;
	for (const { producto: p, recortado, estimado } of leida.items) {
		if (await yaEsta(datos, p.nombre)) {
			repetidos++;
			continue;
		}
		if (!(await datos.guardarProducto(null, p))) {
			fallidos++;
			continue;
		}
		creados++;
		if (recortado) recortados++;
		if (estimado) stockEstimado = true;
		if (p.descripcion?.includes('Variantes:')) conVariantes++;
	}
	return json({
		leidos: leida.leidos,
		creados,
		repetidos,
		recortados,
		fallidos,
		conVariantes: fuente === 'empretienda' ? conVariantes : 0,
		stockEstimado,
		// Una página incompleta es la última (Shopify no siempre dice cuántos productos tiene).
		fin: fuente === 'empretienda' || leida.leidos < POR_PAGINA
	});
};
