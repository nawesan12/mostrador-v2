import { error, json } from '@sveltejs/kit';
import * as v from 'valibot';
import { datosDe } from '#lib/server/datos/index.ts';
import { cuerpo, ESTADOS } from '#lib/server/turnos-esquemas.ts';
import type { RequestHandler } from './$types';

/** Confirmar, cancelar, atendido o ausente (y volver atrás con Deshacer). */
export const PATCH: RequestHandler = async ({ locals, params, request }) => {
	const { estado } = await cuerpo(request, v.object({ estado: v.picklist(ESTADOS) }));
	const t = await datosDe(locals).estadoTurno(params.id, estado);
	if (t === 'no-existe') error(404, 'Ese turno no existe.');
	return json(t);
};
