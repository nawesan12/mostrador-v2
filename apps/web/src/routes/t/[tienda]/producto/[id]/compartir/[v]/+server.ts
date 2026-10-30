import { error } from '@sveltejs/kit';
import { responderImagen } from '#lib/server/compartir/index.ts';
import { COOKIE } from '#lib/server/sandbox/sandbox.ts';
import { resolverTienda } from '#lib/server/tienda/publica.ts';
import { tarjetaProducto, versionDe } from '#lib/tienda/compartir.ts';
import type { RequestHandler } from './$types';

/** La imagen al compartir un producto: su foto, nombre, precio y la tienda. */
export const GET: RequestHandler = async ({ params, cookies }) => {
	const v = params.v.match(/^([a-z0-9]{1,13})\.jpg$/)?.[1];
	if (!v) error(404, 'No existe.');
	const r = await resolverTienda(params.tienda, cookies.get(COOKIE));
	const p = r?.tienda.productos.find((x) => x.slug === params.id || x.id === params.id);
	if (!r || !p) error(404, 'Este producto no existe o ya no está publicado.');
	const tarjeta = tarjetaProducto(r.tienda, p);
	return responderImagen(
		tarjeta,
		`p-${r.tienda.slug}`,
		v === versionDe(tarjeta) ? 'public, max-age=31536000, immutable' : 'public, max-age=300'
	);
};
