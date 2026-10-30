import { json } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** Cerrar todas las sesiones menos esta. */
export const DELETE: RequestHandler = async ({ locals }) =>
	json({ cerradas: await datosDe(locals).cerrarOtrasSesiones() });
