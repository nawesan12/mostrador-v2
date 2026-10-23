import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ parent, params }) => {
	const { tienda } = await parent();
	const producto = tienda.productos.find((p) => p.id === params.id);
	if (!producto) error(404, 'Este producto no existe o ya no está publicado.');
	return { producto };
};
