import { error } from '@sveltejs/kit';
import { categoria, delaCategoria } from '#lib/ayuda/articulos.ts';
import type { PageLoad } from './$types';

export const load: PageLoad = ({ params }) => {
	const c = categoria(params.categoria);
	if (!c) error(404, 'Esa sección de la ayuda no existe.');
	return { slug: c.slug, articulos: delaCategoria(c.slug).map((a) => a.slug) };
};
