import type { Cookies } from '@sveltejs/kit';
import { dev } from '$app/env';

/**
 * Cookie de sesión del panel. `__Host-` en producción: sólo HTTPS, sin Domain y con Path=/, así no
 * la puede pisar un subdominio. En desarrollo (http://localhost) va sin el prefijo.
 */
export const COOKIE_SESION = dev ? 'mostrador_sesion' : '__Host-mostrador_sesion';

export function guardarSesion(cookies: Cookies, token: string, venceEn: string) {
	cookies.set(COOKIE_SESION, token, {
		path: '/',
		httpOnly: true,
		secure: !dev,
		sameSite: 'lax',
		expires: new Date(venceEn)
	});
}

export function borrarSesion(cookies: Cookies) {
	cookies.delete(COOKIE_SESION, { path: '/', secure: !dev });
}
