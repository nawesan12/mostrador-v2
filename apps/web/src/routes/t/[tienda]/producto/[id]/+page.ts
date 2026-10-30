import { error, redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ parent, params, url }) => {
	const { tienda } = await parent();
	// Por el slug (/producto/mate-de-calabaza) o por el id de antes, que lleva al slug con un 301.
	const producto = tienda.productos.find((p) => p.slug === params.id || p.id === params.id);
	if (!producto) error(404, 'Este producto no existe o ya no está publicado.');
	if (producto.slug && producto.slug !== params.id)
		redirect(301, url.pathname.replace(/[^/]+$/, producto.slug) + url.search);
	return { producto };
};
