import { error, json } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** «Ver anteriores» de la campanita: la página que sigue al cursor (§45). */
export const GET: RequestHandler = async ({ locals, url }) => {
	const cursor = url.searchParams.get('cursor') ?? '';
	if (!/^[\w.-]{1,120}$/.test(cursor)) error(400, 'Falta desde dónde seguir.');
	return json(await datosDe(locals).notificaciones(cursor), {
		headers: { 'cache-control': 'no-store' }
	});
};
