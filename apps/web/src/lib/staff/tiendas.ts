/** Tiendas en el motor de listas del panel, para el backoffice del equipo. */
import * as I from '#lib/icons.ts';
import { pesosCortos } from '#lib/panel/formato.ts';
import type { ConfigLista } from '#lib/panel/listas/tipos.ts';
import type { components } from '@mostrador/contrato/esquema';
import { centavos, ESTADOS_TIENDA, fechaCorta, nombrePlan } from './estado.ts';

type FilaTienda = components['schemas']['FilaTiendaEquipo'];

const ini = (n: string) =>
	n
		.split(' ')
		.map((w) => w[0])
		.slice(0, 2)
		.join('');

export function configTiendas(): ConfigLista<FilaTienda> {
	return {
		lista: 'tiendas-equipo',
		titulo: 'Tiendas',
		unidad: 'tiendas',
		buscar: 'Nombre, dirección, titular o email',
		kpi: (l) =>
			`${l.total.toLocaleString('es-AR')} tiendas · ${pesosCortos((l.resumen.mrr ?? 0) / 100)} por mes`,
		vistas: [
			{ id: 'todas', t: 'Todas', icon: I.phStorefrontFill, color: 'var(--mu)', filtros: {} },
			{
				id: 'prueba',
				t: 'En prueba',
				icon: I.phHourglassFill,
				color: '#3E6BFD',
				filtros: { seg: ['prueba'] }
			},
			{
				id: 'rechazado',
				t: 'Cobro rechazado',
				icon: I.phWarningFill,
				color: '#B83A2E',
				filtros: { seg: ['cobro_rechazado'] }
			}
		],
		segmentos: (Object.keys(ESTADOS_TIENDA) as (keyof typeof ESTADOS_TIENDA)[]).map((k) => ({
			k,
			t: ESTADOS_TIENDA[k].t,
			c: ESTADOS_TIENDA[k].c
		})),
		facetas: [
			{
				k: 'plan',
				t: 'Plan',
				opciones: ['gratis', 'emprende', 'pro', 'negocio'].map((p) => ({ k: p, t: nombrePlan(p) }))
			},
			{ k: 'rubro', t: 'Rubro' },
			{ k: 'plantilla', t: 'Plantilla' },
			{ k: 'partner', t: 'La trajo' }
		],
		rango: { t: 'Alta hace menos de (días)', max: () => 365, paso: 7, formato: (v) => `${v} días` },
		banderas: [],
		ordenes: [
			{ k: 'recientes', t: 'Más nuevas' },
			{ k: 'mrr', t: 'Más pagan' },
			{ k: 'nombre', t: 'A–Z' }
		],
		fila: (t) => {
			const e = ESTADOS_TIENDA[t.estado];
			return {
				ini: ini(t.nombre),
				iniBg: '#FCEDE5',
				iniFg: '#C24E25',
				radio: '12px',
				t: t.nombre,
				sub: `${t.titular} · ${t.rubro} · ${t.plantilla}`,
				vL: `Alta el ${fechaCorta(t.creadaEn)}`,
				vC: 'var(--mu)',
				chip: { t: e.t, bg: e.bg, fg: e.c },
				v: t.mrr ? centavos(t.mrr) : nombrePlan(t.plan),
				vs: t.mrr ? `${nombrePlan(t.plan)} · por mes` : t.partner || 'sin partner'
			};
		},
		lote: []
	};
}
