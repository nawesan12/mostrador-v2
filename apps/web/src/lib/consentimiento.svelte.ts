/**
 * Elección de cookies del visitante, como en el handoff: localStorage['mostrador-cookies'] guarda
 * 'all', 'nec' o un JSON {an, mk}. Analíticas (GA4) y marketing (Meta Pixel) sólo se cargan si están
 * permitidas; Cloudflare Web Analytics no usa cookies y va siempre.
 */
export type Eleccion = { an: boolean; mk: boolean };

const CLAVE = 'mostrador-cookies';

export function leerEleccion(): Eleccion | null {
	let v: string | null;
	try {
		v = localStorage.getItem(CLAVE);
	} catch {
		return null;
	}
	if (!v) return null;
	if (v === 'all') return { an: true, mk: true };
	if (v === 'nec') return { an: false, mk: false };
	try {
		const o = JSON.parse(v) as Partial<Eleccion>;
		return { an: !!o.an, mk: !!o.mk };
	} catch {
		return null;
	}
}

export function guardarEleccion(v: 'all' | 'nec' | Eleccion) {
	try {
		localStorage.setItem(CLAVE, typeof v === 'string' ? v : JSON.stringify(v));
	} catch {
		// Sin almacenamiento (modo privado): la elección vale para esta visita.
	}
	window.dispatchEvent(new CustomEvent('mostrador:cookies'));
}
