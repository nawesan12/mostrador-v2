import { Postulacion } from '#lib/server/formularios-esquemas.ts';
import { procesar } from '#lib/server/formularios.ts';
import type { Actions } from './$types';

export const actions: Actions = {
	default: async (event) => {
		const f = await event.request.formData();
		const entrada = {
			nombre: String(f.get('nombre') ?? ''),
			email: String(f.get('email') ?? ''),
			agencia: String(f.get('agencia') ?? ''),
			web: String(f.get('web') ?? ''),
			especialidades: f.getAll('especialidades').map(String)
		};
		return procesar(event, Postulacion, entrada, (api, p) => api.postular(p));
	}
};
