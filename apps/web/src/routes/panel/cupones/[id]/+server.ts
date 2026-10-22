import { error, json } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { AccionCupon } from '#lib/server/datos/tipos.ts';
import type { RequestHandler } from './$types';

const ACCIONES: AccionCupon[] = ['pausar', 'activar', 'duplicar', 'minimo', 'deshacer-minimo'];

/** Pausar, activar, duplicar o ponerle un mínimo. */
export const POST: RequestHandler = async ({ locals, params, request }) => {
	const c = (await request.json().catch(() => null)) as {
		accion?: unknown;
		minimo?: unknown;
	} | null;
	const accion = c?.accion as AccionCupon;
	if (!ACCIONES.includes(accion)) error(400, 'Falta qué hacer.');
	const minimo = Number(c?.minimo ?? 0);
	if (accion === 'minimo' && !(Number.isInteger(minimo) && minimo > 0 && minimo <= 10_000_000))
		error(400, 'Poné el mínimo de compra.');
	const r = await datosDe(locals).accionCupon(params.id, accion, minimo);
	if (r === 'no-existe') error(404, 'Ese cupón no existe.');
	if (r === 'no-se-puede') error(409, 'Ese cupón ya terminó. Duplicalo para usarlo de nuevo.');
	return json(r);
};
