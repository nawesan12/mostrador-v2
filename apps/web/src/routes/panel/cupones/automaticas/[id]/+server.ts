import { error } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** Prender o apagar una promoción automática. */
export const PUT: RequestHandler = async ({ locals, params, request }) => {
	const c = (await request.json().catch(() => null)) as { activa?: unknown } | null;
	if (typeof c?.activa !== 'boolean') error(400, 'Falta si va prendida.');
	if ((await datosDe(locals).fijarAutomatica(params.id, c.activa)) === 'no-existe')
		error(404, 'Esa promoción no existe.');
	return new Response(null, { status: 204 });
};
