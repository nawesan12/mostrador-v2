import { error, json } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { AccionLote } from '#lib/server/datos/tipos.ts';
import type { RequestHandler } from './$types';

const ACCIONES: AccionLote[] = ['preparar', 'etiquetas', 'facturar', 'avisar'];

/** La barra de acciones sobre los pedidos seleccionados. */
export const POST: RequestHandler = async ({ locals, request }) => {
	const c = (await request.json().catch(() => null)) as {
		numeros?: unknown;
		accion?: unknown;
	} | null;
	const numeros = Array.isArray(c?.numeros)
		? c.numeros.filter((n): n is number => Number.isSafeInteger(n))
		: [];
	if (!numeros.length || numeros.length > 500 || !ACCIONES.includes(c?.accion as AccionLote))
		error(400, 'Elegí pedidos y una acción.');
	return json(await datosDe(locals).pedidosEnLote(numeros, c!.accion as AccionLote));
};
