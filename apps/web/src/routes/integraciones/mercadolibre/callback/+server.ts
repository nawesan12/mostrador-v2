import { redirect } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** Vuelta del OAuth de Mercado Libre (la URL registrada en la aplicación). */
export const GET: RequestHandler = async ({ locals, url }) => {
	const codigo = url.searchParams.get('code') ?? '';
	const estado = url.searchParams.get('state') ?? '';
	const r =
		url.searchParams.has('error') || !codigo || !estado
			? 'no-se-puede'
			: await datosDe(locals).volverMercadoLibre(codigo.slice(0, 512), estado.slice(0, 512));
	redirect(303, `/panel/integraciones?mercadolibre=${r === 'ok' ? 'conectada' : 'error'}`);
};
