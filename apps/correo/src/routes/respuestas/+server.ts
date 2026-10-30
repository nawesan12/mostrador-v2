import { json } from '@sveltejs/kit';
import { leerRespuestas } from '#lib/server/ajustes.ts';
import { Jmap } from '#lib/server/jmap.ts';
import type { RequestHandler } from './$types';

/** Las respuestas guardadas, para insertarlas al redactar. */
export const GET: RequestHandler = async ({ locals }) =>
	json({ respuestas: await leerRespuestas(new Jmap(locals.cred!)) });
