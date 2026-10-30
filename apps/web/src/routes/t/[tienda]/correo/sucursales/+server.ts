import { error, json } from '@sveltejs/kit';
import { ErrorApi } from '#lib/server/bff.ts';
import { correoDe } from '#lib/server/correo-argentino.ts';
import { usaSandbox } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** Las sucursales de Correo Argentino de una provincia, para elegir dónde retirar. */
export const GET: RequestHandler = async (event) => {
	if (usaSandbox()) error(409, 'Esta tienda no envía por Correo Argentino.');
	const provincia = (event.url.searchParams.get('provincia') ?? '').toUpperCase();
	const cp = event.url.searchParams.get('cp') ?? '';
	if (!/^[A-Z]$/.test(provincia)) error(400, 'Elegí la provincia.');
	try {
		const l = await correoDe(event).sucursales(
			event.params.tienda,
			provincia,
			/^\d{4}$/.test(cp) ? cp : ''
		);
		return json(l, { headers: { 'cache-control': 'private, max-age=600' } });
	} catch (e) {
		if (e instanceof ErrorApi) error(e.estado, e.message);
		throw e;
	}
};
