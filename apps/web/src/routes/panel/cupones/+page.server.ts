import { datosDe } from '#lib/server/datos/index.ts';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals, parent }) => ({
	...(await datosDe(locals).cupones()),
	// Para el link que aplica el cupón solo (/t/{slug}/c/{codigo}).
	slug: (await parent()).tienda.slug
});
