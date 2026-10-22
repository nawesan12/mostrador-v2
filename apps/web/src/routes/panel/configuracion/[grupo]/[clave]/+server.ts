import { error } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** Prender o apagar una opción de seguridad o un aviso. */
export const PUT: RequestHandler = async ({ locals, params, request }) => {
	if (params.grupo !== 'seguridad' && params.grupo !== 'avisos') error(404, 'No existe.');
	const c = (await request.json().catch(() => null)) as { activo?: unknown } | null;
	if (typeof c?.activo !== 'boolean') error(400, 'Falta si va activo.');
	if ((await datosDe(locals).fijarOpcion(params.grupo, params.clave, c.activo)) === 'no-existe')
		error(404, 'Esa opción no existe.');
	return new Response(null, { status: 204 });
};
