import { error, json } from '@sveltejs/kit';
import * as v from 'valibot';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

const Editable = v.object({
	nombre: v.pipe(v.string(), v.trim(), v.minLength(2, 'Poné un nombre.'), v.maxLength(120)),
	descripcion: v.optional(v.pipe(v.string(), v.maxLength(4000))),
	categoria: v.optional(v.nullable(v.pipe(v.string(), v.trim(), v.maxLength(60)))),
	precio: v.pipe(v.number(), v.integer(), v.minValue(0), v.maxValue(1_000_000_000)),
	costo: v.optional(v.pipe(v.number(), v.integer(), v.minValue(0), v.maxValue(1_000_000_000))),
	sku: v.optional(v.pipe(v.string(), v.trim(), v.maxLength(40))),
	stock: v.optional(v.pipe(v.number(), v.integer(), v.minValue(0), v.maxValue(1_000_000))),
	// Una foto de Unsplash (id) o una URL https (los productos importados de Empretienda).
	foto: v.optional(
		v.nullable(
			v.union([
				v.pipe(v.string(), v.regex(/^\d{10,}-[0-9a-f]+$/)),
				v.pipe(v.string(), v.maxLength(500), v.url(), v.startsWith('https://'))
			])
		)
	),
	canales: v.optional(v.record(v.picklist(['web', 'ml', 'ig', 'google']), v.boolean()))
});

/** La ficha del producto (el editor). */
export const GET: RequestHandler = async ({ locals, params }) => {
	const p = await datosDe(locals).producto(params.id);
	if (!p) error(404, 'Ese producto no existe.');
	return json(p, { headers: { 'cache-control': 'no-store' } });
};

/** Guardar (o crear, con id "nuevo"). */
export const PUT: RequestHandler = async ({ locals, params, request }) => {
	const r = v.safeParse(Editable, await request.json().catch(() => null));
	if (!r.success) error(400, r.issues[0].message);
	const datos = datosDe(locals);
	const antes = params.id === 'nuevo' ? null : await datos.producto(params.id);
	const p = await datos.guardarProducto(params.id === 'nuevo' ? null : params.id, r.output);
	if (!p) error(404, 'Ese producto no existe.');
	// Mercado Libre: prenderlo publica el producto y apagarlo pausa la publicación.
	const ml = !!r.output.canales?.ml;
	let mercadoLibre: { publicado?: string; error?: string } | undefined;
	if (ml !== !!antes?.canales?.ml && r.output.canales) {
		if (ml) {
			const pub = await datos.publicarEnMercadoLibre(p.id);
			mercadoLibre = 'error' in pub ? { error: pub.error } : { publicado: pub.permalink };
			if ('error' in pub)
				await datos.guardarProducto(p.id, {
					...r.output,
					canales: { ...r.output.canales, ml: false }
				});
		} else await datos.despublicarDeMercadoLibre(p.id);
	}
	return json({ ...p, ...(mercadoLibre ? { mercadoLibre } : {}) });
};
