import type { Cookies } from '@sveltejs/kit';

/**
 * De dónde viene quien se crea la cuenta (/empezar: QR de flyers, roll-up, one-pager, links de
 * partners). Va en una cookie de 90 días y se guarda en la tienda al llegar al onboarding.
 */
export const COOKIE_ORIGEN = 'mostrador-origen';

export type Origen = {
	fuente?: string;
	medio?: string;
	campana?: string;
	ref?: string;
	cupon?: string;
	llego?: string;
};

const valido = (v: string | null) => (v && /^[A-Za-z0-9_.-]{1,40}$/.test(v) ? v : undefined);

export function origenDe(url: URL, ahora = new Date()): Origen | null {
	const q = url.searchParams;
	const o: Origen = {
		fuente: valido(q.get('utm_source')),
		medio: valido(q.get('utm_medium')),
		campana: valido(q.get('utm_campaign')),
		ref: valido(q.get('ref')),
		cupon: valido(q.get('cupon'))?.toUpperCase()
	};
	const limpio = Object.fromEntries(Object.entries(o).filter(([, v]) => v)) as Origen;
	if (!Object.keys(limpio).length) return null;
	return { ...limpio, llego: ahora.toISOString() };
}

export function guardarOrigen(cookies: Cookies, o: Origen) {
	cookies.set(COOKIE_ORIGEN, JSON.stringify(o), {
		path: '/',
		httpOnly: true,
		sameSite: 'lax',
		maxAge: 90 * 86_400
	});
}

export function leerOrigen(cookies: Cookies): Origen | null {
	try {
		const o = JSON.parse(cookies.get(COOKIE_ORIGEN) ?? 'null') as Origen | null;
		return o && typeof o === 'object' ? o : null;
	} catch {
		return null;
	}
}
