import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

const BOM = String.fromCharCode(0xfeff);

/** Los 12 meses en CSV para Excel. */
export const GET: RequestHandler = async ({ locals }) => {
	const i = await datosDe(locals).informes();
	const filas = [
		['Mes', 'Ventas', 'Unidades', 'Clientes nuevos', 'Margen %'],
		...i.meses.map((m, k) => [
			m,
			i.series.ventas[k],
			i.series.productos[k],
			i.series.clientes[k],
			i.series.margen[k] ?? ''
		])
	];
	const csv = BOM + filas.map((f) => f.join(';')).join('\r\n');
	return new Response(csv, {
		headers: {
			'content-type': 'text/csv; charset=utf-8',
			'content-disposition': 'attachment; filename="informe-12-meses.csv"'
		}
	});
};
