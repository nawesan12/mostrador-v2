import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** "Marcar todo como leído". */
export const POST: RequestHandler = async ({ locals }) => {
	await datosDe(locals).marcarNotificacionesLeidas();
	return new Response(null, { status: 204 });
};
