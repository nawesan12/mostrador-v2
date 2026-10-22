/**
 * Devoluciones y reseñas de la demo (design/PanelVentas.dc.html): salen de pedidos entregados, y las
 * preguntas de Mercado Libre de los productos publicados, con la respuesta que propone Hornero.
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

/** Preguntas de Mercado Libre y la respuesta que propone Hornero. */
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

export function resenasDemo(
	d: Demo,
	ahora: number,
	respuestas: Map<string, string>
): ResenasYPreguntas {
	const clientes = new Map(d.clientes.map((c) => [c.id, c]));
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
	const resenas = d.resenas.map((r, k) => {
		const [n, a] = (clientes.get(r.clienteId)?.nombre ?? 'Cliente').split(' ');
		return {
			autor: `${n} ${a?.[0] ?? ''}.`,
			estrellas: r.estrellas,
			texto: r.texto,
			// Una de cada cinco llega desde Google.
			origen: k % 5 === 2 ? 'Google' : (ORIGEN[r.canal] ?? 'Tienda'),
			creada: iso(r.creada)
		};
	});
	const porEstrellas = [5, 4, 3, 2, 1].map((n) => resenas.filter((r) => r.estrellas === n).length);
	return {
		preguntas,
		resenas,
		promedio: resenas.length
			? Math.round((resenas.reduce((s, r) => s + r.estrellas, 0) / resenas.length) * 10) / 10
			: null,
		total: resenas.length,
		porEstrellas,
		recomendaria: resenas.length
			? Math.round(((porEstrellas[0] + porEstrellas[1]) / resenas.length) * 100)
			: 0
	};
}
