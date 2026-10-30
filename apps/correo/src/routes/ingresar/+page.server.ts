import { fail, redirect } from '@sveltejs/kit';
import { Jmap, ErrorJmap } from '#lib/server/jmap.ts';
import { guardar } from '#lib/server/sesion.ts';
import { marca } from '#lib/server/marca.ts';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = ({ locals }) => {
	if (locals.cred) redirect(303, '/');
	return { marca: marca() };
};

export const actions: Actions = {
	default: async ({ request, cookies, url }) => {
		const f = await request.formData();
		const email = String(f.get('email') ?? '')
			.trim()
			.toLowerCase();
		const clave = String(f.get('clave') ?? '');
		if (!email || !clave) return fail(400, { email, error: 'Escribí tu casilla y tu clave.' });
		try {
			await new Jmap({ email, clave }).cuentas();
		} catch (e) {
			const msg =
				e instanceof ErrorJmap && e.estado === 401
					? e.message
					: 'No pudimos conectar con el correo. Probá de nuevo.';
			return fail(e instanceof ErrorJmap && e.estado === 401 ? 401 : 502, { email, error: msg });
		}
		guardar(cookies, { email, clave }, url.protocol === 'https:');
		redirect(303, '/');
	}
};
