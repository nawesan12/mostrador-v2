import { error } from '@sveltejs/kit';
import { enCastellano } from '#lib/panel/textos-export.ts';
import { datosDe } from '#lib/server/datos/index.ts';
import type { TipoExport } from '#lib/server/datos/tipos.ts';
import { TIPO_XLSX, xlsx } from '#lib/server/xlsx.ts';
import { csv, zip } from '#lib/server/zip.ts';
import type { RequestHandler } from './$types';

const TIPOS: TipoExport[] = [
	'productos',
	'clientes',
	'pedidos',
	'facturas',
	'resenas',
	'conversaciones'
];
const PESTANA: Record<TipoExport, string> = {
	productos: 'Productos',
	clientes: 'Clientes',
	pedidos: 'Pedidos',
	facturas: 'Facturas',
	resenas: 'Reseñas',
	conversaciones: 'Conversaciones'
};

/**
 * Exportar en CSV o en Excel (.xlsx de verdad); "todo" junta los seis CSV en un ZIP. Los códigos
 * internos (canal, pago, envío, estado) salen en castellano. Cada descarga queda en
 * "Exportaciones recientes".
 */
export const GET: RequestHandler = async ({ locals, params, url }) => {
	const datos = datosDe(locals);
	const formato = url.searchParams.get('formato') === 'excel' ? 'excel' : 'csv';
	const dia = new Date().toISOString().slice(0, 10);
	// Si no se pudo anotar, el archivo igual baja.
	const registrar = (tipo: TipoExport | 'todo', f: 'csv' | 'excel') =>
		datos.registrarExportacion(tipo, f).catch(() => {});
	if (params.tipo === 'todo') {
		const enc = new TextEncoder();
		const tablas = await Promise.all(TIPOS.map((t) => datos.exportar(t)));
		const archivos = TIPOS.map((t, i) => ({
			nombre: `${t}.csv`,
			datos: enc.encode(csv(enCastellano(tablas[i]), 'csv'))
		}));
		await registrar('todo', 'csv');
		return new Response(zip(archivos).buffer as ArrayBuffer, {
			headers: {
				'content-type': 'application/zip',
				'content-disposition': `attachment; filename="tulocalweb-${dia}.zip"`,
				'cache-control': 'no-store'
			}
		});
	}
	const tipo = TIPOS.find((t) => t === params.tipo);
	if (!tipo) error(404, 'No se exporta eso.');
	const filas = enCastellano(await datos.exportar(tipo));
	await registrar(tipo, formato);
	if (formato === 'excel')
		return new Response(xlsx(filas, PESTANA[tipo]).buffer as ArrayBuffer, {
			headers: {
				'content-type': TIPO_XLSX,
				'content-disposition': `attachment; filename="${tipo}-${dia}.xlsx"`,
				'cache-control': 'no-store'
			}
		});
	return new Response(csv(filas, 'csv'), {
		headers: {
			'content-type': 'text/csv; charset=utf-8',
			'content-disposition': `attachment; filename="${tipo}-${dia}.csv"`,
			'cache-control': 'no-store'
		}
	});
};
