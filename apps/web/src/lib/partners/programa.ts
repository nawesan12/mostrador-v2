/**
 * El programa de partners publicado en /partners: Partner 20 %, Plata 25 % (desde 5 tiendas activas)
 * y Oro 30 % (desde 20), recurrente durante 12 meses desde el alta de cada tienda. Es la misma regla
 * que apps/api/internal/dominio/partners/comision.go, con los mismos casos
 * (testdata/comisiones.json). Lo demás (resumen, próximo pago, tareas) sale de los datos del panel.
 */
import type { components } from '@mostrador/contrato/esquema';
import { PLANES } from '#lib/panel/planes.ts';
import { SITE } from '#lib/site.ts';

export type PanelPartner = components['schemas']['PanelPartner'];
export type TiendaReferida = components['schemas']['TiendaReferida'];
export type PagoPartner = components['schemas']['PagoPartner'];
export type PerfilPartner = components['schemas']['PerfilPartner'];
export type CobroPartner = components['schemas']['CobroPartner'];
export type NivelPartner = components['schemas']['NivelPartner'];
export type EstadoReferida = TiendaReferida['estado'];

export const NIVELES = [
	{ id: 'partner', nombre: 'Partner', pct: 20, tiendas: 0 },
	{ id: 'plata', nombre: 'Partner Plata', pct: 25, tiendas: 5 },
	{ id: 'oro', nombre: 'Partner Oro', pct: 30, tiendas: 20 }
] as const satisfies readonly { id: NivelPartner; nombre: string; pct: number; tiendas: number }[];

export const MESES_DE_COMISION = 12;
export const DIAS_DE_PRUEBA = 14;

export const nivel = (id: NivelPartner) => NIVELES.find((n) => n.id === id)!;
export const porcentaje = (id: string) => NIVELES.find((n) => n.id === id)?.pct ?? 0;

/** El mayor entre el nivel que dio el equipo y el que alcanzan las tiendas activas. */
export function nivelEfectivo(asignado: string, activas: number): NivelPartner {
	let mejor = 0;
	NIVELES.forEach((n, i) => {
		if (n.id === asignado || activas >= n.tiendas) mejor = Math.max(mejor, i);
	});
	return NIVELES[mejor].id;
}

/** Fecha de Argentina (UTC−3, sin horario de verano) de un instante: [año, mes 1-12, día]. */
function ar(iso: string | number) {
	const d = new Date((typeof iso === 'number' ? iso : Date.parse(iso)) - 3 * 3_600_000);
	return [d.getUTCFullYear(), d.getUTCMonth() + 1, d.getUTCDate()] as const;
}
const dos = (n: number) => String(n).padStart(2, '0');

/** "2026-10" del mes de un instante, en hora de Argentina. */
export const mesDe = (ms: number) => {
	const [a, m] = ar(ms);
	return `${a}-${dos(m)}`;
};

/** Meses calendario entre el alta y el mes ("2026-10"); 0 = el mes del alta. */
export function mesIndice(alta: string, mes: string) {
	const [a, m] = ar(alta);
	const [ma, mm] = mes.split('-').map(Number);
	return ma * 12 + mm - (a * 12 + m);
}

/** Primer día del mes en que la tienda deja de generar comisión. */
export function finComision(alta: string) {
	const [a, m] = ar(alta);
	const t = a * 12 + (m - 1) + MESES_DE_COMISION;
	return `${Math.floor(t / 12)}-${dos((t % 12) + 1)}-01`;
}

/** Lo que paga la tienda por mes (el anual, dividido 12), en pesos. */
export function precioMensual(plan: string, ciclo: string) {
	const p = PLANES.find((x) => x.id === plan)?.precio ?? 0;
	return ciclo === 'anio' ? Math.floor(Math.floor((p * 12 * 3) / 4) / 12) : p;
}

/** Comisión de una tienda en un mes, en pesos, redondeada al peso. */
export function comision(nivelId: string, plan: string, ciclo: string, alta: string, mes: string) {
	const i = mesIndice(alta, mes);
	if (i < 0 || i >= MESES_DE_COMISION) return 0;
	return Math.floor((precioMensual(plan, ciclo) * porcentaje(nivelId) + 50) / 100);
}

/** Estado de una tienda referida a una hora. */
export function estado(plan: string, alta: string, baja: boolean, ahora: number): EstadoReferida {
	if (baja) return 'baja';
	if (ahora - Date.parse(alta) < DIAS_DE_PRUEBA * 86_400_000) return 'prueba';
	return plan && plan !== 'gratis' ? 'activa' : 'gratis';
}

/** La liquidación de un mes se paga el 10 del mes siguiente. */
export function fechaDePago(mes: string) {
	const [a, m] = mes.split('-').map(Number);
	return m === 12 ? `${a + 1}-01-10` : `${a}-${dos(m + 1)}-10`;
}

/**
 * Cuándo se paga una liquidación, para mostrar: el 10 del mes siguiente; si ya pasó y falta la
 * factura, cuando llegue (no se paga sin factura).
 */
export function cuandoSePaga(
	g: Pick<PagoPartner, 'estado' | 'fechaPago' | 'facturaUrl' | 'pagadoEn'>,
	hoy: string
) {
	if (g.estado === 'pagada') return `pagado el ${fechaCorta(g.pagadoEn ?? g.fechaPago)}`;
	if (g.fechaPago >= hoy) return `se paga el ${fechaLarga(g.fechaPago)}`;
	return g.facturaUrl ? 'se paga en estos días' : 'se paga cuando subas tu factura';
}

/** El link de referido. */
export const linkDe = (codigo: string) => `${SITE}/empezar?ref=${encodeURIComponent(codigo)}`;

const dias = (a: string, b: string) => Math.round((Date.parse(b) - Date.parse(a)) / 86_400_000);
const sumarDias = (fecha: string, n: number) =>
	new Date(Date.parse(`${fecha}T12:00:00Z`) + n * 86_400_000).toISOString().slice(0, 10);

export type Tarea = { id: string; texto: string; ir: 'tiendas' | 'comisiones' | 'perfil' };

/** Lo que muestra Inicio: números del mes, próximo pago, progreso de nivel y tareas. */
export function resumen(p: PanelPartner) {
	const cuenta = (e: EstadoReferida) => p.tiendas.filter((t) => t.estado === e).length;
	const activas = cuenta('activa');
	const estimado = p.tiendas.reduce((s, t) => s + t.comisionMes, 0);
	const generan = p.tiendas.filter((t) => t.comisionMes > 0).length;

	// El próximo pago: el más viejo sin pagar; si no hay, lo que se estima del mes en curso.
	const abierto = [...p.pagos].reverse().find((g) => g.estado !== 'pagada');
	const mes = p.hoy.slice(0, 7);
	const proximo = abierto
		? {
				monto: abierto.monto,
				cuando: cuandoSePaga(abierto, p.hoy),
				estado: abierto.estado,
				mes: abierto.mes
			}
		: {
				monto: estimado,
				cuando: `se paga el ${fechaLarga(fechaDePago(mes))}`,
				estado: 'estimado' as const,
				mes
			};

	const actual = NIVELES.findIndex((n) => n.id === p.partner.nivelEfectivo);
	const sig = NIVELES[actual + 1] ?? null;
	const progreso = sig
		? {
				siguiente: sig,
				faltan: Math.max(0, sig.tiendas - activas),
				pct: Math.min(100, Math.round((activas / sig.tiendas) * 100))
			}
		: null;

	const tareas: Tarea[] = [];
	const vencen = p.tiendas.filter((t) => {
		if (t.estado !== 'prueba') return false;
		const faltan = dias(p.hoy, sumarDias(t.alta, DIAS_DE_PRUEBA));
		return faltan >= 0 && faltan <= 7;
	}).length;
	if (vencen)
		tareas.push({
			id: 'prueba',
			texto:
				vencen === 1
					? '1 tienda tuya termina la prueba esta semana: ayudala a elegir plan'
					: `${vencen} tiendas tuyas terminan la prueba esta semana: ayudalas a elegir plan`,
			ir: 'tiendas'
		});
	if (!p.partner.cobro)
		tareas.push({
			id: 'cobro',
			texto: 'Cargá tus datos de cobro para recibir las comisiones',
			ir: 'comisiones'
		});
	for (const g of p.pagos)
		if (g.estado === 'aprobada' && !g.facturaUrl)
			tareas.push({
				id: `factura-${g.mes}`,
				texto:
					g.fechaPago >= p.hoy
						? `Subí la factura de ${nombreMes(g.mes)} para que te paguemos el ${fechaLarga(g.fechaPago)}`
						: `Subí la factura de ${nombreMes(g.mes)}: la pagamos apenas llegue`,
				ir: 'comisiones'
			});
	const terminan = p.tiendas.filter(
		(t) => t.comisionMes > 0 && dias(p.hoy, t.finComision) <= 31
	).length;
	if (terminan)
		tareas.push({
			id: 'fin',
			texto:
				terminan === 1
					? '1 tienda cumple 12 meses y deja de generar comisión el mes que viene'
					: `${terminan} tiendas cumplen 12 meses y dejan de generar comisión el mes que viene`,
			ir: 'tiendas'
		});
	if (!p.partner.perfil.ciudad || !p.partner.perfil.especialidades.length)
		tareas.push({
			id: 'perfil',
			texto: 'Completá tu perfil público: ciudad, especialidades y casos',
			ir: 'perfil'
		});

	return {
		estimado,
		generan,
		activas,
		prueba: cuenta('prueba'),
		perdidas: cuenta('baja') + cuenta('gratis'),
		proximo,
		progreso,
		tareas
	};
}

/** Espacio duro (no se parte el renglón). */
const NB = String.fromCharCode(160);
const MESES = [
	'enero',
	'febrero',
	'marzo',
	'abril',
	'mayo',
	'junio',
	'julio',
	'agosto',
	'septiembre',
	'octubre',
	'noviembre',
	'diciembre'
];
/** "octubre 2026" (con espacio duro: no se parte). */
export const nombreMes = (mes: string) => {
	const [a, m] = mes.split('-').map(Number);
	return `${MESES[m - 1]}${NB}${a}`;
};
/** "10 de noviembre" */
export const fechaLarga = (fecha: string) => {
	const [, m, d] = fecha.split('-').map(Number);
	return `${d} de${NB}${MESES[m - 1]}`;
};
/** "15/03/2026" */
export const fechaCorta = (fecha: string) => fecha.split('-').reverse().join('/');

/** CUIT o CUIL: 11 dígitos con el verificador (módulo 11), como partners.CuitValido en Go. */
export function cuitValido(c: string) {
	if (!/^\d{11}$/.test(c)) return false;
	const pesos = [5, 4, 3, 2, 7, 6, 5, 4, 3, 2];
	const s = pesos.reduce((t, p, i) => t + Number(c[i]) * p, 0);
	let v = 11 - (s % 11);
	if (v === 11) v = 0;
	if (v === 10) v = 9;
	return v === Number(c[10]);
}
