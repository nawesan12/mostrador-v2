import { datosDe } from '#lib/server/datos/index.ts';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals }) => {
	const datos = datosDe(locals);
	const [c, p] = await Promise.all([datos.configuracion(), datos.permisos()]);
	return { equipo: c.equipo, plan: c.plan.actual, permisos: p.permisos };
};
