import { datosDe } from '#lib/server/datos/index.ts';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals }) => {
	const datos = datosDe(locals);
	const [m, t] = await Promise.all([datos.misDatos(), datos.tienda()]);
	return { ...m, tienda: t.nombre };
};
