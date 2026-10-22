import { error } from '@sveltejs/kit';
import { ARTICULOS } from '#lib/ayuda/articulos.ts';
import type { PageLoad } from './$types';

export const load: PageLoad = ({ params }) => {
	const a = ARTICULOS.find((x) => x.slug === params.articulo && x.categoria === params.categoria);
	if (!a) error(404, 'Ese artículo no existe.');
	return { slug: a.slug };
};
