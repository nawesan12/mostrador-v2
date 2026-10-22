import { json } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import { leerFiltros } from '#lib/panel/pedidos/filtros.ts';
import type { RequestHandler } from './$types';

/** Otra ventana de la lista virtual (mismos filtros, `desde` y `cuantos`). */
export const GET: RequestHandler = async ({ locals, url }) => {
	const f = leerFiltros(url.searchParams);
	const desde = Math.max(0, Number(url.searchParams.get('desde')) || 0);
	const cuantos = Math.min(200, Math.max(1, Number(url.searchParams.get('cuantos')) || 60));
	return json(await datosDe(locals).listarPedidos({ ...f, desde, cuantos }), {
		headers: { 'cache-control': 'no-store' }
	});
};
