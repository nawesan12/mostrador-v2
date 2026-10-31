import { error } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** Elegir qué tienda ver (una persona con varias tiendas). */
export const PUT: RequestHandler = async ({ locals, request }) => {
	const c = (await request.json().catch(() => null)) as { tiendaId?: unknown } | null;
	if (typeof c?.tiendaId !== 'string' || c.tiendaId.length > 64) error(400, 'Elegí una tienda.');
	if ((await datosDe(locals).elegirTienda(c.tiendaId)) !== 'ok')
		error(404, 'No sos parte de esa tienda.');
	return new Response(null, { status: 204 });
};
