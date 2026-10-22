import { error } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** Marcar o desmarcar una tarea de hoy. */
export const PUT: RequestHandler = async ({ locals, params, request }) => {
	const c = (await request.json().catch(() => null)) as { hecha?: unknown } | null;
	if (typeof c?.hecha !== 'boolean') error(400, 'Falta si está hecha.');
	if ((await datosDe(locals).marcarTarea(params.id, c.hecha)) === 'no-existe')
		error(404, 'Esa tarea no existe.');
	return new Response(null, { status: 204 });
};
