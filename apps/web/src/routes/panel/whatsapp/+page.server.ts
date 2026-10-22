import { datosDe } from '#lib/server/datos/index.ts';
import type { PageServerLoad } from './$types';

/** La bandeja y la conversación abierta (?c=id; si no, la primera). */
export const load: PageServerLoad = async ({ locals, url }) => {
	const datos = datosDe(locals);
	const b = await datos.bandeja();
	const id = url.searchParams.get('c') ?? b.conversaciones[0]?.id ?? null;
	return {
		...b,
		abierta: id ? await datos.conversacion(id) : null,
		elegida: url.searchParams.has('c')
	};
};
