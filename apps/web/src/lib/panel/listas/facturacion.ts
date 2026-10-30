/** Facturación en el motor de listas (handoff PanelListas `CFG.Facturación`). */
import * as I from '#lib/icons.ts';
import { pesos, pesosCortos, plural } from '../formato.ts';
import { rutaDe } from '../navegacion.ts';
import { CANALES, hace } from './clientes.ts';
import type { ConfigLista } from './tipos.ts';

export type Comprobante = {
	id: string;
	/** "0004-00012812" (demo) o "B 0004-00012812" (Go); notas de crédito con "NC"; sin número, "—". */
	numero: string;
	tipo: 'A' | 'B' | 'C';
	cliente: string;
	dias: number;
	estado: 'ok' | 'pend' | 'rej' | 'void' | 'nofac';
	/** Por qué ARCA la rechazó o por qué no se facturó. */
	motivo?: string;
	total: number;
	canal: string;
	enviado: boolean;
	pedido: number | null;
};

const ESTADO = {
	ok: ['Autorizada por ARCA', '#2F8F63', I.phSealCheckFill],
	pend: ['Esperando a ARCA', '#9A6A10', I.phHourglassMediumFill],
	rej: ['Rechazada', '#B83A2E', I.phXCircleFill],
	void: ['Anulada con nota de crédito', 'var(--mu)', I.phProhibitBold],
	nofac: ['No facturada', 'var(--mu)', I.phInfoFill]
} as const;
const TIPO = {
	A: ['#ECEBFB', '#4A47A3'],
	B: ['#FCEDE5', '#C24E25'],
	C: ['#E6F2EC', '#2F8F63']
} as const;

/** Es una nota de crédito (no se anula ni se reenvía como factura). */
export const esNota = (c: Comprobante) => c.numero.startsWith('NC');

/** El número sin la letra ni "NC": 0004-00012812. */
export const numeroDe = (c: Comprobante) => c.numero.replace(/^NC /, '').replace(/^[ABC] /, '');

/** "Factura B 0004-00012812", "Nota de crédito B 0004-00000301" o "Sin factura". */
export function nombreComprobante(c: Comprobante) {
	if (c.estado === 'nofac' || c.numero === '—') return 'Sin factura';
	return `${esNota(c) ? 'Nota de crédito' : 'Factura'} ${c.tipo} ${numeroDe(c)}`;
}

/** Lo que dice el estado, con el motivo real cuando ARCA rechazó o la venta no se facturó. */
export function estadoDe(c: Comprobante) {
	const [t, col, icon] = ESTADO[c.estado];
	const conMotivo = (c.estado === 'rej' || c.estado === 'nofac') && c.motivo;
	return { t: conMotivo ? `${t}: ${c.motivo}` : t, col, icon };
}

const sinArca = (t: string) => ({
	k: 'arca' as const,
	title: t,
	desc: 'Llega cuando conectes ARCA con tu certificado'
});

export function configFacturacion(emitir: () => void): ConfigLista<Comprobante> {
	return {
		lista: 'facturacion',
		titulo: 'Facturación',
		unidad: 'comprobantes',
		buscar: 'Cliente, número o pedido',
		kpi: (l) =>
			`${plural(l.total, 'comprobante', 'comprobantes')} · ${pesosCortos(l.resumen.facturado ?? 0)} facturado`,
		primario: { t: 'Emitir factura', icon: I.phReceiptFill, hacer: emitir },
		// Finanzas no calcula el tope de la categoría del monotributo: el consejo no lo promete.
		tip: (l) =>
			l.resumen.ultimos30
				? {
						texto: `Facturaste ${pesosCortos(l.resumen.ultimos30)} en los últimos 30 días. En Finanzas ves lo que entró y lo que salió.`,
						boton: 'Ver en Finanzas',
						async hacer() {
							location.href = rutaDe('Finanzas');
						}
					}
				: null,
		vacio: {
			hornero: 'ticket',
			titulo: 'Todavía no hay comprobantes',
			texto:
				'Con ARCA conectada, cada venta cobrada se factura sola y aparece acá. Si no, queda como no facturada con su motivo.',
			accion: { t: 'Ver la conexión con ARCA', href: '/panel/integraciones?arca' }
		},
		vistas: [
			{ id: 'todas', t: 'Todas', icon: I.phFilesFill, color: 'var(--mu)', filtros: {} },
			{
				id: 'resolver',
				t: 'Para resolver',
				icon: I.phWarningCircleFill,
				color: '#B83A2E',
				filtros: { seg: ['pend', 'rej'] }
			},
			{
				id: 'a',
				t: 'Facturas A',
				icon: I.phBuildingsFill,
				color: '#4A47A3',
				filtros: { facetas: { tipo: ['A'] } }
			}
		],
		segmentos: [
			{ k: 'ok', t: 'Autorizadas', c: '#2F8F63' },
			{ k: 'pend', t: 'Pendientes', c: '#E3A93B' },
			{ k: 'rej', t: 'Rechazadas', c: '#B83A2E' },
			{ k: 'void', t: 'Anuladas', c: '#9C8B7E' },
			{ k: 'nofac', t: 'No facturadas', c: '#C4B5A8' }
		],
		facetas: [
			{
				k: 'tipo',
				t: 'Tipo',
				opciones: [
					{ k: 'A', t: 'Factura A' },
					{ k: 'B', t: 'Factura B' },
					{ k: 'C', t: 'Factura C' }
				]
			},
			{ k: 'canal', t: 'Venta por', opciones: CANALES.filter((c) => c.k !== 'ig') },
			{
				k: 'per',
				t: 'Período',
				opciones: [
					{ k: 'm', t: 'Este mes' },
					{ k: 'q', t: 'Trimestre' },
					{ k: 'y', t: 'Año' }
				]
			}
		],
		rango: { t: 'Monto desde', max: () => 180_000, paso: 5000, formato: pesos },
		banderas: [{ k: 'nomail', t: 'Sin enviar al cliente' }],
		ordenes: [
			{ k: 'new', t: 'Más nuevas' },
			{ k: 'total', t: 'Mayor monto' }
		],
		fila: (c) => {
			const e = estadoDe(c);
			const [bg, fg] = TIPO[c.tipo];
			const cuando = `${hace(c.dias)}${c.pedido ? ` · pedido #${c.pedido}` : ''}`;
			return {
				ini: c.tipo,
				iniBg: bg,
				iniFg: fg,
				radio: '12px',
				insignia: { marca: c.canal },
				t: c.cliente,
				tags:
					c.enviado || c.estado !== 'ok'
						? []
						: [{ icon: I.phEnvelopeSimpleOpenBold, col: '#9A6A10', t: 'Sin enviar' }],
				// En el celular no entra la columna del estado: el motivo va abajo del nombre.
				sub:
					(c.estado === 'rej' || c.estado === 'nofac') && c.motivo
						? `${c.motivo} · ${cuando}`
						: `${nombreComprobante(c)} · ${cuando}`,
				vL: e.t,
				vC: e.col,
				vI: e.icon,
				v: pesos(c.total),
				vs: c.tipo === 'A' ? `IVA ${pesos((c.total * 0.21) / 1.21)}` : 'final',
				act:
					c.estado === 'rej'
						? { t: 'Corregir', icon: I.phPencilSimpleFill, bg: 'rgba(184,58,46,.1)', fg: '#B83A2E' }
						: c.estado !== 'ok'
							? undefined
							: c.enviado
								? { t: 'PDF', icon: I.phFilePdfFill, bg: 'var(--chip)', fg: 'var(--tx)' }
								: { t: 'Enviar', icon: I.phEnvelopeSimpleFill, bg: 'var(--chip)', fg: 'var(--tx)' }
			};
		},
		async accionFila(c, ctx) {
			if (c.estado === 'rej') {
				const n = await ctx.accion([c.id], 'reintentar');
				if (n)
					ctx.ui.toast({
						k: 'arca',
						title: `Reintentada para ${c.cliente}`,
						desc: 'Volvió a la cola de ARCA'
					});
				return;
			}
			if (c.enviado) {
				ctx.ui.toast(sinArca(`PDF de ${c.numero}`));
				return;
			}
			await ctx.accion([c.id], 'enviar');
			ctx.ui.toast({
				k: 'ok',
				title: `Factura enviada a ${c.cliente.split(' ')[0]}`,
				desc: 'Le llega por mail con el PDF'
			});
		},
		lote: [
			{
				t: 'Enviar por mail',
				icon: I.phEnvelopeSimpleFill,
				async hacer(ids, ctx) {
					const n = await ctx.accion(ids, 'enviar');
					// 0: falló (ya avisó) o no había emitidas entre las elegidas.
					if (!n) return;
					ctx.ui.toast({
						k: 'ok',
						title: `${n} ${n === 1 ? 'factura enviada' : 'facturas enviadas'}`,
						desc: 'Las anuladas y pendientes no se mandan'
					});
				}
			},
			{
				t: 'Descargar PDF',
				icon: I.phFilePdfFill,
				hacer: async (_, ctx) => void ctx.ui.toast(sinArca('PDF con QR de ARCA'))
			},
			{
				t: 'Reintentar',
				icon: I.phArrowClockwiseBold,
				async hacer(ids, ctx) {
					const n = await ctx.accion(ids, 'reintentar');
					if (!n) return;
					ctx.ui.toast({
						k: 'ok',
						title: `${n} ${n === 1 ? 'comprobante reintentado' : 'comprobantes reintentados'}`,
						desc: 'Volvieron a la cola de ARCA'
					});
				}
			}
		]
	};
}
