import { error } from '@sveltejs/kit';
import * as v from 'valibot';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

const Prefs = v.object({
	matriz: v.pipe(
		v.record(v.pipe(v.string(), v.regex(/^[a-z]+:(panel|whatsapp|email|celular)$/)), v.boolean()),
		v.check((m) => Object.keys(m).length <= 60)
	),
	silencio: v.boolean(),
	resumen: v.picklist(['whatsapp', 'email', 'off'])
});

/** Qué avisos llegan y por dónde. */
export const PUT: RequestHandler = async ({ locals, request }) => {
	const r = v.safeParse(Prefs, await request.json().catch(() => null));
	if (!r.success) error(400, r.issues[0].message);
	await datosDe(locals).guardarPreferenciasAvisos(r.output);
	return new Response(null, { status: 204 });
};
