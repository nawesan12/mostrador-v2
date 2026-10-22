import { error, json } from '@sveltejs/kit';
import * as v from 'valibot';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

const Nuevo = v.object({
	monto: v.pipe(v.number(), v.integer(), v.minValue(1, 'Poné el monto.'), v.maxValue(100_000_000)),
	concepto: v.pipe(v.string(), v.trim(), v.minLength(2, 'Escribí qué se cobra.'), v.maxLength(140)),
	cuotas: v.optional(v.boolean(), true),
	pedido: v.optional(v.nullable(v.pipe(v.number(), v.integer())), null)
});

/** Un link de pago nuevo. */
export const POST: RequestHandler = async ({ locals, request }) => {
	const r = v.safeParse(Nuevo, await request.json().catch(() => null));
	if (!r.success) error(400, r.issues[0].message);
	return json({ codigo: await datosDe(locals).crearLink(r.output) }, { status: 201 });
};
