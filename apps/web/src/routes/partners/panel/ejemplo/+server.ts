import { env } from 'cloudflare:workers';
import { redirect } from '@sveltejs/kit';
import { dev } from '$app/env';
import { usaSandbox } from '#lib/server/datos/index.ts';
import { COOKIE, crear, dbSandbox, VIDA_MS } from '#lib/server/sandbox/sandbox.ts';
import { RUBROS } from '#lib/server/sandbox/rubros.ts';
import type { RequestHandler } from './$types';

/**
 * "Ver el panel de partners de ejemplo" (/partners): usa la demo del visitante o arma una (como
 * /demo, con el primer rubro) y lleva al panel. Con la API real no hay ejemplo: se ingresa.
 */
export const POST: RequestHandler = async ({ locals, cookies }) => {
	if (!usaSandbox()) redirect(303, '/ingresar');
	if (!locals.sandbox) {
		const sb = await crear(dbSandbox(env.SANDBOX), RUBROS[0].id, Date.now());
		if (sb)
			cookies.set(COOKIE, sb.id, {
				path: '/',
				httpOnly: true,
				secure: !dev,
				sameSite: 'lax',
				maxAge: VIDA_MS / 1000
			});
	}
	redirect(303, '/partners/panel');
};
