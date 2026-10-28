import { error } from '@sveltejs/kit';
import { resolverTienda } from '#lib/server/tienda/publica.ts';
import { COOKIE } from '#lib/server/sandbox/sandbox.ts';
import type { LayoutServerLoad } from './$types';

/** La tienda online de /t/[tienda] (la del visitante en la demo, o la vitrina de la plantilla). */
export const load: LayoutServerLoad = async ({ params, cookies, setHeaders, url, request }) => {
	const r = await resolverTienda(params.tienda, cookies.get(COOKIE));
	if (!r) error(404, 'Esta tienda no existe o todavía no está publicada.');
	// La vitrina se cachea un minuto; la tienda del visitante, los turnos y su confirmación, nunca.
	const fija = r.tienda.vitrina && !/\/turnos?(\/|$)/.test(url.pathname);
	setHeaders({
		'cache-control': fija ? 'public, max-age=60' : 'private, no-store',
		vary: 'User-Agent'
	});
	// Las plantillas eligen la cabecera y la grilla por el ancho: si el servidor ya sabe que es un
	// celular, el primer render sale como el del celular y no salta al hidratar (CLS).
	const movil = /Mobi|Android|iPhone|iPad/i.test(request.headers.get('user-agent') ?? '');
	return { tienda: r.tienda, movil };
};
