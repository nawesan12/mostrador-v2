import { error, json } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** Duplicar: una copia pausada, sin stock ni SKU, para cargar un producto parecido. */
export const POST: RequestHandler = async ({ locals, params }) => {
	const p = await datosDe(locals).duplicarProducto(params.id);
	if (!p) error(404, 'Ese producto no existe.');
	return json(p, { status: 201 });
};
