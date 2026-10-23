/**
 * CUIT: los 11 dígitos y el dígito verificador (el mismo cálculo que `arca.CUITValido` en Go,
 * apps/api/internal/integraciones/arca/clave.go). Se usa en el panel y en el servidor.
 */
const PESOS = [5, 4, 3, 2, 7, 6, 5, 4, 3, 2];

/** Saca guiones, espacios y puntos. */
export const limpiarCuit = (s: string) => s.replace(/[-\s.]/g, '');

export function cuitValido(cuit: string): boolean {
	if (!/^\d{11}$/.test(cuit)) return false;
	const s = PESOS.reduce((t, p, i) => t + Number(cuit[i]) * p, 0);
	let v = 11 - (s % 11);
	if (v === 11) v = 0;
	else if (v === 10) v = 9;
	return Number(cuit[10]) === v;
}

/** 20448592341 → 20-44859234-1 (mientras se escribe, deja lo que haya). */
export function formatoCuit(s: string): string {
	const d = s.replace(/\D/g, '').slice(0, 11);
	if (d.length <= 2) return d;
	if (d.length <= 10) return `${d.slice(0, 2)}-${d.slice(2)}`;
	return `${d.slice(0, 2)}-${d.slice(2, 10)}-${d.slice(10)}`;
}
