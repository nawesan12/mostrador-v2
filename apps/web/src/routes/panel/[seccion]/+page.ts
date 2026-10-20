import { error } from '@sveltejs/kit';
import { SECCIONES, slug } from '#lib/panel/navegacion.ts';
import type { PageLoad } from './$types';

export const load: PageLoad = ({ params }) => {
	const seccion = SECCIONES.find((s) => slug(s) === params.seccion);
	if (!seccion) error(404, 'Esa sección del panel no existe.');
	return { seccion };
};
