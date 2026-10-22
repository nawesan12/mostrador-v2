import { error } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** Cerrar otra sesión. */
export const DELETE: RequestHandler = async ({ locals, params }) => {
	const r = await datosDe(locals).cerrarSesion(params.id);
	if (r === 'no-existe') error(404, 'Esa sesión no existe.');
	if (r === 'no-se-puede') error(409, 'Para cerrar esta, salí de la cuenta.');
	return new Response(null, { status: 204 });
};
