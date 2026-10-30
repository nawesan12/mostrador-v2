import { error } from '@sveltejs/kit';
import * as v from 'valibot';
import { plantilla } from '#lib/plantillas.ts';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

// Las fotos subidas a /panel/fotos vuelven con su origen (https en producción, http en local).
const foto = v.pipe(
	v.string(),
	v.maxLength(300),
	v.regex(/^$|^https?:\/\/[^\s"'<>]+$/, 'Esa foto no es válida: subila de nuevo.')
);
const IDS = ['hero', 'cats', 'prods', 'revs', 'ig', 'wa'] as const;
const Diseno = v.object({
	secciones: v.pipe(
		v.array(v.object({ id: v.picklist(IDS), visible: v.boolean() })),
		v.check(
			(s) => s.length === IDS.length && new Set(s.map((x) => x.id)).size === IDS.length,
			'Faltan secciones.'
		)
	),
	paleta: v.pipe(v.number(), v.integer(), v.minValue(0), v.maxValue(3)),
	plantilla: v.optional(
		v.pipe(
			v.string(),
			v.check((id) => id === '' || !!plantilla(id), 'Elegí una de las plantillas.')
		)
	),
	titulo: v.optional(
		v.pipe(v.string(), v.trim(), v.maxLength(80, 'El título va hasta 80 letras.'))
	),
	lema: v.optional(v.pipe(v.string(), v.trim(), v.maxLength(120, 'El lema va hasta 120 letras.'))),
	portada: v.optional(foto),
	logo: v.optional(foto),
	compartirTitulo: v.optional(
		v.pipe(v.string(), v.trim(), v.maxLength(70, 'El título para compartir va hasta 70 letras.'))
	),
	compartirDescripcion: v.optional(
		v.pipe(
			v.string(),
			v.trim(),
			v.maxLength(160, 'La descripción para compartir va hasta 160 letras.')
		)
	),
	compartirFoto: v.optional(foto)
});

/** Guardar el borrador del diseño. */
export const PUT: RequestHandler = async ({ locals, request }) => {
	const r = v.safeParse(Diseno, await request.json().catch(() => null));
	if (!r.success) error(400, r.issues[0].message);
	await datosDe(locals).guardarMiTienda(r.output);
	return new Response(null, { status: 204 });
};
