/** Facturación en el motor de listas (handoff PanelListas `CFG.Facturación`). */
import * as I from '#lib/icons.ts';
import { pesos, pesosCortos } from '../formato.ts';
import { rutaDe } from '../navegacion.ts';
import { CANALES, hace } from './clientes.ts';
import type { ConfigLista } from './tipos.ts';

export type Comprobante = {
	id: string;
	numero: string;
	tipo: 'A' | 'B';
	cliente: string;
	dias: number;
	estado: 'ok' | 'pend' | 'rej' | 'void';
	total: number;
	canal: string;
	enviado: boolean;
	pedido: number | null;
};

const ESTADO = {
	ok: ['Autorizada por ARCA', '#2F8F63', I.phSealCheckFill],
	pend: ['Esperando a ARCA', '#9A6A10', I.phHourglassMediumFill],
	rej: ['Rechazada: CUIT inválido', '#B83A2E', I.phXCircleFill],
	void: ['Anulada con nota de crédito', 'var(--mu)', I.phProhibitBold]
} as const;
const TIPO = { A: ['#ECEBFB', '#4A47A3'], B: ['#FCEDE5', '#C24E25'] } as const;

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
			`${l.total.toLocaleString('es-AR')} comprobantes · ${pesosCortos(l.resumen.facturado ?? 0)} facturado`,
		primario: { t: 'Emitir factura', icon: I.phReceiptFill, hacer: emitir },
		tip: (l) => ({
			texto: `Facturaste ${pesosCortos(l.resumen.ultimos30 ?? 0)} en los últimos 30 días. En Finanzas ves cuánto te falta para el tope de tu categoría.`,
			boton: 'Ver en Finanzas',
			async hacer() {
				location.href = rutaDe('Finanzas');
			}
		}),
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
			{ k: 'void', t: 'Anuladas', c: '#9C8B7E' }
		],
		facetas: [
			{
				k: 'tipo',
				t: 'Tipo',
				opciones: [
					{ k: 'A', t: 'Factura A' },
					{ k: 'B', t: 'Factura B' }
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
			const [vL, vC, vI] = ESTADO[c.estado];
			const [bg, fg] = TIPO[c.tipo];
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
				sub: `Factura ${c.tipo} ${c.numero} · ${hace(c.dias)}${c.pedido ? ` · pedido #${c.pedido}` : ''}`,
				vL,
				vC,
				vI,
				v: pesos(c.total),
				vs: c.tipo === 'A' ? `IVA ${pesos((c.total * 0.21) / 1.21)}` : 'final',
				act:
					c.estado === 'rej'
						? { t: 'Corregir', icon: I.phPencilSimpleFill, bg: 'rgba(184,58,46,.1)', fg: '#B83A2E' }
						: c.estado === 'pend' || c.estado === 'void'
							? undefined
							: c.enviado
								? { t: 'PDF', icon: I.phFilePdfFill, bg: 'var(--chip)', fg: 'var(--tx)' }
								: { t: 'Enviar', icon: I.phEnvelopeSimpleFill, bg: 'var(--chip)', fg: 'var(--tx)' }
			};
		},
		async accionFila(c, ctx) {
			if (c.estado === 'rej') {
				const n = await ctx.accion([c.id], 'reintentar');
				ctx.ui.toast({
					k: 'ok',
					title: `Reemitida para ${c.cliente}`,
					desc: n ? 'Con el CUIT corregido' : 'Ya estaba autorizada'
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
					ctx.ui.toast({
						k: 'ok',
						title: `${n} ${n === 1 ? 'comprobante reintentado' : 'comprobantes reintentados'}`,
						desc: 'Autorizados por ARCA'
					});
				}
			}
		]
	};
}
