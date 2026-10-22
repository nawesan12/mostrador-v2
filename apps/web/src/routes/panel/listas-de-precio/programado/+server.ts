import { error } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** Cancelar el aumento programado. */
export const DELETE: RequestHandler = async ({ locals }) => {
	if ((await datosDe(locals).cancelarAumento()) === 'no-existe')
		error(404, 'No hay un aumento programado.');
	return new Response(null, { status: 204 });
};
