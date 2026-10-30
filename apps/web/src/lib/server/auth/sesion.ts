import type { Cookies } from '@sveltejs/kit';
import { dev } from '$app/env';

/**
 * Cookie de sesión del panel. `__Host-` en producción: sólo HTTPS, sin Domain y con Path=/, así no
 * la puede pisar un subdominio. En desarrollo (http://localhost) va sin el prefijo.
 */
export const COOKIE_SESION = dev ? 'tlw_sesion' : '__Host-tlw_sesion';
/** El nombre de antes del cambio de marca: se sigue aceptando hasta que venza, así nadie se desloguea. */
export const COOKIE_SESION_VIEJA = dev ? 'mostrador_sesion' : '__Host-mostrador_sesion';

export function guardarSesion(cookies: Cookies, token: string, venceEn: string) {
	cookies.delete(COOKIE_SESION_VIEJA, { path: '/', secure: !dev });
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
	cookies.delete(COOKIE_SESION_VIEJA, { path: '/', secure: !dev });
}
