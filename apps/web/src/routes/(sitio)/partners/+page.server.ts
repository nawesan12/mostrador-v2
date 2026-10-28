import { Postulacion } from '#lib/server/formularios-esquemas.ts';
import { usaSandbox } from '#lib/server/datos/index.ts';
import { procesar } from '#lib/server/formularios.ts';
import type { Actions, PageServerLoad } from './$types';

/** En la demo pública se puede recorrer el panel de partners de ejemplo. */
export const load: PageServerLoad = () => ({ demo: usaSandbox() });

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
