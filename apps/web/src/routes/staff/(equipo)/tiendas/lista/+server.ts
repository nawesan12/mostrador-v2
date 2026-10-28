import { json } from '@sveltejs/kit';
import { leer } from '#lib/panel/listas/filtros.ts';
import { equipoDe } from '#lib/server/staff/index.ts';
import type { RequestHandler } from './$types';

/** Otra ventana de la lista virtual de tiendas (Go chequea el permiso con la sesión). */
export const GET: RequestHandler = async (event) => {
	const f = leer(event.url.searchParams);
	const desde = Math.max(0, Number(event.url.searchParams.get('desde')) || 0);
	const cuantos = Math.min(200, Math.max(1, Number(event.url.searchParams.get('cuantos')) || 60));
	return json(await equipoDe(event).tiendas({ ...f, desde, cuantos }), {
		headers: { 'cache-control': 'no-store' }
	});
};
