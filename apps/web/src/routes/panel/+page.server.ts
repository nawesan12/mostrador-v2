import { fail } from '@sveltejs/kit';
import { INTEGRACIONES } from '#lib/panel/integraciones.ts';
import { datosDe, usaSandbox } from '#lib/server/datos/index.ts';
import type { Periodo } from '#lib/server/datos/tipos.ts';
import type { Actions, PageServerLoad } from './$types';

const PERIODOS: Periodo[] = ['hoy', '7d', '30d'];

/** El Resumen (dashboard): el período sale de la URL (?periodo=hoy|7d|30d; 7 días si no está). */
export const load: PageServerLoad = async ({ locals, url }) => {
	const p = url.searchParams.get('periodo');
	const periodo = PERIODOS.includes(p as Periodo) ? (p as Periodo) : '7d';
	const d = datosDe(locals);
	const [tablero, estado] = await Promise.all([
		d.tablero(periodo),
		// Con la API, la tarjeta de Integraciones muestra lo que de verdad está conectado (el tablero
		// de Go no lo trae). Si falla, la tarjeta invita a conectar en vez de tirar el Resumen.
		usaSandbox() ? null : d.integraciones().catch(() => null)
	]);
	const conectadas = estado
		? INTEGRACIONES.filter((i) => estado.conectadas[i.id]).map((i) => ({
				id: i.id,
				marca: i.marca,
				nombre: i.nombre
			}))
		: null;
	return { tablero, conectadas };
};

export const actions: Actions = {
	/** "Editar" la meta del mes. Acepta "19.600.000", "19600000" o "$ 19.600.000". */
	meta: async ({ locals, request }) => {
		const texto = String((await request.formData()).get('objetivo') ?? '');
		// Sólo pesos: dígitos con puntos de miles y, si quiere, el signo "$".
		const digitos = texto.replace(/[\s$.]/g, '');
		const objetivo = Number(digitos);
		if (!/^\d{1,12}$/.test(digitos) || objetivo > 100_000_000_000)
			return fail(400, { meta: 'Escribí la meta en pesos, por ejemplo 20.000.000.' });
		await datosDe(locals).guardarMeta(objetivo);
		return { meta: null };
	},
	ocultarCertificado: async ({ locals }) => {
		await datosDe(locals).ocultarAviso('certificado');
	},
	recordarCarritos: async ({ locals }) => {
		return { enviados: await datosDe(locals).recordarCarritos() };
	}
};
