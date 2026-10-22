import { error } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** Prender o apagar un permiso de un rol. */
export const PUT: RequestHandler = async ({ locals, params, request }) => {
	const rol = params.rol;
	if (rol !== 'admin' && rol !== 'ventas' && rol !== 'deposito')
		error(409, 'Lo del dueño no se cambia.');
	const c = (await request.json().catch(() => null)) as { activo?: unknown } | null;
	if (typeof c?.activo !== 'boolean') error(400, 'Falta si va activo.');
	const r = await datosDe(locals).fijarPermiso(params.permiso, rol, c.activo);
	if (r === 'no-existe') error(404, 'Ese permiso no existe.');
	if (r === 'no-se-puede') error(409, 'Ese permiso es sólo del dueño.');
	return new Response(null, { status: 204 });
};
