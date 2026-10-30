import { redirect } from '@sveltejs/kit';
import type { Handle } from '@sveltejs/kit/hooks';
import { abrir, COOKIE } from '#lib/server/sesion.ts';

/**
 * Sin sesión sólo se ve /ingresar. El tema sale de una cookie (oscuro por defecto) y va en el
 * <html> desde el servidor, para que no parpadee al cargar. CSP estricta: nada de afuera, y el
 * correo se dibuja en un iframe sin scripts (frame-src 'self' para el srcdoc).
 */
const CSP = [
	"default-src 'self'",
	"script-src 'self' 'unsafe-inline'",
	"style-src 'self' 'unsafe-inline'",
	"img-src 'self' data: blob: https:",
	"font-src 'self' data:",
	"connect-src 'self'",
	"frame-src 'self' blob:",
	"object-src 'none'",
	"base-uri 'self'",
	"form-action 'self'",
	"frame-ancestors 'self'"
].join('; ');

export const handle: Handle = async ({ event, resolve }) => {
	event.locals.cred = abrir(event.cookies.get(COOKIE));
	const publica = event.url.pathname === '/ingresar';
	if (!event.locals.cred && !publica) redirect(303, '/ingresar');
	const tema = event.cookies.get('correo_tema') === 'claro' ? 'claro' : 'oscuro';
	const res = await resolve(event, {
		transformPageChunk: ({ html }) => html.replace('%correo.tema%', tema)
	});
	// /adjunto trae la suya (más estricta): no se pisa.
	if (!res.headers.has('Content-Security-Policy')) res.headers.set('Content-Security-Policy', CSP);
	res.headers.set('X-Content-Type-Options', 'nosniff');
	res.headers.set('Referrer-Policy', 'no-referrer');
	// El visor muestra los PDF en un iframe del mismo sitio; de afuera nadie lo puede enmarcar.
	res.headers.set('X-Frame-Options', 'SAMEORIGIN');
	res.headers.set('Strict-Transport-Security', 'max-age=31536000');
	res.headers.set('X-Robots-Tag', 'noindex, nofollow');
	return res;
};
