import { error } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** «Mandame un aviso de prueba»: lo manda el worker a la suscripción de esta sesión. */
export const POST: RequestHandler = async ({ locals }) => {
	if ((await datosDe(locals).probarPushWeb()) !== 'ok')
		error(409, 'Primero activá los avisos en este celular.');
	return new Response(null, { status: 202 });
};
