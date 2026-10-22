import { error, json } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** Transferir lo disponible al banco. */
export const POST: RequestHandler = async ({ locals }) => {
	const r = await datosDe(locals).retirar();
	if (r === 'nada') error(409, 'No hay plata disponible para transferir.');
	if (r === 'sin-cobros') error(409, 'Conectá Mercado Pago para transferir.');
	return json(r);
};
