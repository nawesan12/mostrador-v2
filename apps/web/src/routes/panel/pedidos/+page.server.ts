import { datosDe } from '#lib/server/datos/index.ts';
import { leerFiltros } from '#lib/panel/pedidos/filtros.ts';
import type { PageServerLoad } from './$types';

/** Pedidos: la primera ventana de la lista con los filtros de la URL. */
export const load: PageServerLoad = async ({ locals, url }) => {
	const f = leerFiltros(url.searchParams);
	return { lista: await datosDe(locals).listarPedidos({ ...f, desde: 0, cuantos: 60 }) };
};
