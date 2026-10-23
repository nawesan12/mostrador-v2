import { datosDe, usaSandbox } from '#lib/server/datos/index.ts';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals }) => {
	const d = datosDe(locals);
	const [estado, mp] = await Promise.all([d.integraciones(), d.mercadoPago()]);
	return { estado, mp, demo: usaSandbox() };
};
