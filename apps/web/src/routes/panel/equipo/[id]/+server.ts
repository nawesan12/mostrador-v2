import { error } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

const resp = (r: string) => {
	if (r === 'no-existe') error(404, 'Esa persona no está en el equipo.');
	if (r === 'no-se-puede') error(409, 'A la dueña no se le cambia el rol.');
	return new Response(null, { status: 204 });
};

/** Cambiar el rol. */
export const PUT: RequestHandler = async ({ locals, params, request }) => {
	const c = (await request.json().catch(() => null)) as { rol?: unknown } | null;
	const rol = c?.rol;
	if (rol !== 'admin' && rol !== 'ventas' && rol !== 'deposito') error(400, 'Elegí un rol.');
	return resp(await datosDe(locals).cambiarRol(params.id, rol));
};

/** Quitar del equipo (o retirar la invitación). */
export const DELETE: RequestHandler = async ({ locals, params }) =>
	resp(await datosDe(locals).quitarDelEquipo(params.id));
