/**
 * Reglas y etiquetas del backoffice que comparten la demo y la API: el estado de una tienda (la
 * misma regla que `equipo` en Go), el precio de cada abono y el plazo legal del arrepentimiento.
 */
import type { components } from '@mostrador/contrato/esquema';
import { PLANES, precioAnual } from '#lib/panel/planes.ts';

export type EstadoTienda = components['schemas']['EstadoTiendaEquipo'];
export type Plan = (typeof PLANES)[number]['id'];

export const DIAS_PRUEBA = 14;
const DIA = 86_400_000;

/** cerrada → pausada → cancelada → cobro rechazado → prueba → activa (como Go). */
export function estadoDe(t: {
	plan: Plan;
	creadaEn: number;
	pruebaHasta?: number | null;
	canceladaEn?: number | null;
	cerradaEn?: number | null;
	pausadaEn?: number | null;
	suscripcion?: string | null;
	ahora: number;
}): EstadoTienda {
	if (t.cerradaEn) return 'cerrada';
	if (t.pausadaEn) return 'pausada';
	if (t.canceladaEn) return 'cancelada';
	if (t.plan !== 'gratis' && t.suscripcion === 'paused') return 'cobro_rechazado';
	if (t.ahora < (t.pruebaHasta ?? t.creadaEn + DIAS_PRUEBA * DIA)) return 'prueba';
	return 'activa';
}

/** Centavos por mes del abono (el anual, con el 25 % menos). */
export function abonoMensual(plan: Plan, ciclo: 'mes' | 'anio') {
	const p = PLANES.find((x) => x.id === plan)?.precio ?? 0;
	return (ciclo === 'anio' ? precioAnual(p) : p) * 100;
}

export const nombrePlan = (p: string) => PLANES.find((x) => x.id === p)?.nombre ?? p;

export const ESTADOS_TIENDA: Record<EstadoTienda, { t: string; c: string; bg: string }> = {
	prueba: { t: 'En prueba', c: '#3E6BFD', bg: 'rgba(62,107,253,.12)' },
	activa: { t: 'Activa', c: '#2F8F63', bg: 'rgba(47,143,99,.13)' },
	cobro_rechazado: { t: 'Cobro rechazado', c: '#B83A2E', bg: 'rgba(184,58,46,.1)' },
	cancelada: { t: 'Canceló el plan', c: '#9A6A10', bg: 'rgba(227,169,59,.18)' },
	pausada: { t: 'Pausada', c: '#6B5244', bg: 'var(--chip)' },
	cerrada: { t: 'Cerrada', c: '#7E6656', bg: 'var(--chip)' }
};

export const ESTADOS_CONSULTA = {
	nueva: { t: 'Nueva', c: '#C24E25', bg: 'rgba(210,88,47,.12)' },
	en_curso: { t: 'En curso', c: '#9A6A10', bg: 'rgba(227,169,59,.18)' },
	resuelta: { t: 'Resuelta', c: '#2F8F63', bg: 'rgba(47,143,99,.13)' }
} as const;

export const ESPECIALIDADES: Record<string, string> = {
	diseno: 'Diseño',
	desarrollo: 'Desarrollo',
	marketing: 'Marketing',
	contabilidad: 'Contabilidad',
	fotografia: 'Fotografía'
};

export const NIVELES: Record<string, { t: string; pct: number }> = {
	partner: { t: 'Partner', pct: 20 },
	plata: { t: 'Plata', pct: 25 },
	oro: { t: 'Oro', pct: 30 }
};

const fechaAR = (ms: number) =>
	new Intl.DateTimeFormat('en-CA', { timeZone: 'America/Argentina/Buenos_Aires' }).format(ms);

/**
 * Último día para devolver el dinero de un arrepentimiento: 10 días hábiles (lunes a viernes)
 * desde que se pidió, en hora de la Argentina. Los feriados no se descuentan.
 */
export function plazoArrepentimiento(creado: number) {
	let d = new Date(`${fechaAR(creado)}T12:00:00Z`);
	let n = 0;
	while (n < 10) {
		d = new Date(d.getTime() + DIA);
		const w = d.getUTCDay();
		if (w !== 0 && w !== 6) n++;
	}
	return d.toISOString().slice(0, 10);
}

/** Días hábiles que faltan hasta `fecha` (YYYY-MM-DD); negativo si ya pasó. */
export function habilesHasta(fecha: string, ahora: number) {
	const hoy = new Date(`${fechaAR(ahora)}T12:00:00Z`);
	const fin = new Date(`${fecha}T12:00:00Z`);
	const signo = fin >= hoy ? 1 : -1;
	let n = 0;
	for (let d = hoy; signo > 0 ? d < fin : d > fin; d = new Date(d.getTime() + signo * DIA)) {
		const w = new Date(d.getTime() + signo * DIA).getUTCDay();
		if (w !== 0 && w !== 6) n += signo;
	}
	return n;
}

/** "28 de octubre" (o con año si no es este). */
export function fechaCorta(iso: string, ahora = Date.now()) {
	const d = new Date(iso.length === 10 ? `${iso}T12:00:00-03:00` : iso);
	const mismo = fechaAR(d.getTime()).slice(0, 4) === fechaAR(ahora).slice(0, 4);
	return new Intl.DateTimeFormat('es-AR', {
		timeZone: 'America/Argentina/Buenos_Aires',
		day: 'numeric',
		month: 'long',
		...(mismo ? {} : { year: 'numeric' })
	}).format(d);
}

/** Los montos del backoffice vienen en centavos (como en Go): "$ 49.900". */
export const centavos = (c: number) => '$ ' + Math.round(c / 100).toLocaleString('es-AR');
