import { error } from '@sveltejs/kit';
import * as v from 'valibot';
import { datosDe } from '#lib/server/datos/index.ts';
import { validarGa4, validarPixel } from '#lib/tienda/medicion.ts';
import type { RequestHandler } from './$types';

const id = v.nullable(v.pipe(v.string(), v.maxLength(40)));
const Medicion = v.object({
	ga4: id,
	pixel: id,
	feeds: v.object({ gm: v.boolean(), meta: v.boolean() })
});

/** GA4, píxel de Meta y feeds cargados: se guarda todo junto (vacío borra el ID). */
export const PUT: RequestHandler = async ({ locals, request }) => {
	const b = v.safeParse(Medicion, await request.json().catch(() => null));
	if (!b.success) error(400, 'Pedido inválido.');
	const ga4 = validarGa4(b.output.ga4);
	if (!ga4.ok) error(400, ga4.error);
	const pixel = validarPixel(b.output.pixel);
	if (!pixel.ok) error(400, pixel.error);
	const r = await datosDe(locals).guardarMedicion({
		ga4: ga4.valor,
		pixel: pixel.valor,
		feeds: b.output.feeds
	});
	if (r !== 'ok') error(400, 'Revisá los IDs.');
	return new Response(null, { status: 204 });
};
