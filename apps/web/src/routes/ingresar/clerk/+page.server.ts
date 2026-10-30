import { redirect } from '@sveltejs/kit';
import { ErrorApi } from '#lib/server/bff.ts';
import { guardarSesion } from '#lib/server/auth/sesion.ts';
import { destino } from '#lib/server/auth/volver.ts';
import { cuentasDe, usaSandbox } from '#lib/server/datos/index.ts';
import type { Actions, PageServerLoad } from './$types';

/**
 * Vuelta de Google o Apple por Clerk (docs/decisiones.md §31). La página termina el ida y vuelta en
 * el navegador y postea acá el token de sesión de Clerk; Go lo verifica y devuelve la sesión de Tu
 * Local Web, que queda en la cookie httpOnly. Una cuenta nueva sigue al onboarding.
 */
export const load: PageServerLoad = () => {
	if (usaSandbox()) redirect(303, '/ingresar');
};

export const actions: Actions = {
	default: async (event) => {
		const f = await event.request.formData();
		const error = String(f.get('error') ?? '');
		const token = String(f.get('token') ?? '');
		if (usaSandbox()) redirect(303, '/ingresar');
		if (error || !token || token.length > 8192)
			redirect(
				303,
				`/ingresar?error=${encodeURIComponent(/^[a-z_]{1,40}$/.test(error) ? error : 'cancelado')}`
			);
		let hacia: string;
		try {
			const r = await cuentasDe(event).ingresarConClerk(token);
			guardarSesion(event.cookies, r.token, r.venceEn);
			hacia = r.nueva ? '/onboarding?paso=1' : destino(event.cookies);
		} catch (e) {
			if (!(e instanceof ErrorApi)) throw e;
			redirect(303, `/ingresar?error=${encodeURIComponent(e.codigo)}`);
		}
		redirect(303, hacia);
	}
};
