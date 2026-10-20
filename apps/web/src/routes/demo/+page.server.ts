import { env } from 'cloudflare:workers';
import { fail, redirect } from '@sveltejs/kit';
import { dev } from '$app/env';
import { RUBROS } from '#lib/server/sandbox/rubros.ts';
import { COOKIE, crear, dbSandbox, VIDA_MS } from '#lib/server/sandbox/sandbox.ts';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals }) => ({
	rubros: RUBROS,
	actual: locals.sandbox?.rubro ?? null
});

export const actions: Actions = {
	default: async ({ request, cookies }) => {
		const rubro = String((await request.formData()).get('rubro') ?? '');
		const sb = await crear(dbSandbox(env.SANDBOX), rubro);
		if (!sb) return fail(400, { error: 'Elegí uno de los rubros de la lista.' });
		cookies.set(COOKIE, sb.id, {
			path: '/',
			httpOnly: true,
			secure: !dev,
			sameSite: 'lax',
			maxAge: VIDA_MS / 1000
		});
		redirect(303, '/panel');
	}
};
