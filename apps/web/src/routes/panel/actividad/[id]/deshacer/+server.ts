import { error } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** Deshacer un cambio de la actividad. */
export const POST: RequestHandler = async ({ locals, params }) => {
	const r = await datosDe(locals).deshacerEvento(params.id);
	if (r === 'no-existe') error(404, 'Ese cambio no existe.');
	if (r === 'no-se-puede') error(409, 'Ese cambio no se puede deshacer.');
	return new Response(null, { status: 204 });
};
