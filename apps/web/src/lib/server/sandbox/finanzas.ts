/**
 * Finanzas de la demo (design/PanelOperacion.dc.html, sFin): lo disponible en Mercado Pago, lo que
 * van a liquidar los demás medios, los gastos del mes y el flujo de las próximas dos semanas, con
 * los pedidos de la tienda de ejemplo y lo que hizo el visitante (campañas, retiros).
 */
import type { components } from '@mostrador/contrato/esquema';
import type { Demo } from './demo.ts';
import { DIAS_DE_PRUEBA } from './resumen.ts';

type Finanzas = components['schemas']['Finanzas'];
type Tablero = components['schemas']['Tablero'];
type Movimiento = Finanzas['movimientos'][number];

const HORA = 3_600_000;
const DIA = 24 * HORA;
const AR = -3 * HORA;
/** Lo que cuesta en promedio un envío que paga la tienda. */
const COSTO_ENVIO = 4200;
const PLAN_PRO = 49_900;

const diaAr = (t: number) => new Date(t + AR);
const fecha = (t: number) =>
	diaAr(t).toLocaleDateString('es-AR', { weekday: 'long', day: 'numeric', timeZone: 'UTC' });
const NOMBRE: Record<string, string> = {
	payway: 'Payway · tarjetas',
	modo: 'MODO · apps bancarias',
	uala: 'Ualá Bis'
};

export type Retiro = { monto: number; en: number };
export type GastoCampana = { personas: number; en: number };

export function finanzasDemo(
	d: Demo,
	ahora: number,
	t: Tablero,
	margen: number,
	retiros: Retiro[],
	campanas: GastoCampana[],
	costoMensaje: number
): Finanzas {
	const hoy = Math.floor((ahora + AR) / DIA) * DIA - AR;
	const vivos = d.pedidos.filter((p) => p.estado !== 'Cancelado' && p.creado <= ahora);
	const mes = vivos.filter((p) => p.creado > ahora - 30 * DIA);
	// Mercado Pago acredita cada día lo del día anterior: eso es lo disponible, menos lo retirado hoy.
	const mpAyer = vivos.filter((p) => p.pago === 'mp' && p.creado > hoy - DIA && p.creado <= hoy);
	const liqMp = mpAyer.reduce((s, p) => s + p.total, 0);
	const retiradoHoy = retiros.filter((r) => r.en >= hoy).reduce((s, r) => s + r.monto, 0);
	const disponible = Math.max(0, liqMp - retiradoHoy);

	const otras = t.liquidaciones.filter((l) => l.medio !== 'mp');
	const aLiquidar = otras.reduce((s, l) => s + l.monto, 0);
	const envios = mes.filter((p) => p.envio !== 'retiro');
	const gastoEnvios = envios.length * COSTO_ENVIO;
	const gastoCampanas = campanas
		.filter((c) => c.en > ahora - 30 * DIA)
		.reduce((s, c) => s + c.personas * costoMensaje, 0);
	const gastos = gastoEnvios + gastoCampanas;
	const ventasMes = mes.reduce((s, p) => s + p.total, 0);
	const ganancia = Math.round((ventasMes * margen) / 100) - gastos;

	// Flujo: lo que entra (Mercado Pago cada día, los demás en su día) y lo que sale (envíos).
	const mpDiario = Math.round(
		mes.filter((p) => p.pago === 'mp').reduce((s, p) => s + p.total, 0) / 30
	);
	const enviosDiario = Math.round(gastoEnvios / 30);
	const fin = hoy + DIAS_DE_PRUEBA * DIA + 12 * HORA;
	const flujo = Array.from({ length: 14 }, (_, i) => {
		const dia = hoy + i * DIA;
		const clave = diaAr(dia).toISOString().slice(0, 10);
		const otrasHoy = otras.filter((l) => l.fecha.slice(0, 10) === clave);
		const semana = diaAr(dia).getUTCDay();
		return {
			dia: diaAr(dia).getUTCDate(),
			entra: (i === 0 ? liqMp : mpDiario) + otrasHoy.reduce((s, l) => s + l.monto, 0),
			sale: (semana === 0 ? 0 : enviosDiario) + (Math.abs(dia - fin) < DIA / 2 ? PLAN_PRO : 0)
		};
	});

	const ultimaTransf = vivos.find((p) => p.pago === 'transf' && p.estado !== 'Sin pagar');
	const cliente = ultimaTransf
		? d.clientes.find((c) => c.id === ultimaTransf.clienteId)
		: undefined;
	const semana = envios.filter((p) => p.creado > ahora - 7 * DIA);
	const movs: (Movimiento & { en: number })[] = [
		...(liqMp
			? [
					{
						id: 'mp',
						tipo: 'mp' as const,
						titulo: 'Liquidación Mercado Pago',
						detalle: `hoy · ${mpAyer.length} ${mpAyer.length === 1 ? 'venta' : 'ventas'}`,
						monto: liqMp,
						en: hoy + 9 * HORA
					}
				]
			: []),
		...otras.map((l) => ({
			id: l.medio,
			tipo: l.medio as Movimiento['tipo'],
			titulo: NOMBRE[l.medio] ?? l.medio,
			// La fecha viene como día de Argentina: el mediodía cae en ese día.
			detalle: `${fecha(Date.parse(l.fecha.slice(0, 10)) + 15 * HORA)} · programada`,
			monto: l.monto,
			en: Date.parse(l.fecha.slice(0, 10)) + 15 * HORA
		})),
		...(semana.length
			? [
					{
						id: 'envios',
						tipo: 'andreani' as const,
						titulo: 'Andreani · envíos de la semana',
						detalle: `ayer · ${semana.length} envíos`,
						monto: -semana.length * COSTO_ENVIO,
						en: hoy - DIA + 18 * HORA
					}
				]
			: []),
		...campanas.map((c, i) => ({
			id: `campana-${i}`,
			tipo: 'gasto' as const,
			titulo: 'Campaña de WhatsApp',
			detalle: `${fecha(c.en)} · ${c.personas} mensajes`,
			monto: -c.personas * costoMensaje,
			en: c.en
		})),
		...(ultimaTransf && cliente
			? [
					{
						id: 'transf',
						tipo: 'transf' as const,
						titulo: `Transferencia de ${cliente.nombre}`,
						detalle: `conciliada con #${ultimaTransf.numero}`,
						monto: ultimaTransf.total,
						en: ultimaTransf.creado
					}
				]
			: []),
		...retiros.map((r, i) => ({
			id: `retiro-${i}`,
			tipo: 'retiro' as const,
			titulo: 'Transferencia a tu banco',
			detalle: `${fecha(r.en)} · Banco Galicia`,
			monto: -r.monto,
			en: r.en
		})),
		{
			id: 'plan',
			tipo: 'gasto',
			titulo: 'Plan Pro · Tu Local Web',
			detalle: `${fecha(fin)} · cuando termina la prueba`,
			monto: -PLAN_PRO,
			en: fin
		}
	];
	return {
		disponible,
		cuenta: 'CBU terminada en 4412',
		banco: 'Banco Galicia',
		caja: {
			ventas: mes.length,
			liquidaciones: Math.round(otras.length * 4.3) + 30,
			gastos: envios.length + campanas.length + 1,
			meta: 'Meta del mes',
			metaPct: t.meta.objetivo
				? Math.min(100, Math.round((t.meta.actual / t.meta.objetivo) * 100))
				: 0,
			falta: Math.max(0, t.meta.objetivo - t.meta.actual),
			sincronizado: new Date(ahora - 2 * 60_000).toISOString()
		},
		kpis: {
			aLiquidar,
			aLiquidarDetalle: otras.length
				? `${otras
						.map((l) => NOMBRE[l.medio]?.split(' ')[0] ?? l.medio)
						.join(', ')
						.replace(/, ([^,]*)$/, ' y $1')} esta semana`
				: 'Nada esta semana',
			gastos,
			gastosDetalle: gastoCampanas ? 'Envíos y campañas' : 'Envíos que pagó la tienda',
			ganancia,
			gananciaDetalle: '',
			margen,
			mes: diaAr(ahora).toLocaleDateString('es-AR', { month: 'long', timeZone: 'UTC' })
		},
		flujo,
		movimientos: movs
			.sort((a, b) => b.en - a.en)
			.map((m) => ({
				id: m.id,
				tipo: m.tipo,
				titulo: m.titulo,
				detalle: m.detalle,
				monto: m.monto
			})),
		conciliado: true
	};
}
