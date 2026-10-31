import { datosDe } from '#lib/server/datos/index.ts';
import { diaAr, diaSemana, fechaValida, sumarDias } from '#lib/turnos/libres.ts';
import type { PageServerLoad } from './$types';

/** La agenda: un día o la semana (de lunes a domingo) que lo contiene. */
export const load: PageServerLoad = async ({ locals, url, depends }) => {
	// Tiempo real (§44): se recarga con los avisos de lo que muestra.
	depends('tienda:turnos');
	const q = url.searchParams.get('dia') ?? '';
	const dia = fechaValida(q) ? q : diaAr(Date.now());
	const vista = url.searchParams.get('vista') === 'semana' ? 'semana' : 'dia';
	const desde = vista === 'semana' ? sumarDias(dia, -((diaSemana(dia) + 6) % 7)) : dia;
	const hasta = vista === 'semana' ? sumarDias(desde, 6) : dia;
	const datos = datosDe(locals);
	const [turnos, servicios, profesionales, ajustes] = await Promise.all([
		datos.agendaTurnos(desde, hasta),
		datos.serviciosTurnos(),
		datos.profesionalesTurnos(),
		datos.ajustesTurnos()
	]);
	return {
		dia,
		vista,
		desde,
		hasta,
		hoy: diaAr(Date.now()),
		turnos,
		servicios,
		profesionales,
		ajustes
	};
};
