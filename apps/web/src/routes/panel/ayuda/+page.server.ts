import { env } from 'cloudflare:workers';
import { leerEstado } from '#lib/server/estado/lectura.ts';
import type { PageServerLoad } from './$types';

/** El estado del sistema, para la tarjeta de Ayuda (lo mismo que /estado). */
export const load: PageServerLoad = async () => {
	const e = await leerEstado(env.ESTADO, 90).catch(() => null);
	return {
		estado: e
			? {
					disponibilidad: e.disponibilidad,
					todoBien:
						e.componentes.every((c) => c.ok !== false) && !e.incidentes.some((i) => i.fin === null)
				}
			: null
	};
};
