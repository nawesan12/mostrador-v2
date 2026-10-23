import { error, json } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** Devuelve la seña con Mercado Pago (cancelar no la devuelve sola). */
export const POST: RequestHandler = async ({ locals, params }) => {
	const t = await datosDe(locals).devolverSena(params.id);
	if (t === 'no-existe') error(404, 'Ese turno no existe.');
	if (t === 'sin-sena') error(409, 'Este turno no tiene una seña pagada para devolver.');
	if (t === 'sin-cobros') error(409, 'Conectá Mercado Pago para devolver la seña.');
	return json(t);
};
