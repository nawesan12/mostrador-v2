import { env } from 'cloudflare:workers';
import { error, redirect } from '@sveltejs/kit';
import { SITE } from '#lib/site.ts';
import { resolverTienda } from '#lib/server/tienda/publica.ts';
import { COOKIE } from '#lib/server/sandbox/sandbox.ts';
import type { LayoutServerLoad } from './$types';

/** La tienda online de /t/[tienda] (la del visitante en la demo, o la vitrina de la plantilla). */
export const load: LayoutServerLoad = async ({
	params,
	cookies,
	setHeaders,
	url,
	request,
	locals,
	depends
}) => {
	// Tiempo real (§44): el stock y las fichas se recargan con los avisos de la tienda.
	depends('tienda:vidriera');
	const r = await resolverTienda(params.tienda, cookies.get(COOKIE));
	if (!r) error(404, 'Esta tienda no existe o todavía no está publicada.');
	// Con dominio propio principal, /t/<slug>/… vive allá: 301 con el mismo camino (Google y los
	// links compartidos pasan al dominio). En el dominio, los links van sin /t/<slug>.
	const base = `/t/${r.tienda.slug}`;
	if (!locals.dominio && r.tienda.dominio && !r.tienda.vitrina && request.method === 'GET')
		redirect(
			301,
			`https://${r.tienda.dominio}${url.pathname.slice(base.length) || '/'}${url.search}`,
			{ external: [`https://${r.tienda.dominio}`] }
		);
	if (locals.dominio) r.tienda.base = '';
	// La vitrina se cachea un minuto; la tienda del visitante, los turnos y su confirmación, nunca.
	const fija = r.tienda.vitrina && !/\/turnos?(\/|$)/.test(url.pathname);
	setHeaders({
		'cache-control': fija ? 'public, max-age=60' : 'private, no-store',
		vary: 'User-Agent'
	});
	// Las plantillas eligen la cabecera y la grilla por el ancho: si el servidor ya sabe que es un
	// celular, el primer render sale como el del celular y no salta al hidratar (CLS).
	const movil = /Mobi|Android|iPhone|iPad/i.test(request.headers.get('user-agent') ?? '');
	// Tiempo real sólo con la API y en una tienda de verdad (no en la demo ni en las vitrinas). En un
	// dominio propio el WebSocket va a tulocalweb.com.ar (Caddy no lo rutea en cada dominio).
	const tiempoReal =
		r.datos || r.tienda.vitrina
			? null
			: { origen: String(env.TIEMPO_REAL_ORIGEN ?? '') || (locals.dominio ? SITE : null) };
	return { tienda: r.tienda, movil, tiempoReal };
};
