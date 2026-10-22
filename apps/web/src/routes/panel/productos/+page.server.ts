import { datosDe } from '#lib/server/datos/index.ts';
import { leer } from '#lib/panel/listas/filtros.ts';
import type { PageServerLoad } from './$types';

/** La primera ventana de la lista con los filtros de la URL. */
export const load: PageServerLoad = async ({ locals, url }) => ({
	lista: await datosDe(locals).listar('productos', {
		...leer(url.searchParams),
		desde: 0,
		cuantos: 60
	})
});
