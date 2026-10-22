import { error } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { AccionRelampago } from '#lib/server/datos/tipos.ts';
import type { RequestHandler } from './$types';

const ACCIONES: AccionRelampago[] = ['sumar-hora', 'liberar', 'terminar', 'reanudar'];

/** Sumar una hora, liberar 20 unidades, terminar o deshacer el final. */
export const POST: RequestHandler = async ({ locals, request }) => {
	const c = (await request.json().catch(() => null)) as { accion?: unknown } | null;
	const accion = c?.accion as AccionRelampago;
	if (!ACCIONES.includes(accion)) error(400, 'Falta qué hacer.');
	const r = await datosDe(locals).accionRelampago(accion);
	if (r === 'no-existe') error(404, 'No hay una oferta relámpago.');
	if (r === 'no-se-puede') error(409, 'La oferta ya terminó.');
	return new Response(null, { status: 204 });
};
