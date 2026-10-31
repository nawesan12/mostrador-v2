import { json } from '@sveltejs/kit';
import { filtroDe } from '#lib/panel/actividad.ts';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** La página siguiente de la Actividad (el botón «Ver más»), con los mismos filtros. */
export const GET: RequestHandler = async ({ locals, url }) =>
	json(await datosDe(locals).actividad(filtroDe(url.searchParams)), {
		headers: { 'cache-control': 'no-store' }
	});
