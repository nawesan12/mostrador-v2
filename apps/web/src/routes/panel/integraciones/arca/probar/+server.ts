import { json } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** Probar la conexión: el ticket del WSAA y el último comprobante autorizado. */
export const POST: RequestHandler = async ({ locals }) => json(await datosDe(locals).arcaProbar());
