import { datosDe } from '#lib/server/datos/index.ts';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals }) => {
	const d = datosDe(locals);
	const [ajustes, servicios, profesionales, mp] = await Promise.all([
		d.ajustesTurnos(),
		d.serviciosTurnos(),
		d.profesionalesTurnos(),
		d.mercadoPago()
	]);
	return { ajustes, servicios, profesionales, mercadoPago: mp.conectada };
};
