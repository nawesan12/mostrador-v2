import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** Publicar el borrador. */
export const POST: RequestHandler = async ({ locals }) => {
	await datosDe(locals).publicarTienda();
	return new Response(null, { status: 204 });
};
