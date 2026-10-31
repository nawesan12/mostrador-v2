import { datosDe } from '#lib/server/datos/index.ts';
import { leer } from '#lib/panel/listas/filtros.ts';
import type { PageServerLoad } from './$types';

/** La primera ventana de la lista con los filtros de la URL, y ARCA (qué letra se emite). */
export const load: PageServerLoad = async ({ locals, url, depends }) => {
	// Tiempo real (§44): se recarga con los avisos de lo que muestra.
	depends('tienda:facturacion');
	const d = datosDe(locals);
	const [lista, arca] = await Promise.all([
		d.listar('facturacion', { ...leer(url.searchParams), desde: 0, cuantos: 60 }),
		d.arca()
	]);
	return {
		lista,
		arca: { conectada: arca.conectada, condicion: arca.condicion }
	};
};
