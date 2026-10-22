import { error } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { NombreLista } from '#lib/server/datos/tipos.ts';
import { leer } from '#lib/panel/listas/filtros.ts';
import type { RequestHandler } from './$types';

/** Columnas de cada lista: título y campo del registro. */
const COLUMNAS: Record<NombreLista, [string, string][]> = {
	clientes: [
		['Nombre', 'nombre'],
		['Ciudad', 'ciudad'],
		['Teléfono', 'telefono'],
		['Pedidos', 'pedidos'],
		['Gastado', 'gastado'],
		['Días desde la última compra', 'ultima'],
		['Segmento', 'seg']
	],
	productos: [
		['Nombre', 'nombre'],
		['SKU', 'sku'],
		['Categoría', 'categoria'],
		['Precio', 'precio'],
		['Costo', 'costo'],
		['Stock', 'stock'],
		['Vendidos en 30 días', 'vendidos30'],
		['Estado', 'estado']
	],
	inventario: [
		['Producto', 'nombre'],
		['SKU', 'sku'],
		['Stock', 'stock'],
		['Mínimo', 'min'],
		['Ubicación', 'loc'],
		['Proveedor', 'proveedor']
	],
	facturacion: [
		['Comprobante', 'numero'],
		['Tipo', 'tipo'],
		['Cliente', 'cliente'],
		['Total', 'total'],
		['Estado', 'estado']
	],
	envios: [
		['Pedido', 'pedido'],
		['Cliente', 'cliente'],
		['Ciudad', 'ciudad'],
		['Transporte', 'transporte'],
		['Seguimiento', 'seguimiento'],
		['Estado', 'estado'],
		['Costo', 'costo']
	]
};

const celda = (v: unknown) => {
	const s = typeof v === 'number' ? String(Math.round(v)) : String(v ?? '');
	return /[",;\n\r]|^[=+\-@]/.test(s)
		? `"${s.replace(/"/g, '""').replace(/^([=+\-@])/, "'$1")}"`
		: s;
};

/** "Exportar": la lista con los filtros aplicados, en CSV para Excel en español. */
export const GET: RequestHandler = async ({ locals, params, url }) => {
	const lista = params.lista as NombreLista;
	const cols = COLUMNAS[lista];
	if (!cols) error(404, 'Esa lista no existe.');
	const datos = datosDe(locals);
	const f = leer(url.searchParams);
	const filas: Record<string, unknown>[] = [];
	for (let desde = 0; desde < 5000; desde += 200) {
		const l = await datos.listar(lista, { ...f, desde, cuantos: 200 });
		filas.push(...(l.filas as Record<string, unknown>[]));
		if (desde + 200 >= l.total) break;
	}
	const cuerpo = [
		cols.map((c) => c[0]).join(';'),
		...filas.map((r) => cols.map((c) => celda(r[c[1]])).join(';'))
	];
	return new Response('﻿' + cuerpo.join('\r\n'), {
		headers: {
			'content-type': 'text/csv; charset=utf-8',
			'content-disposition': `attachment; filename="${lista}-${new Date().toISOString().slice(0, 10)}.csv"`,
			'cache-control': 'no-store'
		}
	});
};
