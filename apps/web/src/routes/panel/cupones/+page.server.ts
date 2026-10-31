import { datosDe } from '#lib/server/datos/index.ts';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals, parent, depends }) => {
	// Tiempo real (§44): se recarga con los avisos de lo que muestra.
	depends('tienda:cupones');
	return {
		...(await datosDe(locals).cupones()),
		// El ticket promedio real de los últimos 30 días, para lo que calcula Tincho (0: sin ventas o
		// sin montos; Cupones es de quien tiene Catálogo, que ve montos).
		ticket: (await datosDe(locals).tablero('30d')).kpis.ticket?.actual ?? 0,
		// Para el link que aplica el cupón solo (/t/{slug}/c/{codigo}).
		slug: (await parent()).tienda.slug
	};
};
