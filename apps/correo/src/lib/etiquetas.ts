/**
 * Etiquetas (handoff 2n): son keywords JMAP del correo ("pedido_abierto"), así las ve cualquier
 * programa y las pueden poner las reglas. El nombre sale de la keyword y el color, de las del
 * diseño; las nuevas toman uno fijo según su nombre.
 */
export const keyword = (nombre: string) =>
	nombre
		.normalize('NFD')
		.replace(/[̀-ͯ]/g, '')
		.toLowerCase()
		.replace(/[^a-z0-9]+/g, '_')
		.replace(/^_+|_+$/g, '')
		.slice(0, 60);

/** Las del diseño, con su familia de color (fondo y texto de --marBg/--marT, etc.). */
export const PROPUESTAS: [string, string][] = [
	['Presupuesto', 'mar'],
	['Pedido abierto', 'esp'],
	['Proyectos', 'mar'],
	['Proveedor', 'inf'],
	['Pago', 'ok'],
	['Reclamo', 'pr'],
	['Moroso', 'pr']
];
const FAMILIAS = ['mar', 'esp', 'inf', 'ok', 'pr', 'ce'];

/** Las keywords del sistema ($seen, $flagged…) y las de otros programas no son etiquetas. */
export const esEtiqueta = (k: string) =>
	!k.startsWith('$') && /^[a-z0-9_]{1,60}$/.test(k) && !['nonjunk', 'junk', 'notjunk'].includes(k);

export function nombre(k: string): string {
	const p = PROPUESTAS.find(([n]) => keyword(n) === k);
	if (p) return p[0];
	const t = k.replace(/_/g, ' ');
	return t.charAt(0).toUpperCase() + t.slice(1);
}

export function familia(k: string): string {
	const p = PROPUESTAS.find(([n]) => keyword(n) === k);
	if (p) return p[1];
	let h = 0;
	for (const c of k) h = (h * 31 + c.charCodeAt(0)) >>> 0;
	return FAMILIAS[h % FAMILIAS.length];
}
