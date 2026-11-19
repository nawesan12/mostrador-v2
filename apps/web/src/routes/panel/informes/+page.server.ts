import { isHttpError } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { PageServerLoad } from './$types';

/**
 * Informes es del plan Crece (con el newsletter): sin el plan, la pantalla se ve igual con la franja
 * que dice desde qué plan está (no un error).
 */
export const load: PageServerLoad = async ({ locals }) => {
	try {
		return { informes: await datosDe(locals).informes(), sinPlan: false as const };
	} catch (e) {
		if (isHttpError(e) && e.status === 402 && e.body.code === 'tope_plan')
			return { informes: null, sinPlan: true as const };
		throw e;
	}
};
