import { fail } from '@sveltejs/kit';
import { ErrorApi } from '#lib/server/bff.ts';
import { formulariosDe, usaSandbox } from '#lib/server/datos/index.ts';
import type { Actions, PageServerLoad } from './$types';

// El link del mail sólo muestra el botón: dar de baja es un POST, así los clientes de correo que
// abren los links para previsualizarlos no desuscriben a nadie.
export const load: PageServerLoad = ({ url }) => ({ token: url.searchParams.get('token') ?? '' });

export const actions: Actions = {
	default: async (event) => {
		const token = String((await event.request.formData()).get('token') ?? '');
		if (!token) return fail(400, { error: 'Ese link de baja está incompleto.' });
		if (usaSandbox()) return { ok: true };
		try {
			await formulariosDe(event).baja(token);
			return { ok: true };
		} catch (e) {
			if (!(e instanceof ErrorApi)) throw e;
			return fail(e.estado >= 500 ? 503 : e.estado, { error: e.message });
		}
	}
};
