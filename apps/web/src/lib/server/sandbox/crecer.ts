/**
 * Crecer en la demo (design/PanelCrecer.dc.html): marketing, canales e informes, con los números
 * de los pedidos de la tienda de ejemplo.
 */
import type { components } from '@mostrador/contrato/esquema';
import type { Demo } from './demo.ts';

type Marketing = components['schemas']['Marketing'];
type Campana = components['schemas']['Campana'];
type Canales = components['schemas']['Canales'];
type Informes = components['schemas']['Informes'];

const MIN = 60_000;
const DIA = 1440 * MIN;
const pesos = (n: number) => '$ ' + Math.round(n).toLocaleString('es-AR');
/** $ 1,1 M, o en pesos si no llega al millón. */
const millones = (n: number) =>
	n >= 1e6
		? '$ ' + (n / 1e6).toLocaleString('es-AR', { maximumFractionDigits: 1 }) + ' M'
		: pesos(n);
const mesDe = (t: number, f: 'long' | 'short' = 'long') =>
	new Date(t)
		.toLocaleDateString('es-AR', { month: f, timeZone: 'America/Argentina/Buenos_Aires' })
		.replace('.', '');

const vivos = (d: Demo, desde: number, hasta: number) =>
	d.pedidos.filter((p) => p.estado !== 'Cancelado' && p.creado > desde && p.creado <= hasta);

// ------------------------------------------------------------------ Marketing

/** Lo que hizo el visitante con una campaña. */
export type CampanaDemo = {
	id: string;
	audiencia: string;
	personas: number;
	en: number;
	programada: number | null;
};

export function marketingDemo(
	d: Demo,
	ahora: number,
	segmentos: { todos: number; vip: number; riesgo: number; categoria: number },
	cupones: { codigo: string; descripcion: string; usos: string }[],
	cupon: string | null,
	nuevas: CampanaDemo[],
	costoMensaje: number
): Marketing {
	const cat = d.catalogo.categorias[0] ?? 'productos';
	const vendidos = new Map<string, number>();
	for (const p of d.pedidos)
		for (const i of p.items)
			vendidos.set(i.productoId, (vendidos.get(i.productoId) ?? 0) + i.cantidad);
	const prod = [...d.catalogo.productos].sort(
		(a, b) => (vendidos.get(b.id) ?? 0) - (vendidos.get(a.id) ?? 0)
	)[0];
	const nombre = prod?.nombre ?? 'el producto';
	const conCupon = cupon
		? `Con tu código ${cupon} tenés 10% off hasta el domingo.`
		: 'Para vos, 10% off este finde.';
	const fecha = (t: number) =>
		new Date(t)
			.toLocaleDateString('es-AR', {
				day: 'numeric',
				month: 'short',
				timeZone: 'America/Argentina/Buenos_Aires'
			})
			.replace('.', '');
	const propias: Campana[] = nuevas.map((c) => {
		if (c.programada && c.programada > ahora)
			return {
				id: c.id,
				canal: 'wa',
				titulo: 'Campaña programada',
				detalle: `WhatsApp · ${c.personas} personas · sábado 10:00`,
				estado: 'Programada',
				metricas: [
					{ k: 'Personas', v: String(c.personas) },
					{ k: 'Costo', v: pesos(c.personas * costoMensaje) },
					{ k: 'Sale', v: 'Sáb 10 h' }
				]
			};
		const desde = c.programada ?? c.en;
		const min = Math.max(0, (ahora - desde) / MIN);
		const leidos = Math.min(92, Math.round(min * 6));
		return {
			id: c.id,
			canal: 'wa',
			titulo: 'Tu campaña',
			detalle: `WhatsApp · ${c.personas} personas · ${fecha(desde)}`,
			estado: min < 30 ? 'Enviando' : 'Terminada',
			metricas: [
				{ k: 'Leídos', v: `${leidos}%` },
				{ k: 'Ventas', v: String(Math.floor(c.personas * 0.11 * Math.min(1, min / 120))) },
				{ k: 'Costo', v: pesos(c.personas * costoMensaje) }
			]
		};
	});
	const anterior = mesDe(ahora - 30 * DIA);
	return {
		tienda: { nombre: d.catalogo.tienda, foto: prod?.foto ?? null },
		audiencias: [
			{ id: 'todos', nombre: 'Todos', cantidad: segmentos.todos },
			{ id: 'vip', nombre: 'VIP', cantidad: segmentos.vip },
			{ id: 'riesgo', nombre: 'En riesgo', cantidad: segmentos.riesgo },
			{ id: 'categoria', nombre: `Compraron ${cat.toLowerCase()}`, cantidad: segmentos.categoria }
		],
		costoMensaje,
		retorno: 31,
		mensajes: [
			`¡Hola {nombre}! Este finde ${nombre} tiene 10% off solo para clientes. ¿Te lo reservo?`,
			`{nombre}, ¡volvió ${nombre}! ${conCupon}`,
			`Hola {nombre} 👋 Armamos algo especial con ${nombre.toLowerCase()}. Para vos, 10% off este finde.`
		],
		producto: prod ? { nombre: prod.nombre, foto: prod.foto } : null,
		cupon,
		campanas: [
			...propias,
			{
				id: 'k1',
				canal: 'wa',
				titulo: 'Promo del finde',
				detalle: `WhatsApp · ${segmentos.todos} personas · ${fecha(ahora - 4 * DIA)}`,
				estado: 'Terminada',
				metricas: [
					{ k: 'Leídos', v: '92%' },
					{ k: 'Ventas', v: String(Math.round(segmentos.todos * 0.108)) },
					{ k: 'Retorno', v: '×38' }
				]
			},
			{
				id: 'k2',
				canal: 'email',
				titulo: `Novedades de ${anterior}`,
				detalle: `Email · ${(segmentos.todos * 5.6).toLocaleString('es-AR', { maximumFractionDigits: 0 })} personas · ${fecha(ahora - 10 * DIA)}`,
				estado: 'Terminada',
				metricas: [
					{ k: 'Abiertos', v: '38%' },
					{ k: 'Clics', v: '9%' },
					{ k: 'Ventas', v: String(Math.round(segmentos.todos * 0.058)) }
				]
			},
			{
				id: 'k3',
				canal: 'wa',
				titulo: 'Vuelta a clientes en riesgo',
				detalle: `WhatsApp · ${segmentos.riesgo} personas · automática`,
				estado: 'Activa',
				metricas: [
					{ k: 'Volvieron', v: String(Math.round(segmentos.riesgo * 0.17)) },
					{ k: 'Ventas', v: millones(Math.round(segmentos.riesgo * 0.17) * 19_000) },
					{ k: 'Retorno', v: '×24' }
				]
			}
		],
		cupones
	};
}

// ------------------------------------------------------------------ Canales

const CANALES: [Canales['canales'][number]['id'], string, number][] = [
	['web', 'Tienda online', 0],
	['wa', 'WhatsApp', 0],
	['ml', 'Mercado Libre', 0.14],
	['pos', 'Local', 0],
	['ig', 'Instagram', 0],
	['tg', 'Telegram', 0],
	['gg', 'Google Shopping', 0],
	['tk', 'TikTok Shop', 0]
];

export function canalesDemo(
	d: Demo,
	ahora: number,
	pausados: Map<string, boolean>,
	enIg: number
): Canales {
	const mes = vivos(d, ahora - 30 * DIA, ahora);
	const porCanal = new Map<string, { v: number; n: number }>();
	for (const p of mes) {
		const x = porCanal.get(p.canal) ?? { v: 0, n: 0 };
		porCanal.set(p.canal, { v: x.v + p.total, n: x.n + 1 });
	}
	const wa = porCanal.get('wa')?.n ?? 0;
	const pos = porCanal.get('pos');
	const visitas = Math.round(((porCanal.get('web')?.n ?? 0) / 0.034) * 1);
	const NOTAS: Record<string, string> = {
		web: `Sin comisión · conversión 3,4% sobre ${visitas.toLocaleString('es-AR')} visitas`,
		wa: `Tincho cerró ${wa ? 61 : 0}% de las ventas`,
		ml: 'Comisión 14% · ya cubierta con +10%',
		pos: `Ticket promedio ${pesos(pos?.n ? pos.v / pos.n : 0)}`,
		ig: `${enIg} productos etiquetados`,
		tg: 'Tincho responde las consultas',
		gg: 'Gratis · 3.100 impresiones',
		tk: 'Conectalo para vender en videos'
	};
	return {
		mes: mesDe(ahora),
		canales: CANALES.map(([id, nombre, comision]) => {
			const conectado = id !== 'tk';
			const x = porCanal.get(id);
			return {
				id,
				nombre,
				vendido: x?.v ?? 0,
				pedidos: x?.n ?? 0,
				comision,
				nota: NOTAS[id],
				activo: conectado && !pausados.get(id),
				conectado
			};
		})
	};
}

// ------------------------------------------------------------------ Informes

/** Cómo venía la tienda en el año (la curva del handoff), para los meses antes de la demo. */
const CURVA = {
	ventas: [9.8, 10.4, 11.2, 10.9, 12.6, 13.1, 12.4, 14.8, 15.2, 16.9, 17.8, 19.6],
	productos: [610, 640, 702, 688, 790, 812, 776, 902, 940, 1010, 1080, 1204],
	clientes: [64, 70, 82, 77, 98, 104, 96, 120, 128, 140, 146, 154],
	margen: [34, 33, 35, 36, 35, 37, 36, 38, 37, 39, 38, 38]
};

export function informesDemo(d: Demo, ahora: number, margen: number): Informes {
	const tramo = (k: number) => vivos(d, ahora - (k + 1) * 30 * DIA, ahora - k * 30 * DIA);
	const actual = tramo(0);
	const previo = tramo(1);
	const ventas = (ps: typeof actual) => ps.reduce((s, p) => s + p.total, 0);
	const unidades = (ps: typeof actual) =>
		ps.reduce((s, p) => s + p.items.reduce((a, i) => a + i.cantidad, 0), 0);
	const nuevos = (desde: number, hasta: number) =>
		d.clientes.filter((c) => c.alta > desde && c.alta <= hasta).length;
	const reales = {
		ventas: [ventas(previo), ventas(actual)],
		productos: [unidades(previo), unidades(actual)],
		clientes: [nuevos(ahora - 60 * DIA, ahora - 30 * DIA), nuevos(ahora - 30 * DIA, ahora)],
		margen: [margen - 1, margen]
	};
	// Los dos últimos meses son los de la demo; los diez anteriores siguen la curva.
	const serie = (k: keyof typeof CURVA) => {
		const c = CURVA[k];
		const escala = reales[k][1] / c[11];
		return [...c.slice(0, 10).map((v) => Math.round(v * escala)), ...reales[k]];
	};
	const meses = Array.from({ length: 12 }, (_, i) => mesDe(ahora - (11 - i) * 30 * DIA, 'short'));
	const ticket = (ps: typeof actual) => (ps.length ? Math.round(ventas(ps) / ps.length) : 0);
	const vs = (a: number, b: number) => (b ? Math.round(((a - b) / b) * 1000) / 10 : null);
	const sabados = actual.filter(
		(p) => p.canal === 'wa' && new Date(p.creado - 3 * 3_600_000).getUTCDay() === 6
	).length;
	const wa = actual.filter((p) => p.canal === 'wa').length;
	const veces = wa ? Math.max(1.2, (sabados * 7) / wa) : 2.3;
	const [c1, c2] = [
		d.catalogo.categorias[0] ?? 'un producto',
		d.catalogo.categorias[1] ?? d.catalogo.categorias[0] ?? 'otro'
	];
	const vendidos = new Map<string, number>();
	for (const p of actual)
		for (const i of p.items) vendidos.set(i.nombre, (vendidos.get(i.nombre) ?? 0) + i.cantidad);
	const estrella = [...vendidos].sort((a, b) => b[1] - a[1])[0]?.[0] ?? 'Tu producto estrella';
	return {
		meses,
		series: {
			ventas: serie('ventas'),
			productos: serie('productos'),
			clientes: serie('clientes'),
			margen: serie('margen')
		},
		kpis: {
			ventas: ventas(actual),
			ventasVs: vs(ventas(actual), ventas(previo)),
			ticket: ticket(actual),
			ticketVs: vs(ticket(actual), ticket(previo)),
			margen,
			margenVs: 1,
			// Lo que habría pagado en una plataforma que cobra comisión por venta (2 %).
			ahorro: Math.round((ventas(actual) * 0.021) / 1000) * 1000,
			mesAnterior: mesDe(ahora - 30 * DIA)
		},
		historia: 14,
		insights: [
			`Los sábados vendés ${veces.toLocaleString('es-AR', { maximumFractionDigits: 1 })}× más por WhatsApp. Programá tus campañas el viernes a la noche.`,
			`Quien compra ${c1.toLowerCase()} vuelve a comprar ${c2.toLowerCase()} en 34 días. Activá un recordatorio automático.`,
			`${estrella} se agota cada 3 semanas: subir el precio 6% no cambiaría la demanda.`
		]
	};
}
