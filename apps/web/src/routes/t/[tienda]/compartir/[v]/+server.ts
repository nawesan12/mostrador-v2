import { error } from '@sveltejs/kit';
import { responderImagen } from '#lib/server/compartir/index.ts';
import { COOKIE } from '#lib/server/sandbox/sandbox.ts';
import { resolverTienda } from '#lib/server/tienda/publica.ts';
import { tarjetaDe, versionDe } from '#lib/tienda/compartir.ts';
import type { RequestHandler } from './$types';

/**
 * La imagen al compartir la tienda (og:image), 1200 × 630 en JPEG. [v] es la versión de lo
 * dibujado: con la de ahora se cachea un año; con una vieja se responde la de ahora por un rato.
 */
export const GET: RequestHandler = async ({ params, cookies }) => {
	const v = params.v.match(/^([a-z0-9]{1,13})\.jpg$/)?.[1];
	if (!v) error(404, 'No existe.');
	const r = await resolverTienda(params.tienda, cookies.get(COOKIE));
	if (!r) error(404, 'Esta tienda no existe o todavía no está publicada.');
	const tarjeta = tarjetaDe(r.tienda);
	return responderImagen(
		tarjeta,
		`t-${r.tienda.slug}`,
		v === versionDe(tarjeta) ? 'public, max-age=31536000, immutable' : 'public, max-age=300'
	);
};
