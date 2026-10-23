import { error } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { TipoExport } from '#lib/server/datos/tipos.ts';
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

/** Exportar en CSV o Excel; "todo" junta los seis CSV en un ZIP. */
export const GET: RequestHandler = async ({ locals, params, url }) => {
	const datos = datosDe(locals);
	const formato = url.searchParams.get('formato') === 'excel' ? 'excel' : 'csv';
	const dia = new Date().toISOString().slice(0, 10);
	if (params.tipo === 'todo') {
		const enc = new TextEncoder();
		const archivos = [];
		for (const t of TIPOS)
			archivos.push({
				nombre: `${t}.csv`,
				datos: enc.encode(csv(await datos.exportar(t), formato))
			});
		return new Response(zip(archivos).buffer as ArrayBuffer, {
			headers: {
				'content-type': 'application/zip',
				'content-disposition': `attachment; filename="mostrador-${dia}.zip"`
			}
		});
	}
	const tipo = TIPOS.find((t) => t === params.tipo);
	if (!tipo) error(404, 'No se exporta eso.');
	return new Response(csv(await datos.exportar(tipo), formato), {
		headers: {
			'content-type': 'text/csv; charset=utf-8',
			'content-disposition': `attachment; filename="${tipo}-${dia}${formato === 'excel' ? '-excel' : ''}.csv"`
		}
	});
};
