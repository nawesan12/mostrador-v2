import { COOKIE } from '#lib/server/sandbox/sandbox.ts';
import { resolverTienda } from '#lib/server/tienda/publica.ts';
import { proximosDe } from '#lib/server/tienda/turnos.ts';
import type { PageServerLoad } from './$types';

/** La portada: si la tienda toma turnos, los libres de hoy y el próximo de cada profesional. */
export const load: PageServerLoad = async ({ parent, params, cookies }) => {
	const { tienda } = await parent();
	if (!tienda.turnos) return { proximos: null };
	const r = await resolverTienda(params.tienda, cookies.get(COOKIE));
	return { proximos: r ? await proximosDe(r) : null };
};
