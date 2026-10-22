import { error, json } from '@sveltejs/kit';
import * as v from 'valibot';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

const Invitacion = v.object({
	contacto: v.pipe(
		v.string(),
		v.trim(),
		v.minLength(6, 'Poné un email o un WhatsApp.'),
		v.maxLength(120),
		v.check(
			(s) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(s) || /^\+?[\d\s-]{8,20}$/.test(s),
			'Poné un email o un WhatsApp.'
		)
	),
	rol: v.picklist(['admin', 'ventas', 'deposito'])
});

/** Invitar a alguien al equipo. */
export const POST: RequestHandler = async ({ locals, request }) => {
	const r = v.safeParse(Invitacion, await request.json().catch(() => null));
	if (!r.success) error(400, r.issues[0].message);
	const m = await datosDe(locals).invitar(r.output.contacto, r.output.rol);
	if (m === 'repetido') error(409, 'Esa persona ya está en el equipo.');
	return json(m, { status: 201 });
};
