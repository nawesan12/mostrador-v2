import { error, json } from '@sveltejs/kit';
import * as v from 'valibot';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

const Aumento = v.object({
	pct: v.pipe(v.number(), v.integer(), v.minValue(1, 'Elegí cuánto subir.'), v.maxValue(50)),
	redondeo: v.picklist(['no', '90', '500']),
	programar: v.optional(v.boolean(), false)
});

/** Subir todos los precios (o programarlo); DELETE deshace el último. */
export const POST: RequestHandler = async ({ locals, request }) => {
	const r = v.safeParse(Aumento, await request.json().catch(() => null));
	if (!r.success) error(400, r.issues[0].message);
	return json(
		await datosDe(locals).aumentarPrecios(r.output.pct, r.output.redondeo, r.output.programar)
	);
};

export const DELETE: RequestHandler = async ({ locals }) => {
	if ((await datosDe(locals).deshacerAumento()) === 'no-existe')
		error(404, 'No hay un aumento para deshacer.');
	return new Response(null, { status: 204 });
};
