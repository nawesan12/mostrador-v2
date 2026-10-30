import { error } from '@sveltejs/kit';
import * as v from 'valibot';
import { EsquemaEntregas } from '#lib/panel/entregas.ts';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** Guarda retiro y moto (que quede al menos una forma de entrega lo validan Go y el sandbox). */
export const PUT: RequestHandler = async ({ locals, request }) => {
	const r = v.safeParse(EsquemaEntregas, await request.json().catch(() => null));
	if (!r.success) error(400, r.issues[0].message);
	await datosDe(locals).guardarEntregas(r.output);
	return new Response(null, { status: 204 });
};
