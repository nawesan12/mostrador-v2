/**
 * Los paths de la tienda vieja, como en Go (vidriera.NormalizarRuta y vidriera.HaciaValido), con
 * la misma tabla de casos (apps/api/internal/dominio/vidriera/testdata/rutas.json).
 */

export const MAX_RUTA = 500;

// Caracteres de control o la barra invertida: nunca en un path guardado.
const RARO = /[\p{Cc}\\]/u;

/**
 * Sólo el path (sin dominio, query ni fragmento), decodificado, en minúsculas, sin barras repetidas
 * ni la del final. '' si no sirve (vacío, la raíz, muy largo o con caracteres raros).
 */
export function normalizarRuta(entrada: string): string {
	let s = entrada.trim();
	const i = s.indexOf('://');
	if (i >= 0 && i < 10) {
		s = s.slice(i + 3);
		const j = s.indexOf('/');
		s = j >= 0 ? s.slice(j) : '/';
	}
	const q = s.search(/[?#]/);
	if (q >= 0) s = s.slice(0, q);
	try {
		s = decodeURIComponent(s);
	} catch {
		// Un % suelto: queda como vino, como en Go.
	}
	s = s.toLowerCase().replace(/\/{2,}/g, '/');
	if (!s.startsWith('/')) s = '/' + s;
	s = s.replace(/\/+$/, '');
	// El largo en bytes, como Go.
	if (!s || new TextEncoder().encode(s).length > MAX_RUTA || RARO.test(s)) return '';
	return s;
}

/** Un path de la tienda en Mostrador: nunca "//…" ni "/\…" (el navegador los lee como otro dominio). */
export function haciaValido(s: string): boolean {
	if (!s.startsWith('/') || s.length > MAX_RUTA) return false;
	if (s.length > 1 && (s[1] === '/' || s[1] === '\\')) return false;
	return !/[\p{Cc}\\ ]/u.test(s);
}
