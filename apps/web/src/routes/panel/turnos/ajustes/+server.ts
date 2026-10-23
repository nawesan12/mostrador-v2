import { json } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import { Ajustes, cuerpo } from '#lib/server/turnos-esquemas.ts';
import type { RequestHandler } from './$types';

export const PUT: RequestHandler = async ({ locals, request }) =>
	json(await datosDe(locals).guardarAjustesTurnos(await cuerpo(request, Ajustes)));
