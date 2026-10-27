import { env } from 'cloudflare:workers';
import { error } from '@sveltejs/kit';
import { CLAVE } from '#lib/server/archivos.ts';
import type { RequestHandler } from './$types';

/** Las fotos de productos (R2). La clave lleva un uuid: no cambia nunca, se cachea un año. */
export const GET: RequestHandler = async ({ params, request }) => {
	if (!CLAVE.test(params.clave)) error(404, 'No existe.');
	const o = await env.FILES.get(params.clave, { onlyIf: request.headers });
	if (!o) error(404, 'No existe.');
	const h = new Headers();
	o.writeHttpMetadata(h);
	h.set('etag', o.httpEtag);
	h.set('cache-control', 'public, max-age=31536000, immutable');
	h.set('x-content-type-options', 'nosniff');
	// Sin cuerpo: se cumplió el If-None-Match.
	if (!('body' in o)) return new Response(null, { status: 304, headers: h });
	return new Response(o.body, { headers: h });
};
