/** Cómo se muestran las piezas del panel de partners (chips, iniciales, nombres de plan). */
import { ESTADOS } from '#lib/marcas.ts';
import { PLANES } from '#lib/panel/planes.ts';
import type { EstadoReferida, PagoPartner } from './programa.ts';

type Chip = { t: string; bg: string; fg: string };
const chip = (t: string, [bg, fg]: readonly [string, string]): Chip => ({ t, bg, fg });

/** Los colores de los estados del panel (marcas.ts). */
export const ESTADO_TIENDA: Record<EstadoReferida, Chip> = {
	activa: chip('Activa', ESTADOS.Pagado),
	prueba: chip('En prueba', ESTADOS['Pago pendiente']),
	gratis: chip('Sin plan pago', ESTADOS.Entregado),
	baja: chip('De baja', ESTADOS.Cancelado)
};
export const ESTADO_PAGO: Record<PagoPartner['estado'] | 'estimado', Chip> = {
	estimado: chip('Estimado', ESTADOS.Entregado),
	pendiente: chip('Pendiente', ESTADOS['Pago pendiente']),
	aprobada: chip('Aprobada', ESTADOS['Para preparar']),
	pagada: chip('Pagada', ESTADOS.Pagado)
};

export const nombrePlan = (id: string) => PLANES.find((p) => p.id === id)?.nombre ?? id;

export const iniciales = (nombre: string) =>
	nombre
		.split(/\s+/)
		.filter((p) => /\p{L}/u.test(p[0] ?? ''))
		.slice(0, 2)
		.map((p) => p[0])
		.join('')
		.toUpperCase();

/** Fondos de las iniciales (los tonos de los avatares del panel). */
const TONOS = [
	['rgba(210,88,47,.13)', '#C24E25'],
	['rgba(47,143,99,.13)', '#2F8F63'],
	['rgba(227,169,59,.18)', '#9A6A10'],
	['rgba(62,107,253,.12)', '#3E6BFD'],
	['rgba(126,102,86,.14)', '#7E6656']
] as const;
export function tono(nombre: string) {
	let h = 0;
	for (const c of nombre) h = (h * 31 + c.charCodeAt(0)) >>> 0;
	return TONOS[h % TONOS.length];
}

/** Guarda algo del panel; devuelve el mensaje de error, si hubo. */
export async function guardar(cuerpo: Record<string, unknown>) {
	const r = await fetch('/partners/panel/guardar', {
		method: 'POST',
		headers: { 'content-type': 'application/json' },
		body: JSON.stringify(cuerpo)
	});
	if (r.ok) return null;
	const j = (await r.json().catch(() => null)) as { message?: string } | null;
	return j?.message ?? 'No pudimos guardarlo. Probá de nuevo.';
}
