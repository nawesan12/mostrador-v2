import { filtroDe } from '#lib/panel/actividad.ts';
import { datosDe } from '#lib/server/datos/index.ts';
import { csv } from '#lib/server/zip.ts';
import type { RequestHandler } from './$types';

/**
 * La actividad en CSV con los filtros que se están viendo (con BOM y punto y coma: así la abre bien
 * Excel en castellano). La descarga queda anotada en la Actividad (lo hace Go).
 */
export const GET: RequestHandler = async ({ locals, url }) => {
	const filas = await datosDe(locals).exportarActividad(filtroDe(url.searchParams));
	const dia = new Date(Date.now() - 3 * 3_600_000).toISOString().slice(0, 10);
	return new Response(csv(filas, 'excel'), {
		headers: {
			'content-type': 'text/csv; charset=utf-8',
			'content-disposition': `attachment; filename="actividad-${dia}.csv"`,
			'cache-control': 'no-store'
		}
	});
};
