import { datosDe } from '#lib/server/datos/index.ts';
import { leerFiltros } from '#lib/panel/pedidos/filtros.ts';
import type { PageServerLoad } from './$types';

/**
 * Pedidos: la primera ventana de la lista con los filtros de la URL, y dónde y cuándo se retira
 * (para el WhatsApp de «listo para retirar»).
 */
export const load: PageServerLoad = async ({ locals, url, depends }) => {
	// Tiempo real (§44): pedidos nuevos y cambios de estado o pago aparecen solos.
	depends('tienda:pedidos');
	const f = leerFiltros(url.searchParams);
	const datos = datosDe(locals);
	const [lista, entregas] = await Promise.all([
		datos.listarPedidos({ ...f, desde: 0, cuantos: 60 }),
		// Si no se puede leer, el mensaje dice «listo para retirar» sin la dirección.
		datos.entregas().catch(() => null)
	]);
	return {
		lista,
		retiro: {
			direccion: entregas?.retiroDireccion ?? '',
			horario: entregas?.retiroHorario ?? ''
		}
	};
};
