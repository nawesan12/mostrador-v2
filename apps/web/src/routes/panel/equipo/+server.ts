import { error, json } from '@sveltejs/kit';
import * as v from 'valibot';
import { INVITABLES } from '#lib/panel/permisos.ts';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

const Invitacion = v.object({
	contacto: v.pipe(
		v.string(),
		v.trim(),
		v.toLowerCase(),
		v.minLength(6, 'Poné un email.'),
		v.maxLength(120, 'Ese email es demasiado largo.'),
		v.check((s) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(s), 'Ese email no parece válido.')
	),
	rol: v.picklist(INVITABLES, 'Elegí un rol.')
});

/** Invitar a alguien al equipo: le llega un mail con el link (sólo la dueña; Go lo vuelve a mirar). */
export const POST: RequestHandler = async ({ locals, request }) => {
	const r = v.safeParse(Invitacion, await request.json().catch(() => null));
	if (!r.success) error(400, r.issues[0].message);
	const m = await datosDe(locals).invitar(r.output.contacto, r.output.rol);
	if (m === 'repetido') error(409, 'Esa persona ya está en el equipo o ya tiene una invitación.');
	return json(m, { status: 201 });
};
