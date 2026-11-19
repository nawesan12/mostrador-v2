import { fail, redirect } from '@sveltejs/kit';
import { ErrorApi } from '#lib/server/bff.ts';
import { borrarSesion } from '#lib/server/auth/sesion.ts';
import { cuentasDe, usaSandbox } from '#lib/server/datos/index.ts';
import type { Actions, PageServerLoad } from './$types';

/**
 * El link «No fui yo» del mail que avisa que cambiaron adónde llega la plata. No usa el token al
 * abrirse (los lectores de mail abren los links): hace falta tocar el botón.
 */
export const load: PageServerLoad = ({ url }) => {
	if (usaSandbox()) redirect(303, '/prueba');
	const token = url.searchParams.get('token');
	return { token: token && token.length <= 200 ? token : null };
};

export const actions: Actions = {
	default: async (event) => {
		const token = String((await event.request.formData()).get('token') ?? '');
		try {
			const r = await cuentasDe(event).revertirCobro(token);
			// Se cerraron todas las sesiones del equipo, también ésta.
			borrarSesion(event.cookies);
			return { hecho: r.tipo };
		} catch (e) {
			if (e instanceof ErrorApi && e.estado < 500) return fail(e.estado, { mensaje: e.message });
			return fail(503, { mensaje: 'No pudimos volverlo atrás. Probá de nuevo en un momento.' });
		}
	}
};
