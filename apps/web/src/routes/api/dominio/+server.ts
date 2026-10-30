import { json } from '@sveltejs/kit';
import { slugDeDominio } from '#lib/server/tienda/dominio.ts';
import type { RequestHandler } from './$types';

/** La tienda del dominio desde el que se pide (la usa el hook `reroute`, también en el navegador). */
export const GET: RequestHandler = async ({ url, setHeaders }) => {
	setHeaders({ 'cache-control': 'public, max-age=300' });
	return json({ slug: await slugDeDominio(url.hostname) });
};
