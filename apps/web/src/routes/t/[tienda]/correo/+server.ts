import { error, json } from '@sveltejs/kit';
import { ErrorApi } from '#lib/server/bff.ts';
import { correoDe } from '#lib/server/correo-argentino.ts';
import { usaSandbox } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/**
 * La cotización real de Correo Argentino para el checkout (Go le pregunta a MiCorreo). En la demo no
 * hay cuenta de MiCorreo: no se cotiza. Nunca se responde un precio que no haya dado Correo.
 */
export const GET: RequestHandler = async (event) => {
	if (usaSandbox()) error(409, 'Esta tienda no envía por Correo Argentino.');
	const cp = event.url.searchParams.get('cp') ?? '';
	const unidades = Number(event.url.searchParams.get('unidades') ?? '1');
	if (!/^\d{4}$/.test(cp)) error(400, 'El código postal tiene que tener 4 números.');
	try {
		const c = await correoDe(event).cotizar(
			event.params.tienda,
			cp,
			Number.isSafeInteger(unidades) && unidades > 0 ? Math.min(unidades, 4950) : 1
		);
		return json(c, { headers: { 'cache-control': 'no-store' } });
	} catch (e) {
		if (e instanceof ErrorApi) error(e.estado, e.message);
		throw e;
	}
};
