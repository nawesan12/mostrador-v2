import { redirect } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/**
 * Vuelta del OAuth de Mercado Pago (la URL registrada en la aplicación). El state de un solo uso y el
 * canje del código los hace la API; acá sólo se pasa y se vuelve al panel con el resultado.
 */
export const GET: RequestHandler = async ({ locals, url }) => {
	const codigo = url.searchParams.get('code') ?? '';
	const estado = url.searchParams.get('state') ?? '';
	// La persona canceló en Mercado Pago, o falta algo.
	const r =
		url.searchParams.has('error') || !codigo || !estado
			? 'no-se-puede'
			: await datosDe(locals).volverMercadoPago(codigo.slice(0, 512), estado.slice(0, 512));
	redirect(303, `/panel/integraciones?mercadopago=${r === 'ok' ? 'conectada' : 'error'}`);
};
