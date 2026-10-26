import { error, json } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import { cuerpo, Profesional } from '#lib/server/turnos-esquemas.ts';
import type { RequestHandler } from './$types';

export const PUT: RequestHandler = async ({ locals, params, request }) => {
	const r = await datosDe(locals).guardarProfesional(params.id, await cuerpo(request, Profesional));
	if (!r) error(404, 'Ese profesional ya no existe.');
	return json(r);
};

export const DELETE: RequestHandler = async ({ locals, params }) => {
	if ((await datosDe(locals).borrarProfesional(params.id)) === 'no-existe')
		error(404, 'Ese profesional ya no existe.');
	return new Response(null, { status: 204 });
};
