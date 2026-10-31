import type { Cookies } from '@sveltejs/kit';
import { dev } from '$app/env';

/**
 * El token de una invitación al equipo mientras se va y vuelve de Google o Apple (el ida y vuelta
 * pasa por otro sitio): /invitacion lo deja y /ingresar/clerk lo usa al volver. Dura una hora.
 */
export const COOKIE_INVITACION = 'tlw_invitacion';

export const guardarInvitacion = (cookies: Cookies, token: string) =>
	cookies.set(COOKIE_INVITACION, token, {
		path: '/',
		httpOnly: true,
		secure: !dev,
		sameSite: 'lax',
		maxAge: 60 * 60
	});

/** El token guardado (y se olvida). */
export function tomarInvitacion(cookies: Cookies) {
	const t = cookies.get(COOKIE_INVITACION);
	if (t) cookies.delete(COOKIE_INVITACION, { path: '/', secure: !dev });
	return t && t.length <= 200 ? t : undefined;
}
