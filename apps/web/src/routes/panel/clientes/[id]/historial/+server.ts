import { json } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** Los pedidos importados de otra plataforma de un cliente (la ficha los muestra con su plataforma). */
export const GET: RequestHandler = async ({ locals, params }) =>
	json(await datosDe(locals).historialCliente(params.id.slice(0, 200)), {
		headers: { 'cache-control': 'no-store' }
	});
