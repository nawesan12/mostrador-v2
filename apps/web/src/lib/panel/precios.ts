/** Aumento por inflación (design/PanelCatalogo.dc.html, Listas de precio). */
export type Redondeo = 'no' | '90' | '500';

/** Redondeo "lindo": $ 8.990 en vez de $ 9.012, o al múltiplo de 500. */
export function redondear(v: number, r: Redondeo) {
	if (r === '90') return Math.max(90, Math.round(v / 100) * 100 - 10);
	if (r === '500') return Math.max(500, Math.round(v / 500) * 500);
	return Math.round(v);
}

export const aumentar = (precio: number, pct: number, r: Redondeo) =>
	redondear(precio * (1 + pct / 100), r);

/** El próximo lunes a las 00:00 de Argentina (UTC−3). */
export function proximoLunes(ahora: number) {
	const ar = new Date(ahora - 3 * 3_600_000);
	const dias = (8 - ar.getUTCDay()) % 7 || 7;
	return Date.UTC(ar.getUTCFullYear(), ar.getUTCMonth(), ar.getUTCDate() + dias, 3);
}
