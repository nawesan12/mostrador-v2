import { error } from '@sveltejs/kit';
import * as v from 'valibot';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

const IDS = ['hero', 'cats', 'prods', 'revs', 'ig', 'wa'] as const;
const Diseno = v.object({
	secciones: v.pipe(
		v.array(v.object({ id: v.picklist(IDS), visible: v.boolean() })),
		v.check(
			(s) => s.length === IDS.length && new Set(s.map((x) => x.id)).size === IDS.length,
			'Faltan secciones.'
		)
	),
	paleta: v.pipe(v.number(), v.integer(), v.minValue(0), v.maxValue(3))
});

/** Guardar el borrador del diseño. */
export const PUT: RequestHandler = async ({ locals, request }) => {
	const r = v.safeParse(Diseno, await request.json().catch(() => null));
	if (!r.success) error(400, r.issues[0].message);
	await datosDe(locals).guardarMiTienda(r.output);
	return new Response(null, { status: 204 });
};
