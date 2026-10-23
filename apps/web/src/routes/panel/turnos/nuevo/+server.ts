import { error, json } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import { cuerpo, TurnoNuevo } from '#lib/server/turnos-esquemas.ts';
import type { RequestHandler } from './$types';

/** Un turno cargado a mano (sin seña). 409 si se superpone y no se pidió tomarlo igual. */
export const POST: RequestHandler = async ({ locals, request }) => {
	const n = await cuerpo(request, TurnoNuevo);
	const t = await datosDe(locals).crearTurno(n);
	if (t === 'ocupado') error(409, 'Ese horario se superpone con otro turno.');
	if (t === 'invalido') error(400, 'Elegí un servicio y alguien que atienda.');
	return json(t, { status: 201 });
};
