import { error, fail, redirect } from '@sveltejs/kit';
import * as v from 'valibot';
import { COOKIE } from '#lib/server/sandbox/sandbox.ts';
import { resolverTienda } from '#lib/server/tienda/publica.ts';
import { reservar } from '#lib/server/tienda/turnos.ts';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ parent }) => {
	const { tienda } = await parent();
	if (!tienda.turnos) error(404, 'Esta tienda no toma turnos online.');
};

const Reserva = v.object({
	servicioId: v.pipe(v.string(), v.regex(/^[A-Za-z0-9_-]{1,64}$/, 'Elegí un servicio.')),
	profesionalId: v.pipe(v.string(), v.regex(/^([A-Za-z0-9_-]{1,64})?$/, 'Elegí quién te atiende.')),
	inicio: v.pipe(
		v.string(),
		v.regex(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:00-03:00$/, 'Elegí un día y un horario.')
	),
	nombre: v.pipe(v.string(), v.trim(), v.minLength(2, 'Poné tu nombre.'), v.maxLength(80)),
	telefono: v.pipe(v.string(), v.trim(), v.regex(/^[\d\s+()-]{8,20}$/, 'Revisá el WhatsApp.')),
	email: v.pipe(v.string(), v.trim(), v.email('Revisá el email.'), v.maxLength(120))
});

export const actions: Actions = {
	default: async ({ request, params, cookies, platform }) => {
		const r = await resolverTienda(params.tienda, cookies.get(COOKIE));
		if (!r || !r.tienda.turnos) error(404, 'Esta tienda no toma turnos online.');
		if (r.tienda.vitrina)
			return fail(409, { error: 'Es una tienda de ejemplo: no se reservan turnos de verdad.' });
		const b = v.safeParse(Reserva, Object.fromEntries(await request.formData()));
		if (!b.success) return fail(400, { error: b.issues[0].message });
		// Precio, duración y seña salen del servicio en el servidor; el horario se vuelve a chequear.
		const res = await reservar(
			r,
			{ ...b.output, profesionalId: b.output.profesionalId || null },
			(p) => platform?.ctx?.waitUntil(p)
		);
		if ('error' in res) return fail(res.estado, { error: res.error });
		if (res.url && /^https:\/\/(www\.)?mercadopago\.com(\.ar)?\//.test(res.url))
			redirect(303, res.url);
		redirect(303, `/t/${r.tienda.slug}/turno/${res.codigo}`);
	}
};
