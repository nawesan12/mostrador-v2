import { Suscripcion } from '#lib/server/formularios-esquemas.ts';
import { procesar } from '#lib/server/formularios.ts';
import type { Actions } from './$types';

export const actions: Actions = {
	default: async (event) => {
		const f = await event.request.formData();
		return procesar(event, Suscripcion, { email: String(f.get('email') ?? '') }, (api, s) =>
			api.suscribir(s.email)
		);
	}
};
