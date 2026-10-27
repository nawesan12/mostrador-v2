import { error } from '@sveltejs/kit';
import { COOKIE } from '#lib/server/sandbox/sandbox.ts';
import { feedXml } from '#lib/server/tienda/feed.ts';
import { resolverTienda } from '#lib/server/tienda/publica.ts';
import type { RequestHandler } from './$types';

/**
 * El feed de productos para Google Merchant Center y Meta Commerce Manager. Lo piden una o dos veces
 * por día: con una hora de caché alcanza y no se gasta de más. En la demo, quien tiene su sandbox ve
 * el de su tienda (sin caché compartida); los demás, el de la vitrina.
 */
export const GET: RequestHandler = async ({ params, cookies, url }) => {
	const r = await resolverTienda(params.tienda, cookies.get(COOKIE));
	if (!r) error(404, 'Esta tienda no existe o todavía no está publicada.');
	const propia = !!r.sandboxId;
	return new Response(feedXml(r.tienda, url.origin), {
		headers: {
			'content-type': 'application/xml; charset=utf-8',
			'cache-control': propia ? 'private, no-store' : 'public, max-age=3600',
			// Con o sin la cookie de la demo, el feed es otro.
			vary: 'Cookie'
		}
	});
};
