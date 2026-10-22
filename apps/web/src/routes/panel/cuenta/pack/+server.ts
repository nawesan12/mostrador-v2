import { error } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** +2.000 respuestas de Hornero este mes; DELETE lo deshace. */
export const POST: RequestHandler = async ({ locals }) => {
	await datosDe(locals).comprarPack();
	return new Response(null, { status: 204 });
};
export const DELETE: RequestHandler = async ({ locals }) => {
	if ((await datosDe(locals).devolverPack()) === 'no-existe')
		error(404, 'No hay un pack para devolver.');
	return new Response(null, { status: 204 });
};
