import { error, json } from '@sveltejs/kit';
import * as v from 'valibot';
import { descubrir, FUENTES } from '#lib/server/importar/fuentes.ts';
import type { RequestHandler } from './$types';

/** Lo que se le sugiere pegar en cada plataforma, para los mensajes de error. */
const EJEMPLO = {
	shopify: 'mitienda.myshopify.com',
	woocommerce: 'mitienda.com.ar',
	empretienda: 'mitienda.empretienda.com.ar'
} as const;

/**
 * Importar tienda, paso 1: del link, la plataforma (la que delata el dominio, la elegida o la que
 * responda) y cuántas páginas hay. Empretienda devuelve además las fichas de su sitemap.
 */
export const POST: RequestHandler = async ({ request }) => {
	const b = v.safeParse(
		v.object({
			fuente: v.picklist(FUENTES),
			tienda: v.pipe(v.string(), v.maxLength(200))
		}),
		await request.json().catch(() => null)
	);
	if (!b.success) error(400, 'Pedido inválido.');
	const r = await descubrir(b.output.fuente, b.output.tienda);
	if (r === 'invalida')
		error(400, `Pegá el link de tu tienda, por\u00a0ejemplo\u00a0${EJEMPLO[b.output.fuente]}.`);
	if (r === 'no-encontrada')
		error(404, 'No encontramos esa tienda. Revisá el link y que esté activa.');
	if (r === 'vacia') error(404, 'La tienda no tiene productos publicados.');
	return json(r);
};
