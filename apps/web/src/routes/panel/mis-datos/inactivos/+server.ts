import { json } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** Programar el borrado de los clientes inactivos; POST lo frena. */
export const DELETE: RequestHandler = async ({ locals }) =>
	json({ clientes: await datosDe(locals).borrarInactivos() });
export const POST: RequestHandler = async ({ locals }) => {
	await datosDe(locals).frenarBorrado();
	return new Response(null, { status: 204 });
};
