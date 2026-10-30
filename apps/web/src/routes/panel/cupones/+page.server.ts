import { datosDe } from '#lib/server/datos/index.ts';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals, parent }) => ({
	...(await datosDe(locals).cupones()),
	// El ticket promedio real de los últimos 30 días, para lo que calcula Tincho (0: sin ventas).
	ticket: (await datosDe(locals).tablero('30d')).kpis.ticket.actual,
	// Para el link que aplica el cupón solo (/t/{slug}/c/{codigo}).
	slug: (await parent()).tienda.slug
});
