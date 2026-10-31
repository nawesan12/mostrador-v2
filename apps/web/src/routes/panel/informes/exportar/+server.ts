import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

const BOM = String.fromCharCode(0xfeff);
const SERIES = {
	ventas: ['Ventas', 'informe-ventas'],
	productos: ['Unidades vendidas', 'informe-unidades'],
	clientes: ['Clientes nuevos', 'informe-clientes-nuevos'],
	margen: ['Margen bruto %', 'informe-margen'],
	visitas: ['Visitas a la tienda online', 'informe-visitas']
} as const;
type Serie = keyof typeof SERIES;

/** La pestaña que se está mirando (Ventas, Productos, Clientes, Margen o Visitas), 12 meses en CSV para Excel. */
export const GET: RequestHandler = async ({ locals, url }) => {
	const pedida = url.searchParams.get('serie') ?? 'ventas';
	const serie: Serie = pedida in SERIES ? (pedida as Serie) : 'ventas';
	const [titulo, archivo] = SERIES[serie];
	const i = await datosDe(locals).informes();
	const filas = [['Mes', titulo], ...i.meses.map((m, k) => [m, i.series[serie]?.[k] ?? ''])];
	const csv = BOM + filas.map((f) => f.join(';')).join('\r\n');
	return new Response(csv, {
		headers: {
			'content-type': 'text/csv; charset=utf-8',
			'content-disposition': `attachment; filename="${archivo}-12-meses.csv"`,
			'cache-control': 'no-store'
		}
	});
};
