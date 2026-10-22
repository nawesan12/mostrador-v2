import { error } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** Aprobar, ofrecer cambio o deshacer. */
export const POST: RequestHandler = async ({ locals, params, request }) => {
	const c = (await request.json().catch(() => null)) as { accion?: unknown } | null;
	const accion = c?.accion;
	if (accion !== 'aprobar' && accion !== 'cambio' && accion !== 'deshacer')
		error(400, 'Falta qué hacer.');
	if ((await datosDe(locals).resolverDevolucion(params.id, accion)) === 'no-existe')
		error(404, 'Esa devolución no existe.');
	return new Response(null, { status: 204 });
};
