import { datosDe } from '#lib/server/datos/index.ts';
import { leerFiltros } from '#lib/panel/pedidos/filtros.ts';
import type { RequestHandler } from './$types';

const MAX = 5000;
const celda = (v: unknown) => {
	const s = String(v ?? '');
	// Comillas para el separador y para que Excel no ejecute fórmulas (=, +, -, @).
	return /[",;\n\r]|^[=+\-@]/.test(s)
		? `"${s.replace(/"/g, '""').replace(/^([=+\-@])/, "'$1")}"`
		: s;
};

/** "Exportar": los pedidos con los filtros aplicados, en CSV (separado por ; para Excel en español). */
export const GET: RequestHandler = async ({ locals, url }) => {
	const datos = datosDe(locals);
	const f = leerFiltros(url.searchParams);
	const filas = [];
	for (let desde = 0; desde < MAX; desde += 200) {
		const l = await datos.listarPedidos({ ...f, desde, cuantos: 200 });
		filas.push(...l.filas);
		if (desde + 200 >= l.total) break;
	}
	const cab = [
		'Número',
		'Fecha',
		'Cliente',
		'Ciudad',
		'Canal',
		'Pago',
		'Envío',
		'Estado',
		'Productos',
		'Total'
	];
	const lineas = filas.map((p) =>
		[
			p.numero,
			new Date(p.creado).toLocaleString('es-AR', { timeZone: 'America/Argentina/Buenos_Aires' }),
			p.cliente,
			p.ciudad,
			p.canal,
			p.pago,
			p.envio,
			p.estado,
			p.productos.map((x) => x.nombre).join(' + '),
			p.total
		]
			.map(celda)
			.join(';')
	);
	const hoy = new Date().toISOString().slice(0, 10);
	return new Response('﻿' + [cab.join(';'), ...lineas].join('\r\n'), {
		headers: {
			'content-type': 'text/csv; charset=utf-8',
			'content-disposition': `attachment; filename="pedidos-${hoy}.csv"`,
			'cache-control': 'no-store'
		}
	});
};
