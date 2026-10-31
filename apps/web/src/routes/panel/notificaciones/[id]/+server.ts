import { error } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** Tocar una notificación de la campanita la marca leída (§45). */
export const POST: RequestHandler = async ({ locals, params }) => {
	if (!/^[\w-]{1,80}$/.test(params.id)) error(400, 'Esa notificación no existe.');
	await datosDe(locals).marcarNotificacionLeida(params.id);
	return new Response(null, { status: 204 });
};
