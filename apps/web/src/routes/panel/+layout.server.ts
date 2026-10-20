import { datosDe } from '#lib/server/datos/index.ts';
import type { LayoutServerLoad } from './$types';

// El panel sale de la fuente de datos de la sesión (sandbox de la demo o API Go).
export const load: LayoutServerLoad = async ({ locals }) => ({
	tienda: await datosDe(locals).tienda()
});
