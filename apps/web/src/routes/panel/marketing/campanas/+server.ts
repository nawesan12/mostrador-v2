import { error, json } from '@sveltejs/kit';
import * as v from 'valibot';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

const Nueva = v.object({
	audiencia: v.picklist(['todos', 'vip', 'riesgo', 'categoria']),
	mensaje: v.pipe(v.string(), v.trim(), v.minLength(10, 'Escribí el mensaje.'), v.maxLength(1000)),
	foto: v.boolean(),
	boton: v.boolean(),
	cupon: v.boolean(),
	programar: v.optional(v.boolean(), false)
});

/** Mandar o programar una campaña. */
export const POST: RequestHandler = async ({ locals, request }) => {
	const r = v.safeParse(Nueva, await request.json().catch(() => null));
	if (!r.success) error(400, r.issues[0].message);
	const c = await datosDe(locals).crearCampana(r.output);
	if (c === 'sin-audiencia') error(409, 'No hay nadie con WhatsApp en ese grupo.');
	if (c === 'sin-whatsapp') error(409, 'Conectá WhatsApp Business para mandar campañas.');
	return json(c, { status: 201 });
};
