import { equipoDe, exigir } from '#lib/server/staff/index.ts';
import type { PageServerLoad } from './$types';

/** La bandeja: contacto y arrepentimientos, con su estado y a quién están asignadas. */
export const load: PageServerLoad = async (event) => {
	const { sesion } = await event.parent();
	exigir(sesion.permisos, 'consultas.ver');
	return { bandeja: await equipoDe(event).consultas() };
};
