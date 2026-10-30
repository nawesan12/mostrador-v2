import { env } from 'cloudflare:workers';
import { fail, redirect } from '@sveltejs/kit';
import { dev } from '$app/env';
import { plantilla as buscarPlantilla } from '#lib/plantillas.ts';
import { RUBROS } from '#lib/server/sandbox/rubros.ts';
import { COOKIE, crear, dbSandbox, VIDA_MS } from '#lib/server/sandbox/sandbox.ts';
import { error } from '@sveltejs/kit';
import { usaSandbox } from '#lib/server/datos/index.ts';
import type { Actions, PageServerLoad } from './$types';

/** Sólo para los tests y el desarrollo local (motor de prueba): en producción no existe. */
export const load: PageServerLoad = async ({ locals }) => {
	if (!usaSandbox()) error(404, 'No existe.');
	return {
		// La plantilla de la demo actual, para marcarla "Elegida".
		actual: locals.sandbox
			? (locals.sandbox.plantilla ?? RUBROS.find((r) => r.id === locals.sandbox!.rubro)?.plantilla)
			: null
	};
};

export const actions: Actions = {
	default: async ({ request, cookies }) => {
		if (!usaSandbox()) error(404, 'No existe.');
		const f = await request.formData();
		const id = String(f.get('plantilla') ?? '');
		const p = buscarPlantilla(id);
		const rubro = p && RUBROS.find((r) => r.nombre === p.rubro);
		const sb = rubro && (await crear(dbSandbox(env.SANDBOX), rubro.id, Date.now(), p.id));
		if (!sb) return fail(400, { error: 'Elegí una de las plantillas de la galería.' });
		cookies.set(COOKIE, sb.id, {
			path: '/',
			httpOnly: true,
			secure: !dev,
			sameSite: 'lax',
			maxAge: VIDA_MS / 1000
		});
		// Al conectar una IA sin demo, se elige la tienda y se vuelve al consentimiento.
		const volver = String(f.get('volver') ?? '');
		redirect(303, volver.startsWith('/oauth/autorizar?') ? volver : '/panel');
	}
};
