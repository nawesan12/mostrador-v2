import { error } from '@sveltejs/kit';
import { INVITABLES, type RolInvitable } from '#lib/panel/permisos.ts';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

const resp = (r: string) => {
	if (r === 'no-existe') error(404, 'Esa persona no está en el equipo.');
	if (r === 'no-se-puede') error(409, 'A la dueña no se la quita ni se le cambia el rol.');
	return new Response(null, { status: 204 });
};

/** Cambiar el rol (de alguien del equipo o de una invitación pendiente). */
export const PUT: RequestHandler = async ({ locals, params, request }) => {
	const c = (await request.json().catch(() => null)) as { rol?: unknown } | null;
	const rol = c?.rol;
	if (!INVITABLES.includes(rol as RolInvitable)) error(400, 'Elegí un rol.');
	return resp(await datosDe(locals).cambiarRol(params.id, rol as RolInvitable));
};

/** Quitar del equipo (cierra sus sesiones en esta tienda) o cancelar la invitación. */
export const DELETE: RequestHandler = async ({ locals, params }) =>
	resp(await datosDe(locals).quitarDelEquipo(params.id));
