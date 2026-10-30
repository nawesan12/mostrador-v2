import type { Cookies } from '@sveltejs/kit';

/**
 * Volver a donde se pidió entrar (el consentimiento de una IA, o el panel cuando venció la sesión): el formulario postea a
 * `?/ingresar` y pierde la consulta, y Google o Apple pasan por otro sitio, así que va en una cookie
 * corta.
 */
export const VOLVER = 'mostrador-volver';
export const volverValido = (v: string | null | undefined): v is string =>
	!!v &&
	v.length < 2000 &&
	!v.includes('\\') &&
	(v.startsWith('/oauth/autorizar?') || /^\/(panel|onboarding)(\/[^/]|\?|$)/.test(v));

/** A dónde ir después de entrar (y se olvida). */
export function destino(cookies: Cookies) {
	const v = cookies.get(VOLVER);
	cookies.delete(VOLVER, { path: '/' });
	return volverValido(v) ? v : '/panel';
}
