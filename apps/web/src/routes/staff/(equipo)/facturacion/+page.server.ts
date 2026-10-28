import { equipoDe, exigir } from '#lib/server/staff/index.ts';
import type { PageServerLoad } from './$types';

/** Facturación de Mostrador: abonos, suscripciones con problemas, créditos y comprobantes. */
export const load: PageServerLoad = async (event) => {
	const { sesion } = await event.parent();
	exigir(sesion.permisos, 'facturacion.ver');
	return { facturacion: await equipoDe(event).facturacion() };
};
