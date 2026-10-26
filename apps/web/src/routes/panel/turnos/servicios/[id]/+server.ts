import { error, json } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import { cuerpo, Servicio } from '#lib/server/turnos-esquemas.ts';
import type { RequestHandler } from './$types';

export const PUT: RequestHandler = async ({ locals, params, request }) => {
	const r = await datosDe(locals).guardarServicio(params.id, await cuerpo(request, Servicio));
	if (!r) error(404, 'Ese servicio ya no existe.');
	return json(r);
};

export const DELETE: RequestHandler = async ({ locals, params }) => {
	if ((await datosDe(locals).borrarServicio(params.id)) === 'no-existe')
		error(404, 'Ese servicio ya no existe.');
	return new Response(null, { status: 204 });
};
