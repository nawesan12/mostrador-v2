import { error } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { Idea } from '#lib/server/datos/tipos.ts';
import type { RequestHandler } from './$types';

const IDEAS: Idea[] = ['carritos', 'stock', 'promo'];

/** Aprobar ({hecha: true}) o deshacer ({hecha: false}) una idea de Tincho. */
export const POST: RequestHandler = async ({ locals, params, request }) => {
	const idea = params.idea as Idea;
	if (!IDEAS.includes(idea)) error(404, 'Esa idea no existe.');
	const cuerpo = (await request.json().catch(() => null)) as { hecha?: unknown } | null;
	if (typeof cuerpo?.hecha !== 'boolean') error(400, 'Falta si la idea queda hecha.');
	await datosDe(locals).marcarIdea(idea, cuerpo.hecha);
	return new Response(null, { status: 204 });
};
