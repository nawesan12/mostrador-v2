import { error } from '@sveltejs/kit';
import * as v from 'valibot';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

const Accion = v.object({
	accion: v.picklist(['ocultar', 'mostrar', 'responder']),
	respuesta: v.optional(v.pipe(v.string(), v.trim(), v.maxLength(1000)), '')
});

/** Ocultar o mostrar una reseña (no se borra), o responderla en público. */
export const POST: RequestHandler = async ({ locals, params, request }) => {
	const b = v.safeParse(Accion, await request.json().catch(() => null));
	if (!b.success) error(400, 'Revisá la respuesta.');
	const r = await datosDe(locals).accionResena(params.id, b.output.accion, b.output.respuesta);
	if (r === 'no-existe') error(404, 'Esa reseña no existe.');
	if (r === 'no-se-puede') error(400, 'Revisá la respuesta.');
	return new Response(null, { status: 204 });
};
