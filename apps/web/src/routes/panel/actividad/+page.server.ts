import { filtroDe } from '#lib/panel/actividad.ts';
import { datosDe } from '#lib/server/datos/index.ts';
import type { PageServerLoad } from './$types';

/** Actividad (§39): la primera página con los filtros de la URL. Go responde 403 a ventas y local. */
export const load: PageServerLoad = async ({ locals, url }) => {
	const filtro = filtroDe(url.searchParams);
	delete filtro.cursor;
	return { filtro, pagina: await datosDe(locals).actividad(filtro) };
};
