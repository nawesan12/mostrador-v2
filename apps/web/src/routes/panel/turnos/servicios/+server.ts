import { json } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import { cuerpo, Servicio } from '#lib/server/turnos-esquemas.ts';
import type { RequestHandler } from './$types';

export const POST: RequestHandler = async ({ locals, request }) =>
	json(await datosDe(locals).guardarServicio(null, await cuerpo(request, Servicio)), {
		status: 201
	});
