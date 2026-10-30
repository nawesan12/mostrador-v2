import { planVigente } from '#lib/panel/planes.ts';
import { datosDe } from '#lib/server/datos/index.ts';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals }) => {
	const datos = datosDe(locals);
	const [c, p] = await Promise.all([datos.configuracion(), datos.permisos()]);
	return {
		equipo: c.equipo,
		plan: planVigente(c.plan.actual, c.plan.enPrueba),
		permisos: p.permisos
	};
};
