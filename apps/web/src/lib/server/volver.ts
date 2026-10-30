import type { Cookies } from '@sveltejs/kit';

/**
 * Adónde volver después de conectar Mercado Pago o Mercado Libre: el onboarding pide volver al
 * mismo paso en vez de terminar en Integraciones. Sólo rutas propias del panel o del onboarding.
 */
const COOKIE = 'tlw_volver';
const PERMITIDA = /^\/(onboarding|panel)(\/[\w/-]*)?(\?[\w=&%-]*)?$/;

export function guardarVuelta(cookies: Cookies, volver: string | null, segura: boolean) {
	if (!volver || !PERMITIDA.test(volver)) return;
	cookies.set(COOKIE, volver, {
		path: '/integraciones',
		httpOnly: true,
		sameSite: 'lax',
		secure: segura,
		maxAge: 15 * 60
	});
}

/** La vuelta guardada (y la borra), o la de siempre: Integraciones del panel. */
export function leerVuelta(cookies: Cookies, param: string, ok: boolean) {
	const v = cookies.get(COOKIE);
	cookies.delete(COOKIE, { path: '/integraciones' });
	const base = v && PERMITIDA.test(v) ? v : '/panel/integraciones';
	return `${base}${base.includes('?') ? '&' : '?'}${param}=${ok ? 'conectada' : 'error'}`;
}
