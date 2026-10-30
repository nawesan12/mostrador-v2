/** Inventario en el motor de listas (handoff PanelListas `CFG.Inventario`). */
import * as I from '#lib/icons.ts';
import { unsplash } from '#lib/landing/formato.ts';
import { pesosCortos } from '../formato.ts';
import type { ConfigLista } from './tipos.ts';

export type ItemInventario = {
	id: string;
	nombre: string;
	foto: string | null;
	sku: string;
	stock: number;
	min: number;
	porDia: number;
	cover: number;
	vendidos30?: number;
	st: 'out' | 'crit' | 'low' | 'ok' | 'over';
	fast: boolean;
	loc: string;
	proveedor: string;
	costo: number;
};

/**
 * Cuánto dura el stock sólo se calcula con ventas que lo justifiquen (3 unidades o más en 30
 * días); si no, "Alcanza para 500 días" sale de un ritmo inventado.
 */
export const proyecta = (r: Pick<ItemInventario, 'vendidos30'>) => (r.vendidos30 ?? 3) >= 3;

export const COLOR_ST = {
	out: '#B83A2E',
	crit: '#D2582F',
	low: '#E3A93B',
	ok: '#2F8F63',
	over: '#3E6BFD'
};
/** Lo que conviene pedir: dos veces el mínimo menos lo que hay. */
export const sugerido = (r: ItemInventario) => Math.max(1, r.min * 2 - r.stock);

export function configInventario(
	pedir: (ids: string[] | null) => void
): ConfigLista<ItemInventario> {
	return {
		lista: 'inventario',
		titulo: 'Inventario',
		unidad: 'SKU',
		buscar: 'Producto, SKU o proveedor',
		kpi: (l) =>
			`${(l.resumen.unidades ?? 0).toLocaleString('es-AR')} unidades · ${pesosCortos(l.resumen.valor ?? 0)} en mercadería`,
		primario: { t: 'Pedir mercadería', icon: I.phTruckFill, hacer: () => pedir(null) },
		tip: (l) => {
			const n = l.resumen.terminanSemana ?? 0;
			if (!n) return null;
			const p = l.resumen.proveedores ?? 1;
			return {
				texto: `${n} SKU se terminan esta semana. Armé el pedido para ${p === 1 ? 'tu proveedor' : `tus ${p} proveedores`} con lo que vendés por día.`,
				boton: 'Ver pedido',
				hacer: async () => pedir(null)
			};
		},
		vistas: [
			{ id: 'todo', t: 'Todo', icon: I.phStackFill, color: 'var(--mu)', filtros: {} },
			{
				id: 'reponer',
				t: 'Para reponer',
				icon: I.phShoppingCartFill,
				color: '#D2582F',
				filtros: { seg: ['out', 'crit', 'low'] }
			},
			{
				id: 'sobrestock',
				t: 'Sobrestock',
				icon: I.phArchiveFill,
				color: '#3E6BFD',
				filtros: { seg: ['over'] }
			}
		],
		segmentos: [
			{ k: 'out', t: 'Sin stock', c: COLOR_ST.out },
			{ k: 'crit', t: 'Crítico', c: COLOR_ST.crit },
			{ k: 'low', t: 'Bajo', c: COLOR_ST.low },
			{ k: 'ok', t: 'Bien', c: COLOR_ST.ok },
			{ k: 'over', t: 'De más', c: COLOR_ST.over }
		],
		facetas: [
			{
				k: 'loc',
				t: 'Ubicación',
				opciones: [
					{ k: 'Local', t: 'Local' },
					{ k: 'Depósito', t: 'Depósito' },
					{ k: 'Local + Depósito', t: 'Los dos' }
				]
			},
			{ k: 'proveedor', t: 'Proveedor' }
		],
		rango: {
			t: 'Alcanza para menos de',
			max: () => 60,
			paso: 1,
			formato: (v) => (v ? `${v} días` : 'Cualquiera')
		},
		banderas: [{ k: 'fast', t: 'Rotación alta (+2 por día)' }],
		ordenes: [
			{ k: 'cover', t: 'Se acaba primero' },
			{ k: 'stock', t: 'Menos unidades' },
			{ k: 'value', t: 'Más plata inmovilizada' }
		],
		fila: (r) => {
			const c = COLOR_ST[r.st];
			const tope = r.min * 4;
			const reponer = r.st === 'out' || r.st === 'crit' || r.st === 'low';
			return {
				img: r.foto ? unsplash(r.foto, 96, '&h=96') : null,
				ini: r.nombre.slice(0, 2).toUpperCase(),
				iniBg: 'var(--sf3)',
				iniFg: 'var(--mu)',
				radio: '12px',
				t: r.nombre,
				sub: `${r.sku ? `${r.sku} · ` : ''}${r.loc} · ${r.proveedor}`,
				medidor: { w: Math.min(100, (r.stock / tope) * 100), c, marca: (r.min / tope) * 100 },
				vL:
					r.stock === 0
						? 'Agotado'
						: !proyecta(r)
							? (r.vendidos30 ?? 0) === 0
								? 'Sin ventas en 30 días'
								: 'Pocas ventas para proyectar'
							: `Alcanza para ${r.cover} ${r.cover === 1 ? 'día' : 'días'}`,
				vC: c,
				vI: r.st === 'out' || r.st === 'crit' ? I.phWarningFill : undefined,
				v: `${r.stock} u.`,
				vs: `mínimo ${r.min}`,
				stepper: r.stock,
				act: reponer
					? { t: 'Pedir', icon: I.phShoppingCartFill, bg: 'rgba(210,88,47,.14)', fg: '#b0441d' }
					: undefined
			};
		},
		async ajustar(r, delta, ctx) {
			await ctx.accion([r.id], 'ajustar', delta);
		},
		async accionFila(r) {
			pedir([r.id]);
		},
		lote: [
			{ t: 'Pedir al proveedor', icon: I.phShoppingCartFill, hacer: async (ids) => pedir(ids) },
			{
				t: 'Mover a local',
				icon: I.phArrowsLeftRightBold,
				async hacer(ids, ctx) {
					ctx.ui.toast({
						k: 'info',
						title: `Mover ${ids.length} SKU al local`,
						desc: 'El stock por ubicación llega con el depósito: por ahora es un solo stock.'
					});
				}
			},
			{
				t: 'Imprimir códigos',
				icon: I.phBarcodeBold,
				async hacer(ids) {
					window.open(
						`/panel/inventario/codigos?ids=${encodeURIComponent(ids.join(','))}`,
						'_blank',
						'noopener'
					);
				}
			}
		]
	};
}
