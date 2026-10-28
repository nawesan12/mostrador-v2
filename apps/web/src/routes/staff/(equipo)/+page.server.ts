import { equipoDe, exigir } from '#lib/server/staff/index.ts';
import type { PageServerLoad } from './$types';

/** Inicio del equipo: tiendas, plata del mes, pruebas por vencer, bajas y alertas. */
export const load: PageServerLoad = async (event) => {
	const { sesion } = await event.parent();
	exigir(sesion.permisos, 'tiendas.ver');
	return { inicio: await equipoDe(event).inicio() };
};
