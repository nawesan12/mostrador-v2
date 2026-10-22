import { error, json } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { NombreLista } from '#lib/server/datos/tipos.ts';
import { leer } from '#lib/panel/listas/filtros.ts';
import type { RequestHandler } from './$types';

const LISTAS: NombreLista[] = ['clientes', 'productos', 'inventario', 'facturacion', 'envios'];
const listaDe = (s: string) => {
	if (!LISTAS.includes(s as NombreLista)) error(404, 'Esa lista no existe.');
	return s as NombreLista;
};

/** Otra ventana de la lista virtual. */
export const GET: RequestHandler = async ({ locals, params, url }) => {
	const f = leer(url.searchParams);
	const desde = Math.max(0, Number(url.searchParams.get('desde')) || 0);
	const cuantos = Math.min(200, Math.max(1, Number(url.searchParams.get('cuantos')) || 60));
	return json(await datosDe(locals).listar(listaDe(params.lista), { ...f, desde, cuantos }), {
		headers: { 'cache-control': 'no-store' }
	});
};

/** Una acción sobre filas (o la sugerencia de Hornero). */
export const POST: RequestHandler = async ({ locals, params, request }) => {
	const c = (await request.json().catch(() => null)) as {
		ids?: unknown;
		accion?: unknown;
		valor?: unknown;
	} | null;
	const ids = Array.isArray(c?.ids)
		? c.ids.filter((x): x is string => typeof x === 'string').slice(0, 500)
		: [];
	if (typeof c?.accion !== 'string' || !/^[a-z-]{2,40}$/.test(c.accion))
		error(400, 'Falta la acción.');
	const hechos = await datosDe(locals).accionLista(listaDe(params.lista), ids, c.accion, c.valor);
	return json({ hechos });
};
