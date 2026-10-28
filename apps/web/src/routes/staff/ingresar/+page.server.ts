import { fail, redirect } from '@sveltejs/kit';
import { ErrorApi } from '#lib/server/bff.ts';
import { usaSandbox } from '#lib/server/datos/index.ts';
import { accesoEquipo, guardarSesionEquipo } from '#lib/server/staff/index.ts';
import type { Actions, PageServerLoad } from './$types';

/** El ingreso del equipo (no es el de las tiendas). En la demo no hace falta: se entra directo. */
export const load: PageServerLoad = async ({ locals }) => {
	if (usaSandbox() || locals.equipo) redirect(303, '/staff');
};

export const actions: Actions = {
	default: async (event) => {
		const f = await event.request.formData();
		const email = String(f.get('email') ?? '').trim();
		const clave = String(f.get('clave') ?? '');
		if (!email || !clave) return fail(400, { email, error: 'Escribí tu email y tu contraseña.' });
		try {
			const s = await accesoEquipo(event).ingresar(email, clave);
			guardarSesionEquipo(event.cookies, s.token, s.venceEn);
		} catch (e) {
			if (e instanceof ErrorApi) return fail(e.estado, { email, error: e.message });
			throw e;
		}
		redirect(303, '/staff');
	}
};
