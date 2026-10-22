import { error } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** Pausar o activar un canal. */
export const PUT: RequestHandler = async ({ locals, params, request }) => {
	const c = (await request.json().catch(() => null)) as { activo?: unknown } | null;
	if (typeof c?.activo !== 'boolean') error(400, 'Falta si va activo.');
	const r = await datosDe(locals).fijarCanal(params.id, c.activo);
	if (r === 'no-existe') error(404, 'Ese canal no existe.');
	if (r === 'no-se-puede') error(409, 'Primero conectalo desde Integraciones.');
	return new Response(null, { status: 204 });
};
