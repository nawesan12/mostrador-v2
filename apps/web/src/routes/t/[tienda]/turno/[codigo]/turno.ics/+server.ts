import { error } from '@sveltejs/kit';
import { usaSandbox } from '#lib/server/datos/index.ts';
import { ics, turnoDe } from '#lib/server/tienda/turnos.ts';
import type { RequestHandler } from './$types';

/** "Agregar al calendario": el turno como evento (.ics), armado al vuelo. */
export const GET: RequestHandler = async ({ params, url }) => {
	const t = await turnoDe(params.tienda, params.codigo, usaSandbox());
	if (!t) error(404, 'No encontramos ese turno.');
	const pagina = `${url.origin}/t/${params.tienda}/turno/${params.codigo}`;
	return new Response(ics(t, pagina), {
		headers: {
			'content-type': 'text/calendar; charset=utf-8',
			'content-disposition': 'attachment; filename="turno.ics"',
			'cache-control': 'private, no-store'
		}
	});
};
