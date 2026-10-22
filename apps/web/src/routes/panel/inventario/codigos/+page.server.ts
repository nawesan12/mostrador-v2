import { datosDe } from '#lib/server/datos/index.ts';
import type { PageServerLoad } from './$types';

/** Etiquetas con código de barras para imprimir (SKU, nombre y precio). */
export const load: PageServerLoad = async ({ locals, url }) => {
	const ids = new Set((url.searchParams.get('ids') ?? '').split(',').filter(Boolean));
	const l = await datosDe(locals).listar('productos', { cuantos: 200 });
	const productos = (
		l.filas as { id: string; nombre: string; sku: string; precio: number }[]
	).filter((p) => !ids.size || ids.has(p.id));
	return { productos };
};
