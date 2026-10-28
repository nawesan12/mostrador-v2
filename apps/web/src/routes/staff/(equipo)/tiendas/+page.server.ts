import { leer } from '#lib/panel/listas/filtros.ts';
import { equipoDe, exigir } from '#lib/server/staff/index.ts';
import type { PageServerLoad } from './$types';

/** La primera ventana de las tiendas con los filtros de la URL. */
export const load: PageServerLoad = async (event) => {
	const { sesion } = await event.parent();
	exigir(sesion.permisos, 'tiendas.ver');
	return {
		lista: await equipoDe(event).tiendas({ ...leer(event.url.searchParams), desde: 0, cuantos: 60 })
	};
};
