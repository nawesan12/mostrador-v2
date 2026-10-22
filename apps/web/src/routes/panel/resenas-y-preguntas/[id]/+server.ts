import { error } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** Publicar la respuesta a una pregunta. */
export const POST: RequestHandler = async ({ locals, params, request }) => {
	const c = (await request.json().catch(() => null)) as { respuesta?: unknown } | null;
	const respuesta = typeof c?.respuesta === 'string' ? c.respuesta.trim() : '';
	if (respuesta.length < 2 || respuesta.length > 2000) error(400, 'Escribí la respuesta.');
	if ((await datosDe(locals).responderPregunta(params.id, respuesta)) === 'no-existe')
		error(404, 'Esa pregunta no existe.');
	return new Response(null, { status: 204 });
};
