import { dev } from '$app/env';
import type { Handle } from '@sveltejs/kit/hooks';
import { env } from 'cloudflare:workers';
import { csrfForbidden } from '#lib/server/csrf.ts';
import { COOKIE, dbSandbox, leer } from '#lib/server/sandbox/sandbox.ts';
import { COOKIE_SESION } from '#lib/server/auth/sesion.ts';

/**
 * Rutas que usan el sandbox de la demo: sólo ahí se lee D1 (la landing no gasta lecturas).
 * /oauth/autorizar ata la IA al sandbox del visitante; /api/mcp lo saca del token, no de la cookie.
 */
const CON_SANDBOX = /^\/(demo|panel|integraciones|oauth)(\/|$)/;

/**
 * Encabezados de seguridad en todas las respuestas (también las tempranas, como el rechazo por CSRF).
 * Nadie puede meter el panel en un iframe, el navegador no adivina tipos y siempre se usa HTTPS.
 * El iframe "Panel de Mostrador en el celular" de la landing es del mismo origen: SAMEORIGIN.
 */
function secure(res: Response): Response {
	const set = (r: Response) => {
		r.headers.set('X-Content-Type-Options', 'nosniff');
		r.headers.set('Referrer-Policy', 'strict-origin-when-cross-origin');
		r.headers.set('Permissions-Policy', 'camera=(self), microphone=(self), geolocation=()');
		r.headers.set('X-Frame-Options', 'SAMEORIGIN');
		if (!dev) r.headers.set('Strict-Transport-Security', 'max-age=31536000; includeSubDomains');
		return r;
	};
	try {
		return set(res);
	} catch {
		// Encabezados inmutables (respuesta de fetch o de la caché): se copia.
		return set(new Response(res.body, res));
	}
}

export const handle: Handle = async ({ event, resolve }) => {
	if (!dev && csrfForbidden(event.request, event.url))
		return secure(
			new Response(`Cross-site ${event.request.method} form submissions are forbidden`, {
				status: 403
			})
		);
	const started = Date.now();
	event.locals.sandbox = null;
	event.locals.token = null;
	if (String(env.DATA_SOURCE) === 'api') {
		// La sesión la valida Go en cada pedido del panel; acá sólo se lee la cookie.
		event.locals.token = event.cookies.get(COOKIE_SESION) ?? null;
	} else if (CON_SANDBOX.test(event.url.pathname)) {
		event.locals.sandbox = await leer(dbSandbox(env.SANDBOX), event.cookies.get(COOKIE));
	}
	const res = secure(await resolve(event));
	try {
		res.headers.set('Server-Timing', `app;dur=${Date.now() - started}`);
	} catch {
		// Respuesta inmutable: el diagnóstico no vale una copia.
	}
	return res;
};
