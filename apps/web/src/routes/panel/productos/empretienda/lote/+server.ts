import { error, json } from '@sveltejs/kit';
import * as v from 'valibot';
import { datosDe } from '#lib/server/datos/index.ts';
import { leer, origenDe, productoDeFicha } from '#lib/server/empretienda.ts';
import type { RequestHandler } from './$types';

/**
 * Paso 2: de a 5 fichas por pedido (cada una son hasta 3 subpedidos y un Worker tiene 50). Las que ya
 * están, por nombre, no se repiten: importar dos veces no duplica.
 */
export const POST: RequestHandler = async ({ locals, request }) => {
	const b = v.safeParse(
		v.object({
			origen: v.pipe(v.string(), v.maxLength(200)),
			fichas: v.pipe(v.array(v.pipe(v.string(), v.maxLength(500))), v.minLength(1), v.maxLength(5))
		}),
		await request.json().catch(() => null)
	);
	const origen = b.success ? origenDe(b.output.origen) : null;
	if (!b.success || !origen) error(400, 'Pedido inválido.');
	// Sólo páginas de esa tienda: el servidor no sale a buscar cualquier URL.
	const fichas = b.output.fichas.filter((f) => f.startsWith(origen + '/'));
	const datos = datosDe(locals);
	let creados = 0,
		repetidos = 0,
		conVariantes = 0;
	const fallidas: string[] = [];
	await Promise.all(
		fichas.map(async (url) => {
			const html = await leer(url);
			const p = html ? productoDeFicha(html, url) : null;
			if (!p) return void fallidas.push(url);
			const ya = await datos.listar('productos', { q: p.nombre, cuantos: 5 });
			if (ya.filas.some((f) => String(f.nombre ?? '').toLowerCase() === p.nombre.toLowerCase()))
				return void repetidos++;
			if (p.descripcion?.includes('Variantes:')) conVariantes++;
			if (await datos.guardarProducto(null, p)) creados++;
			else fallidas.push(url);
		})
	);
	return json({ creados, repetidos, conVariantes, fallidas });
};
