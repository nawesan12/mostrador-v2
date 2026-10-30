/** Envíos en el motor de listas (handoff PanelListas `CFG.Envíos`). */
import * as I from '#lib/icons.ts';
import { plural, pesos, pesosCortos } from '../formato.ts';
import { hace } from './clientes.ts';
import type { ConfigLista, Ctx } from './tipos.ts';

export type Envio = {
	id: string;
	pedido: string;
	cliente: string;
	ciudad: string;
	transporte: 'andreani' | 'oca' | 'correo' | 'moto';
	step: number;
	estado: 'todo' | 'way' | 'late' | 'done' | 'ret';
	dias: number;
	costo: number;
	seguimiento: string;
	reclamado: boolean;
};

export const TRANSPORTE = {
	andreani: ['Andreani', '#C8102E'],
	oca: ['OCA', '#5A2D82'],
	moto: ['Moto', '#8A5A3E'],
	correo: ['Correo Arg.', '#0F4C9A']
} as const;
const COLOR = { todo: '#D2582F', way: '#3E6BFD', late: '#B83A2E', done: '#2F8F63', ret: '#9C8B7E' };

async function etiquetas(ids: string[], ctx: Ctx) {
	const n = await ctx.accion(ids, 'etiquetas');
	ctx.ui.toast({
		k: 'andreani',
		title: `${n} ${n === 1 ? 'etiqueta lista' : 'etiquetas listas'}`,
		desc: n ? 'Los pedidos pasan a En camino' : 'Sólo se imprimen los que están por despachar',
		...(n
			? {
					action: 'Deshacer',
					onAction: async () => {
						await ctx.accion(ids, 'sin-etiqueta');
					}
				}
			: {})
	});
}

export function configEnvios(): ConfigLista<Envio> {
	return {
		lista: 'envios',
		titulo: 'Envíos',
		unidad: 'envíos',
		buscar: 'Cliente, pedido, ciudad o seguimiento',
		kpi: (l) =>
			`${plural(l.total, 'envío', 'envíos')} · ${pesosCortos(l.resumen.costos ?? 0)} en costos`,
		primario: {
			t: 'Imprimir etiquetas',
			icon: I.phPrinterFill,
			async hacer(ctx) {
				await etiquetas([], ctx);
			}
		},
		vacio: {
			hornero: 'envio',
			titulo: 'Todavía no hay envíos',
			texto:
				'Cuando un pedido salga por moto o por Correo Argentino, lo seguís desde acá. Primero elegí cómo entregás.',
			accion: { t: 'Configurar envíos y retiro', href: '/panel/envios/configurar' }
		},
		tip: (l) => {
			const n = l.resumen.demorados ?? 0;
			if (!n) return null;
			return {
				texto: `${n} ${n === 1 ? 'envío lleva' : 'envíos llevan'} más de 2 días sin moverse. ¿Hago el reclamo a cada transporte y les aviso a los clientes?`,
				boton: 'Reclamar todos',
				async hacer(ctx) {
					const hechos = await ctx.accion([], 'reclamar-todos');
					ctx.ui.toast({
						k: 'ai',
						title: 'Reclamos enviados',
						desc: `${hechos} ${hechos === 1 ? 'envío' : 'envíos'} · les avisamos a los clientes`
					});
				}
			};
		},
		vistas: [
			{ id: 'todos', t: 'Todos', icon: I.phPackageFill, color: 'var(--mu)', filtros: {} },
			{
				id: 'hoy',
				t: 'Para hoy',
				icon: I.phLightningFill,
				color: '#D2582F',
				filtros: { seg: ['todo'] }
			},
			{
				id: 'problemas',
				t: 'Con problemas',
				icon: I.phWarningFill,
				color: '#B83A2E',
				filtros: { seg: ['late', 'ret'] }
			}
		],
		segmentos: [
			{ k: 'todo', t: 'Por despachar', c: COLOR.todo },
			{ k: 'way', t: 'En camino', c: COLOR.way },
			{ k: 'late', t: 'Demorados', c: COLOR.late },
			{ k: 'done', t: 'Entregados', c: COLOR.done },
			{ k: 'ret', t: 'Devueltos', c: COLOR.ret }
		],
		facetas: [
			{
				k: 'transporte',
				t: 'Transporte',
				opciones: (Object.keys(TRANSPORTE) as (keyof typeof TRANSPORTE)[]).map((k) => ({
					k,
					t: TRANSPORTE[k][0],
					icon: I.phTruckFill,
					col: TRANSPORTE[k][1]
				}))
			},
			{ k: 'ciudad', t: 'Destino' }
		],
		rango: { t: 'Costo de envío desde', max: () => 9000, paso: 500, formato: pesos },
		banderas: [{ k: 'wk', t: 'Últimos 7 días' }],
		ordenes: [
			{ k: 'new', t: 'Más nuevos' },
			{ k: 'late', t: 'Demorados primero' },
			{ k: 'cost', t: 'Más caros' }
		],
		fila: (e) => {
			const [nombre, color] = TRANSPORTE[e.transporte];
			const c = COLOR[e.estado];
			return {
				ini: nombre.slice(0, 2).toUpperCase(),
				iniBg: color,
				iniFg: '#fff',
				radio: '12px',
				t: e.cliente,
				sub: `${e.pedido} · ${e.ciudad} · ${e.seguimiento}`,
				pasos: ['Etiqueta', 'Retirado', 'En viaje', 'Entregado'].map((_, j) =>
					j < e.step ? c : 'var(--sf3)'
				),
				vL: {
					todo: 'Sin despachar',
					way: `Llega ${e.step >= 2 ? 'mañana' : 'en 2 días'}`,
					late: e.reclamado ? 'Demorado · reclamo hecho' : 'Demorado · sin novedad',
					done: `Entregado ${hace(e.dias)}`,
					ret: 'Devuelto al remitente'
				}[e.estado],
				vC: c,
				vI: e.estado === 'late' ? I.phWarningFill : undefined,
				chip:
					e.estado === 'late' && !e.reclamado
						? { t: 'Reclamar', bg: 'rgba(184,58,46,.1)', fg: '#B83A2E' }
						: undefined,
				v: pesos(e.costo),
				vs: nombre,
				act:
					e.estado === 'todo'
						? { t: 'Etiqueta', icon: I.phPrinterFill, bg: '#2A140C', fg: '#fff' }
						: e.estado === 'late' && !e.reclamado
							? { t: 'Reclamar', icon: I.phMegaphoneFill, bg: 'rgba(184,58,46,.1)', fg: '#B83A2E' }
							: e.estado === 'way' || e.estado === 'late'
								? { t: 'Seguir', icon: I.phMapPinFill, bg: 'var(--chip)', fg: 'var(--tx)' }
								: undefined
			};
		},
		async accionFila(e, ctx) {
			const nombre = TRANSPORTE[e.transporte][0];
			if (e.estado === 'todo') return etiquetas([e.id], ctx);
			if (e.estado === 'late' && !e.reclamado) {
				await ctx.accion([e.id], 'reclamar');
				ctx.ui.toast({
					k: 'ok',
					title: `Reclamo enviado a ${nombre}`,
					desc: `${e.pedido} · ${e.ciudad}`
				});
				return;
			}
			ctx.ui.toast({
				k: 'info',
				title: `Seguimiento ${e.seguimiento}`,
				desc: `${e.pedido} · en camino a ${e.ciudad}`
			});
		},
		lote: [
			{ t: 'Etiquetas', icon: I.phPrinterFill, hacer: etiquetas },
			{
				t: 'Pedir retiro',
				icon: I.phHandPointingFill,
				async hacer(ids, ctx) {
					ctx.ui.toast({
						k: 'info',
						title: `Retiro de ${ids.length} ${ids.length === 1 ? 'paquete' : 'paquetes'}`,
						desc: 'Se pide al transporte cuando conectes su cuenta'
					});
				}
			},
			{
				t: 'Avisar',
				icon: I.phWhatsappLogoFill,
				async hacer(ids, ctx) {
					ctx.ui.toast({
						k: 'wa',
						title: `Aviso a ${ids.length} ${ids.length === 1 ? 'cliente' : 'clientes'}`,
						desc: 'Sale por WhatsApp cuando conectes WhatsApp Cloud'
					});
				}
			}
		]
	};
}
