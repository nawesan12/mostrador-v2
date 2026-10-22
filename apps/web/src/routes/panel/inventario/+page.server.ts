import { datosDe } from '#lib/server/datos/index.ts';
import { leer } from '#lib/panel/listas/filtros.ts';
import type { PageServerLoad } from './$types';

/** Inventario: la primera ventana, y los movimientos de hoy. */
export const load: PageServerLoad = async ({ locals, url }) => {
	const datos = datosDe(locals);
	const [lista, movimientos] = await Promise.all([
		datos.listar('inventario', { ...leer(url.searchParams), desde: 0, cuantos: 60 }),
		datos.movimientosHoy()
	]);
	return { lista, movimientos };
};
