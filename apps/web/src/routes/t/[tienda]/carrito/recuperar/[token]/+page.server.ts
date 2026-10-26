import { clienteDe } from '#lib/server/datos/index.ts';
import { COOKIE } from '#lib/server/sandbox/sandbox.ts';
import { carritoPorToken } from '#lib/server/tienda/carritos.ts';
import { resolverTienda } from '#lib/server/tienda/publica.ts';
import type { PageServerLoad } from './$types';

/** El link del recordatorio: las líneas del carrito guardado (la página las pone en el navegador). */
export const load: PageServerLoad = async (event) => {
	const { params, cookies } = event;
	const r = await resolverTienda(params.tienda, cookies.get(COOKIE));
	const lineas = r ? await carritoPorToken(r, clienteDe(event), params.token) : null;
	return { lineas };
};
