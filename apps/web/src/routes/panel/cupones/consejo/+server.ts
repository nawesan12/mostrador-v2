import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** "Ahora no" en el consejo de Hornero. */
export const DELETE: RequestHandler = async ({ locals }) => {
	await datosDe(locals).descartarConsejo();
	return new Response(null, { status: 204 });
};
