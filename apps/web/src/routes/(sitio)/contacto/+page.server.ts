import { Consulta } from '#lib/server/formularios-esquemas.ts';
import { procesar } from '#lib/server/formularios.ts';
import type { Actions } from './$types';

export const actions: Actions = {
	default: async (event) => {
		const f = Object.fromEntries(await event.request.formData());
		const entrada = {
			nombre: String(f.nombre ?? ''),
			email: String(f.email ?? ''),
			whatsapp: String(f.whatsapp ?? ''),
			web: String(f.web ?? ''),
			facturacion: String(f.facturacion ?? ''),
			origen: String(f.origen ?? ''),
			mensaje: String(f.mensaje ?? '')
		};
		const r = await procesar(event, Consulta, entrada, (api, c) => api.consultar(c));
		// La pantalla de éxito saluda por el nombre (el primero).
		return 'enviado' in r ? { ...r, nombre: entrada.nombre.trim().split(/\s+/)[0] } : r;
	}
};
