/**
 * Devoluciones y reseñas de la demo (design/PanelVentas.dc.html): salen de pedidos entregados, y las
 * preguntas de Mercado Libre de los productos publicados, con la respuesta que propone Tincho. Las
 * reseñas y preguntas que deja el visitante en su tienda online se suman con las mismas reglas que
 * Go (ventas/resenas.go): con compra si traen el código de un pedido cobrado, ocultar sin borrar.
 */
import type { components } from '@mostrador/contrato/esquema';
import type { Demo, Pedido } from './demo.ts';

type Devolucion = components['schemas']['Devolucion'];
type ResenasYPreguntas = components['schemas']['ResenasYPreguntas'];

export type EstadoDevolucion = Devolucion['estado'];
/** Un pedido que cargó el visitante (punto de venta o crear pedido) y su comprador. */
export type PedidoNuevoDemo = { pedido: Pedido; cliente: Demo['clientes'][number] | null };

const DIA = 86_400_000;
const iso = (t: number) => new Date(t).toISOString();

const MOTIVOS: [string, string][] = [
	['El color no es el de la foto, lo quería más oscuro.', 'Reintegro a Mercado Pago'],
	['Llegó con una pequeña marca en el borde.', 'Envío de reemplazo'],
	['Era un regalo y ya tenían uno igual.', 'Cambio por otro producto'],
	['Me quedó chico, quiero cambiarlo por uno más grande.', 'Cambio por otro producto']
];

export function devolucionesDemo(d: Demo, estados: Map<string, EstadoDevolucion>) {
	const productos = new Map(d.catalogo.productos.map((p) => [p.id, p]));
	const clientes = new Map(d.clientes.map((c) => [c.id, c]));
	// Uno de cada 80 pedidos entregados (no del local) pide devolución.
	const pedidos = d.pedidos
		.filter((p) => p.estado === 'Entregado' && p.canal !== 'pos' && p.numero % 80 === 3)
		.slice(0, 4);
	const devoluciones: Devolucion[] = pedidos.map((p, k) => {
		const i = p.items[0];
		const [motivo, resolucion] = MOTIVOS[k % MOTIVOS.length];
		const id = `d${p.numero}`;
		return {
			id,
			producto: i.nombre,
			foto: productos.get(i.productoId)?.foto ?? null,
			cliente: clientes.get(p.clienteId)?.nombre ?? '',
			pedido: p.numero,
			creada: iso(p.creado + 4 * DIA),
			motivo,
			monto: i.precio * i.cantidad,
			resolucion,
			estado: estados.get(id) ?? 'Pendiente'
		};
	});
	const entregados = d.pedidos.filter((p) => p.estado === 'Entregado').length;
	return {
		devoluciones,
		tasa: Math.round((devoluciones.length / Math.max(1, entregados)) * 1000) / 10
	};
}

/** Preguntas de Mercado Libre y la respuesta que propone Tincho. */
const PREGUNTAS: [string, (n: string) => string][] = [
	[
		'¿Hacen envíos a todo el país?',
		() =>
			'¡Hola! Sí, enviamos a todo el país por Andreani y OCA. Te llega en 48 a 72 h hábiles. ¡Saludos!'
	],
	[
		'¿Tienen stock para entrega inmediata?',
		(n) =>
			`¡Hola! Sí, tenemos ${n} en stock y lo despachamos en el día si comprás antes de las 15 h.`
	],
	[
		'¿Hacen factura A?',
		() => '¡Hola! Sí, hacemos factura A. Al comprar, cargá tu CUIT y te llega con el pedido.'
	],
	[
		'¿Tiene garantía?',
		() =>
			'¡Hola! Sí, tiene garantía de 6 meses por defectos de fábrica. Cualquier cosa nos escribís.'
	]
];
const ORIGEN: Record<string, string> = {
	ml: 'Mercado Libre',
	web: 'Tienda',
	wa: 'Tienda',
	ig: 'Tienda',
	tg: 'Tienda'
};

/** Una reseña que dejó el visitante en su tienda online (cambio `resena-nueva`). */
export type ResenaNuevaDemo = {
	id: string;
	autor: string;
	estrellas: number;
	texto: string;
	productoId: string | null;
	/** El pedido cobrado del que viene (con compra), o null. */
	pedido: number | null;
	creada: number;
};

/** Una pregunta que dejó el visitante en una ficha (cambio `pregunta-tienda`, id "t-…"). */
export type PreguntaTiendaDemo = {
	id: string;
	productoId: string;
	nombre: string;
	email: string | null;
	pregunta: string;
	creada: number;
};

/** Lo que cambió el visitante en reseñas y preguntas de su tienda. */
export type CambiosResenas = {
	nuevas: ResenaNuevaDemo[];
	ocultas: Set<string>;
	/** Respuesta pública de cada reseña. */
	respuestas: Map<string, string>;
	preguntas: PreguntaTiendaDemo[];
	/** Nombre y foto de los productos de la tienda (con los que sumó el visitante). */
	productos?: Map<string, { nombre: string; foto: string | null }>;
};
export const SIN_CAMBIOS: CambiosResenas = {
	nuevas: [],
	ocultas: new Set(),
	respuestas: new Map(),
	preguntas: []
};

/** Los estados de pago con los que un pedido cuenta como cobrado (igual que Go). */
export const COBRADO = ['Aprobado', 'Acreditada', 'Cobrado'];

const productosDe = (d: Demo, c: CambiosResenas) =>
	c.productos ??
	new Map(d.catalogo.productos.map((p) => [p.id, { nombre: p.nombre, foto: p.foto }]));

/**
 * Todas las reseñas de la tienda (las de ejemplo y las del visitante), de la más nueva a la más
 * vieja. Las de ejemplo que vienen de la tienda online cuentan como con compra: salen de un pedido.
 */
export function todasLasResenas(d: Demo, c: CambiosResenas) {
	const clientes = new Map(d.clientes.map((x) => [x.id, x]));
	const productos = productosDe(d, c);
	const ejemplo = d.resenas.map((r, k) => {
		const cliente = clientes.get(r.clienteId);
		const [n, a] = (cliente?.nombre ?? 'Cliente').split(' ');
		// Una de cada cinco llega desde Google.
		const origen = k % 5 === 2 ? 'Google' : (ORIGEN[r.canal] ?? 'Tienda');
		const id = `r${r.pedido}`;
		return {
			id,
			autor: `${n} ${a?.[0] ?? ''}.`,
			/** Como se firma en la tienda: nombre y ciudad. */
			firma: [n, cliente?.ciudad].filter(Boolean).join(', '),
			estrellas: r.estrellas,
			texto: r.texto,
			origen,
			creada: r.creada,
			producto: r.producto as string | null,
			conCompra: origen === 'Tienda',
			oculta: c.ocultas.has(id),
			respuesta: c.respuestas.get(id) ?? null,
			ejemplo: true
		};
	});
	const nuevas = c.nuevas.map((r) => ({
		id: r.id,
		autor: r.autor,
		firma: r.autor,
		estrellas: r.estrellas,
		texto: r.texto,
		origen: 'Tienda',
		creada: r.creada,
		producto: r.productoId ? (productos.get(r.productoId)?.nombre ?? null) : null,
		conCompra: r.pedido !== null,
		oculta: c.ocultas.has(r.id),
		respuesta: c.respuestas.get(r.id) ?? null,
		ejemplo: false
	}));
	return [...nuevas, ...ejemplo].sort((a, b) => b.creada - a.creada);
}

export function resenasDemo(
	d: Demo,
	ahora: number,
	respuestas: Map<string, string>,
	cambios: CambiosResenas = SIN_CAMBIOS
): ResenasYPreguntas {
	const delCatalogo = productosDe(d, cambios);
	const enMl = d.catalogo.productos.filter((p) => d.enMl.includes(p.id)).slice(0, 3);
	const preguntas = enMl.map((p, k) => {
		const [pregunta, respuesta] = PREGUNTAS[k % PREGUNTAS.length];
		const id = `q${k + 1}`;
		return {
			id,
			producto: p.nombre,
			foto: p.foto,
			pregunta,
			creada: iso(ahora - (k * 2 + 1) * 3_600_000),
			sugerida: respuesta(p.nombre.toLowerCase()),
			respuesta: respuestas.get(id) ?? null
		};
	});
	// Las de la tienda online: primero las sin responder, después las más nuevas (como Go).
	const deLaTienda = cambios.preguntas
		.filter((q) => delCatalogo.has(q.productoId))
		.map((q) => {
			const p = delCatalogo.get(q.productoId)!;
			return {
				id: q.id,
				producto: p.nombre,
				foto: p.foto,
				pregunta: q.pregunta,
				creada: iso(q.creada),
				sugerida: '',
				respuesta: respuestas.get(q.id) ?? null
			};
		})
		.sort(
			(a, b) =>
				Number(a.respuesta !== null) - Number(b.respuesta !== null) ||
				Date.parse(b.creada) - Date.parse(a.creada)
		);
	const resenas = todasLasResenas(d, cambios).map((r) => ({
		id: r.id,
		autor: r.autor,
		estrellas: r.estrellas,
		texto: r.texto,
		origen: r.origen,
		creada: iso(r.creada),
		producto: r.producto,
		conCompra: r.conCompra,
		oculta: r.oculta,
		respuesta: r.respuesta
	}));
	// El resumen cuenta sólo las visibles.
	const visibles = resenas.filter((r) => !r.oculta);
	const porEstrellas = [5, 4, 3, 2, 1].map((n) => visibles.filter((r) => r.estrellas === n).length);
	return {
		preguntas: [...deLaTienda, ...preguntas],
		resenas,
		promedio: visibles.length
			? Math.round((visibles.reduce((s, r) => s + r.estrellas, 0) / visibles.length) * 10) / 10
			: null,
		total: visibles.length,
		porEstrellas,
		recomendaria: visibles.length
			? Math.round(((porEstrellas[0] + porEstrellas[1]) / visibles.length) * 100)
			: 0
	};
}
