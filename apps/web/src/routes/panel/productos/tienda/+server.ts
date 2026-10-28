import { error, json } from '@sveltejs/kit';
import * as v from 'valibot';
import { fichasDelSitemap, leer, origenDe } from '#lib/server/empretienda.ts';
import type { RequestHandler } from './$types';

/** Hasta cuántas fichas se importan de una tienda. */
const MAX = 500;

/** Paso 1: del link de la tienda, la lista de fichas (sale del sitemap). */
export const POST: RequestHandler = async ({ request }) => {
	const b = v.safeParse(
		v.object({ tienda: v.pipe(v.string(), v.maxLength(200)) }),
		await request.json().catch(() => null)
	);
	const origen = b.success ? origenDe(b.output.tienda) : null;
	if (!origen)
		error(400, 'Pegá el link de tu tienda, por\u00a0ejemplo\u00a0mitienda.empretienda.com.ar.');
	const xml = await leer(`${origen}/sitemap.xml`);
	if (!xml?.includes('<urlset'))
		error(404, 'No encontramos esa tienda. Revisá el link y que esté activa.');
	const fichas = fichasDelSitemap(xml, origen);
	if (!fichas.length) error(404, 'La tienda no tiene productos publicados.');
	return json({ origen, fichas: fichas.slice(0, MAX), omitidas: Math.max(0, fichas.length - MAX) });
};
