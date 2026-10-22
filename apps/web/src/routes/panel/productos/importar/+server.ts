import { error, json } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

const MAX = 200;
const sinAcento = (s: string) => s.normalize('NFD').replace(/[̀-ͯ]/g, '').toLowerCase().trim();
const numero = (s: string | undefined) => {
	const n = Number(
		(s ?? '')
			.replace(/[^\d,.-]/g, '')
			.replace(/\.(?=\d{3}(\D|$))/g, '')
			.replace(',', '.')
	);
	return Number.isFinite(n) && n >= 0 ? Math.round(n) : 0;
};

/**
 * "Importar Excel": un CSV (Excel → Guardar como CSV) con columnas nombre, precio, stock, categoría,
 * SKU y costo (en cualquier orden; separado por ; o ,). Hasta 200 productos por vez.
 */
export const POST: RequestHandler = async ({ locals, request }) => {
	const texto = (await request.text()).replace(new RegExp(`^${String.fromCharCode(0xfeff)}`), '');
	if (texto.length > 512 * 1024) error(413, 'El archivo es muy grande: hasta 512 KB.');
	const lineas = texto.split(/\r?\n/).filter((l) => l.trim());
	if (lineas.length < 2) error(400, 'El archivo no tiene productos.');
	const sep =
		(lineas[0].match(/;/g)?.length ?? 0) >= (lineas[0].match(/,/g)?.length ?? 0) ? ';' : ',';
	const partir = (l: string) => {
		const celdas: string[] = [];
		let actual = '';
		let comillas = false;
		for (const ch of l) {
			if (ch === '"') comillas = !comillas;
			else if (ch === sep && !comillas) {
				celdas.push(actual);
				actual = '';
			} else actual += ch;
		}
		return [...celdas, actual].map((c) => c.trim());
	};
	const cab = partir(lineas[0]).map(sinAcento);
	const col = (...nombres: string[]) => cab.findIndex((c) => nombres.some((n) => c.startsWith(n)));
	const iNombre = col('nombre', 'producto', 'titulo');
	if (iNombre < 0) error(400, 'Falta la columna "nombre".');
	const [iPrecio, iStock, iCat, iSku, iCosto] = [
		col('precio'),
		col('stock', 'cantidad'),
		col('categoria'),
		col('sku', 'codigo'),
		col('costo')
	];
	const datos = datosDe(locals);
	let creados = 0;
	const errores: string[] = [];
	for (const [n, l] of lineas.slice(1, MAX + 1).entries()) {
		const c = partir(l);
		const nombre = (c[iNombre] ?? '').slice(0, 120);
		if (nombre.length < 2) {
			errores.push(`Fila ${n + 2}: sin nombre`);
			continue;
		}
		await datos.guardarProducto(null, {
			nombre,
			precio: numero(c[iPrecio]),
			stock: iStock >= 0 ? numero(c[iStock]) : 0,
			categoria: iCat >= 0 ? c[iCat]?.slice(0, 60) || null : null,
			sku: iSku >= 0 ? c[iSku]?.slice(0, 40) : undefined,
			costo: iCosto >= 0 ? numero(c[iCosto]) : undefined,
			canales: { web: true }
		});
		creados++;
	}
	return json({ creados, errores, omitidos: Math.max(0, lineas.length - 1 - MAX) });
};
