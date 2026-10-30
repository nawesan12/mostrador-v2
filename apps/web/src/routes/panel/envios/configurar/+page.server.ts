import { datosDe } from '#lib/server/datos/index.ts';
import type { PageServerLoad } from './$types';

/** Envíos y retiro: lo guardado, si Correo está conectado y el envío gratis, y el negocio (por defecto). */
export const load: PageServerLoad = async ({ locals }) => {
	const datos = datosDe(locals);
	const [entregas, conf] = await Promise.all([datos.entregas(), datos.configuracion()]);
	return {
		entregas,
		negocio: { direccion: conf.negocio.direccion ?? '', horario: conf.negocio.horario ?? '' }
	};
};
