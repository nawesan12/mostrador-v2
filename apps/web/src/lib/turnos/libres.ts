/**
 * Disponibilidad de turnos (regla 1 de la especificación de turnos con seña). Función pura: la usan
 * la tienda (qué horarios ofrecer), el servidor al reservar (re-chequeo) y la demo. Su espejo en Go
 * es apps/api/internal/dominio/turnos/libres.go; los dos corren la misma tabla de casos
 * (testdata/libres.json, copiada en libres.casos.json).
 *
 * - Grilla cada 15 min desde el `desde` de cada franja; el servicio tiene que entrar entero en una
 *   franja (no cruza dos) y las franjas superpuestas no repiten horarios.
 * - Vista: hoy y los diasVista − 1 días siguientes. Anticipación: t ≥ ahora + anticipacionMin.
 * - Ocupa un turno que no está cancelado; un reservado sin seña pagada, sólo hasta `venceEn`.
 * - El resultado lleva el profesional pedido (null = cualquiera); al reservar, `asignar` elige el
 *   primero libre por orden.
 *
 * Todo en hora de Argentina (UTC−3 fijo, sin horario de verano).
 */
import type { components } from '@mostrador/contrato/esquema';

export type Franja = components['schemas']['Franja'];
export type AjustesTurnos = components['schemas']['AjustesTurnos'];
export type Servicio = components['schemas']['Servicio'];
export type Profesional = components['schemas']['Profesional'];
export type Turno = components['schemas']['Turno'];
export type EstadoTurno = components['schemas']['EstadoTurno'];
export type HorarioLibre = components['schemas']['HorarioLibre'];

/** Lo que importa de un turno tomado para saber si ocupa a alguien. */
export type Ocupado = Pick<
	Turno,
	'profesionalId' | 'inicio' | 'fin' | 'estado' | 'senaPagada' | 'venceEn'
>;

export type Consulta = {
	ajustes: Pick<AjustesTurnos, 'anticipacionMin' | 'diasVista' | 'horario'>;
	servicio: Pick<Servicio, 'id' | 'duracion' | 'activo'>;
	profesionales: Pick<Profesional, 'id' | 'servicios' | 'activo' | 'orden'>[];
	turnos: Ocupado[];
	/** null = cualquiera. */
	profesionalId: string | null;
	dia: string;
	/** Milisegundos. */
	ahora: number;
};

const MIN = 60_000;
/** Argentina: UTC−3. */
const OFFSET = 3 * 60 * MIN;
export const PASO = 15;

const dos = (n: number) => String(n).padStart(2, '0');
export const minutos = (hhmm: string) => Number(hhmm.slice(0, 2)) * 60 + Number(hhmm.slice(3, 5));
export const hhmm = (m: number) => `${dos(Math.floor(m / 60))}:${dos(m % 60)}`;

/** Día de Argentina ("YYYY-MM-DD") de un instante. */
export function diaAr(ms: number) {
	const d = new Date(ms - OFFSET);
	return `${d.getUTCFullYear()}-${dos(d.getUTCMonth() + 1)}-${dos(d.getUTCDate())}`;
}
/** "HH:MM" de Argentina de un instante. */
export function horaAr(ms: number) {
	const d = new Date(ms - OFFSET);
	return `${dos(d.getUTCHours())}:${dos(d.getUTCMinutes())}`;
}
/** Instante (ms) de un día y unos minutos de Argentina. */
export function instante(dia: string, minutosDelDia: number) {
	const [y, m, d] = dia.split('-').map(Number);
	return Date.UTC(y, m - 1, d) + minutosDelDia * MIN + OFFSET;
}
/** ISO con offset −03:00 ("2026-10-23T17:30:00-03:00"), como lo da Go. */
export const isoAr = (ms: number) => `${diaAr(ms)}T${horaAr(ms)}:00-03:00`;
/** Día de la semana (0 = domingo) de un "YYYY-MM-DD". */
export function diaSemana(dia: string) {
	const [y, m, d] = dia.split('-').map(Number);
	return new Date(Date.UTC(y, m - 1, d)).getUTCDay();
}
/** "YYYY-MM-DD" más n días. */
export function sumarDias(dia: string, n: number) {
	const [y, m, d] = dia.split('-').map(Number);
	const t = new Date(Date.UTC(y, m - 1, d + n));
	return `${t.getUTCFullYear()}-${dos(t.getUTCMonth() + 1)}-${dos(t.getUTCDate())}`;
}
export const fechaValida = (dia: string) =>
	/^\d{4}-\d{2}-\d{2}$/.test(dia) && sumarDias(dia, 0) === dia;
/** "HH:MM" válido (se acepta 24:00 como fin del día). */
export const horaValida = (h: string) => /^([01]\d|2[0-3]):[0-5]\d$|^24:00$/.test(h);

/** Si un turno tomado ocupa el horario: cancelado no; reservado sin seña sólo mientras no venció. */
export function ocupa(t: Ocupado, ahora: number) {
	if (t.estado === 'cancelado') return false;
	if (t.estado === 'reservado' && !t.senaPagada && t.venceEn && Date.parse(t.venceEn) <= ahora)
		return false;
	return true;
}

/** Si el profesional hace el servicio (sin lista = todos). */
export const hace = (p: Pick<Profesional, 'servicios'>, servicioId: string) =>
	!p.servicios.length || p.servicios.includes(servicioId);

/** Seña en pesos: el porcentaje del precio, redondeado al peso (igual que SenaDe en Go). */
export const senaDe = (precio: number, pct: number) => Math.floor((precio * pct + 50) / 100);

/** Si el día está en la vista: hoy y los diasVista − 1 siguientes. */
export function enVista(c: Pick<Consulta, 'ajustes' | 'dia' | 'ahora'>) {
	if (!fechaValida(c.dia)) return false;
	const hoy = diaAr(c.ahora);
	return c.dia >= hoy && c.dia < sumarDias(hoy, c.ajustes.diasVista);
}

/** Los candidatos: activos, que hacen el servicio y (si se pidió uno) sólo ese; por orden. */
function candidatos(c: Consulta) {
	return c.profesionales
		.filter(
			(p) =>
				p.activo && hace(p, c.servicio.id) && (c.profesionalId === null || p.id === c.profesionalId)
		)
		.sort((a, b) => a.orden - b.orden);
}

/** Los inicios del día (minutos) que entran enteros en alguna franja, sin repetir, ordenados. */
function inicios(c: Consulta) {
	const dia = diaSemana(c.dia);
	const vistos = new Set<number>();
	for (const f of c.ajustes.horario) {
		if (f.dia !== dia || !horaValida(f.desde) || !horaValida(f.hasta)) continue;
		const fin = minutos(f.hasta);
		for (let m = minutos(f.desde); m + c.servicio.duracion <= fin; m += PASO) vistos.add(m);
	}
	return [...vistos].sort((a, b) => a - b);
}

/** Quién está libre entre a y b (ms), el primero por orden, o null. */
function primeroLibre(c: Consulta, cand: Consulta['profesionales'], a: number, b: number) {
	const tomados = c.turnos.filter((t) => ocupa(t, c.ahora));
	return (
		cand.find(
			(p) =>
				!tomados.some(
					(t) => t.profesionalId === p.id && Date.parse(t.inicio) < b && Date.parse(t.fin) > a
				)
		)?.id ?? null
	);
}

export type Horario = { hora: string; inicio: string; libre: boolean };

/**
 * La grilla del día para el servicio (dentro de las franjas, que entren enteros, desde ahora +
 * anticipación y dentro de la vista), cada horario con si está libre: la tienda dibuja tachados los
 * que no.
 */
export function horariosDelDia(c: Consulta): Horario[] {
	if (!c.servicio.activo || c.servicio.duracion <= 0 || !enVista(c)) return [];
	const cand = candidatos(c);
	if (!cand.length) return [];
	const minimo = c.ahora + c.ajustes.anticipacionMin * MIN;
	return inicios(c).flatMap((m) => {
		const a = instante(c.dia, m);
		if (a < minimo) return [];
		const libre = primeroLibre(c, cand, a, a + c.servicio.duracion * MIN) !== null;
		return [{ hora: hhmm(m), inicio: isoAr(a), libre }];
	});
}

/** Los horarios libres de un día (regla 1), como los da GET .../turnos/libres. */
export const libres = (c: Consulta): HorarioLibre[] =>
	horariosDelDia(c)
		.filter((h) => h.libre)
		.map((h) => ({ inicio: h.inicio, profesionalId: c.profesionalId }));

/**
 * Quién atiende si se reserva en `inicio` (ISO): el pedido, o con "cualquiera" el primero libre por
 * orden. null si no es un horario libre de la grilla.
 */
export function asignar(c: Omit<Consulta, 'dia'>, inicio: string): string | null {
	const t = Date.parse(inicio);
	if (!Number.isFinite(t)) return null;
	const q = { ...c, dia: diaAr(t) };
	if (!q.servicio.activo || !enVista(q) || t < c.ahora + c.ajustes.anticipacionMin * MIN)
		return null;
	if (!inicios(q).some((m) => instante(q.dia, m) === t)) return null;
	return primeroLibre(q, candidatos(q), t, t + c.servicio.duracion * MIN);
}

/** Si [inicio, fin) del profesional se superpone con otro turno que ocupa (para el alta manual). */
export function superpone(
	turnos: (Ocupado & { id?: string })[],
	profesionalId: string,
	inicio: number,
	fin: number,
	ahora: number
) {
	return turnos.some(
		(t) =>
			t.profesionalId === profesionalId &&
			ocupa(t, ahora) &&
			Date.parse(t.inicio) < fin &&
			Date.parse(t.fin) > inicio
	);
}
