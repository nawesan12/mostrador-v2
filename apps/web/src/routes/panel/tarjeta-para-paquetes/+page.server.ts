import { datosDe } from '#lib/server/datos/index.ts';
import { contextoTarjeta } from '#lib/server/tarjeta.ts';
import type { PageServerLoad } from './$types';

/** La tarjeta A6 para los paquetes, con los cupones activos para el QR del dorso. */
export const load: PageServerLoad = async ({ locals, url }) => {
	const datos = datosDe(locals);
	const [ctx, cupones] = await Promise.all([contextoTarjeta(datos, url.origin), datos.cupones()]);
	return {
		...ctx,
		cupones: cupones.cupones.filter((c) => c.estado === 'activo').map((c) => c.codigo)
	};
};
