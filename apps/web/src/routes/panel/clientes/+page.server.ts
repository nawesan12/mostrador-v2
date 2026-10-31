import { datosDe } from '#lib/server/datos/index.ts';
import { leer } from '#lib/panel/listas/filtros.ts';
import type { PageServerLoad } from './$types';

/** La primera ventana de la lista con los filtros de la URL. */
export const load: PageServerLoad = async ({ locals, url, depends }) => {
	// Tiempo real (§44): se recarga con los avisos de lo que muestra.
	depends('tienda:clientes');
	return {
		lista: await datosDe(locals).listar('clientes', {
			...leer(url.searchParams),
			desde: 0,
			cuantos: 60
		})
	};
};
