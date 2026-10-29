import { redirect } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/**
 * Importar desde Tiendanube, ida: arma el state de un solo uso (en la demo, en el sandbox; con la
 * API, en Go) y manda a autorizar la app de socio. Sin la app configurada vuelve a Productos con el
 * aviso. Sólo se redirige a Tiendanube: cualquier otra URL es un error.
 */
export const GET: RequestHandler = async ({ locals }) => {
	const url = await datosDe(locals).conectarTiendanube();
	if (!url) redirect(303, '/panel/productos?tiendanube=no-disponible');
	if (!url.startsWith('https://www.tiendanube.com/apps/'))
		redirect(303, '/panel/productos?tiendanube=error');
	redirect(303, url, { external: ['https://www.tiendanube.com'] });
};
