import { redirect } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/**
 * Vuelta del OAuth de Tiendanube (la URL registrada en la app de socio:
 * https://mostrador.iamnaw.dev/panel/importar/tiendanube/vuelta). Valida el state, canjea el código y
 * crea la importación con la credencial; vuelve a Productos, donde el modal muestra el progreso. Un
 * error vuelve con su motivo (el modal lo traduce).
 */
export const GET: RequestHandler = async ({ locals, url }) => {
	const codigo = url.searchParams.get('code') ?? '';
	const estado = url.searchParams.get('state') ?? '';
	const r =
		url.searchParams.has('error') || !codigo || !estado
			? 'cancelada'
			: await datosDe(locals).volverTiendanube(codigo.slice(0, 512), estado.slice(0, 512));
	redirect(303, `/panel/productos?tiendanube=${typeof r === 'string' ? r : 'importando'}`);
};
