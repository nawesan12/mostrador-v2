import { usaSandbox } from '#lib/server/datos/index.ts';
import { partnersDe } from '#lib/server/partners/index.ts';
import type { LayoutServerLoad } from './$types';

/** El panel de partners entero (es chico): lo piden todas las pantallas. */
export const load: LayoutServerLoad = async (event) => ({
	panel: await partnersDe(event).panel(),
	demo: usaSandbox()
});
