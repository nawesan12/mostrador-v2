/**
 * Servidor MCP de Tu Local Web (transporte Streamable HTTP, sin estado), como el de Caja: las
 * herramientas trabajan sobre `Datos` (la misma interfaz que usa el panel), así que sirven igual
 * para la demo y, cuando llegue, para la API Go. No inventan datos: todo sale de métodos que ya
 * existen. Los montos van en pesos enteros, como en el panel.
 */
import { pesos } from '#lib/panel/formato.ts';
import type {
	CuponNuevo,
	Datos,
	EstadoPedido,
	FiltrosLista,
	FiltrosPedidos,
	PedidoDetalle
} from '../datos/tipos.ts';
import type { Permiso } from './oauth.ts';

export const PROTOCOL_VERSIONS = ['2025-11-25', '2025-06-18', '2025-03-26', '2024-11-05'];
export const SERVER_INFO = { name: 'mostrador', title: 'Tu Local Web', version: '1.0.0' };
export const INSTRUCTIONS = [
	'Tu Local Web es la plataforma de una tienda online argentina: ventas, pedidos, productos, clientes y cobros.',
	'Los montos son pesos argentinos enteros (12500 = $ 12.500), sin centavos.',
	'Las fechas son de Argentina (UTC−3); "hoy" es el día de Argentina.',
	'Un pedido sólo avanza una etapa por vez (Sin pagar → Para preparar → Listo para despachar → En camino → Entregado), vuelve una o se cancela.',
	'Antes de cambiar el estado de un pedido, un precio o el stock, o de crear un cupón, confirmá con la persona.',
	'Respondé en español rioplatense.'
].join(' ');

const ESTADOS: EstadoPedido[] = [
	'Sin pagar',
	'Para preparar',
	'Listo para despachar',
	'En camino',
	'Entregado',
	'Cancelado'
];

/** Permisos del panel (Integraciones › Manejá Tu Local Web desde tu IA) que pide cada acción. */
type PermisoPanel = 'escribir' | 'plata';
const PERMISO_PANEL: Record<PermisoPanel, string> = {
	escribir: 'Cargar y editar productos',
	plata: 'Reembolsos y descuentos'
};

type JsonSchema = Record<string, unknown>;
type Herramienta = {
	name: string;
	title: string;
	description: string;
	inputSchema: JsonSchema;
	/** Cambia algo: hace falta el permiso `acciones`. */
	accion?: boolean;
	/** Además, el interruptor del panel. */
	panel?: PermisoPanel;
	/** No se deshace sola: la IA pide confirmación. */
	destructiva?: boolean;
};

const LIMITE = (max: number, def: number) => ({
	type: 'integer',
	minimum: 1,
	maximum: max,
	default: def,
	description: `Cantidad máxima de resultados (por defecto ${def}).`
});

export const HERRAMIENTAS: Herramienta[] = [
	{
		name: 'resumen_del_dia',
		title: 'Resumen del día',
		description:
			'Cómo viene el día de la tienda: ventas de hoy (en pesos) contra el mismo tramo de ayer, cantidad de pedidos, ticket promedio, pedidos para preparar, chats sin responder, transferencias por confirmar, productos sin stock, la meta del mes y los últimos pedidos.',
		inputSchema: { type: 'object', properties: {}, additionalProperties: false }
	},
	{
		name: 'listar_pedidos',
		title: 'Listar o buscar pedidos',
		description:
			'Lista pedidos del más nuevo al más viejo, con filtros. Cada uno trae número, cliente, ciudad, canal, medio de pago, envío, estado, total en pesos, fecha y productos. Para ver el detalle completo usá ver_pedido.',
		inputSchema: {
			type: 'object',
			properties: {
				texto: {
					type: 'string',
					description: 'Busca por número, cliente, ciudad o producto.'
				},
				estado: { type: 'string', enum: ESTADOS },
				canal: {
					type: 'string',
					enum: ['web', 'wa', 'ml', 'ig', 'pos', 'tg'],
					description:
						'web (tienda online), wa (WhatsApp), ml (Mercado Libre), ig (Instagram), pos (local), tg (Telegram).'
				},
				fecha: {
					type: 'string',
					enum: ['hoy', '7d', '30d', 'todo'],
					default: 'todo',
					description: 'hoy: últimas 24 h; 7d y 30d: últimos 7 o 30 días.'
				},
				minimo: { type: 'integer', minimum: 0, description: 'Total mínimo en pesos.' },
				orden: {
					type: 'string',
					enum: ['nuevos', 'urgentes', 'monto', 'viejos'],
					default: 'nuevos'
				},
				limite: LIMITE(50, 20)
			},
			additionalProperties: false
		}
	},
	{
		name: 'ver_pedido',
		title: 'Ver un pedido',
		description:
			'El detalle de un pedido: productos con cantidades y precios, subtotal, envío, total, comprador (contacto, pedidos y gastado), estado del pago, factura, seguimiento e historial.',
		inputSchema: {
			type: 'object',
			properties: { numero: { type: 'integer', minimum: 1, description: 'Número del pedido.' } },
			required: ['numero'],
			additionalProperties: false
		}
	},
	{
		name: 'listar_productos',
		title: 'Listar o buscar productos',
		description:
			'Lista productos del catálogo con precio y costo en pesos, margen, stock, unidades vendidas en 30 días, categoría y canales donde se publica. El id sirve para ajustar_producto.',
		inputSchema: {
			type: 'object',
			properties: {
				texto: { type: 'string', description: 'Busca por nombre, SKU o categoría.' },
				estado: {
					type: 'string',
					enum: ['activo', 'sin_stock', 'pausado', 'borrador']
				},
				orden: {
					type: 'string',
					enum: ['mas_vendidos', 'menos_stock', 'mayor_precio', 'mayor_margen'],
					default: 'mas_vendidos'
				},
				limite: LIMITE(50, 20)
			},
			additionalProperties: false
		}
	},
	{
		name: 'listar_clientes',
		title: 'Listar o buscar clientes',
		description:
			'Lista clientes con ciudad, canal, teléfono, cantidad de pedidos, total gastado en pesos, días desde la última compra, segmento y etiquetas.',
		inputSchema: {
			type: 'object',
			properties: {
				texto: { type: 'string', description: 'Busca por nombre, ciudad, teléfono o etiqueta.' },
				segmento: {
					type: 'string',
					enum: ['vip', 'frecuentes', 'nuevos', 'en_riesgo', 'dormidos']
				},
				orden: {
					type: 'string',
					enum: ['mas_gastaron', 'compra_reciente', 'mas_pedidos', 'alfabetico'],
					default: 'mas_gastaron'
				},
				limite: LIMITE(50, 20)
			},
			additionalProperties: false
		}
	},
	{
		name: 'informe_de_ventas',
		title: 'Informe de ventas',
		description:
			'Ventas de un período en pesos. hoy, 7d y 30d: ventas, pedidos, ticket promedio y clientes nuevos contra el período anterior del mismo largo, reparto por canal y los más vendidos. 12m: ventas, unidades, clientes nuevos y margen de cada uno de los últimos 12 meses.',
		inputSchema: {
			type: 'object',
			properties: {
				periodo: { type: 'string', enum: ['hoy', '7d', '30d', '12m'], default: '30d' }
			},
			additionalProperties: false
		}
	},
	{
		name: 'cambiar_estado_pedido',
		title: 'Cambiar el estado de un pedido',
		description:
			'Pasa un pedido a otro estado. Sólo se puede avanzar una etapa (Sin pagar → Para preparar → Listo para despachar → En camino → Entregado), volver una o cancelarlo. Confirmá con la persona antes.',
		accion: true,
		destructiva: true,
		inputSchema: {
			type: 'object',
			properties: {
				numero: { type: 'integer', minimum: 1 },
				estado: { type: 'string', enum: ESTADOS }
			},
			required: ['numero', 'estado'],
			additionalProperties: false
		}
	},
	{
		name: 'ajustar_producto',
		title: 'Ajustar stock o precio',
		description:
			'Cambia el precio (en pesos) o el stock de un producto. Para el stock, pasá "stock" con el total nuevo o "sumar_stock" con lo que entra (negativo si sale). Usá listar_productos para el id. Confirmá con la persona antes.',
		accion: true,
		panel: 'escribir',
		inputSchema: {
			type: 'object',
			properties: {
				id: { type: 'string', description: 'Id del producto (de listar_productos).' },
				precio: { type: 'integer', minimum: 0, description: 'Precio nuevo en pesos.' },
				stock: { type: 'integer', minimum: 0, description: 'Stock total nuevo.' },
				sumar_stock: {
					type: 'integer',
					description: 'Unidades que entran (o salen, en negativo).'
				}
			},
			required: ['id'],
			additionalProperties: false
		}
	},
	{
		name: 'crear_cupon',
		title: 'Crear un cupón',
		description:
			'Crea un cupón de descuento para toda la tienda, válido en la tienda online, WhatsApp, Instagram y el local. Confirmá el código y el descuento con la persona antes.',
		accion: true,
		panel: 'plata',
		inputSchema: {
			type: 'object',
			properties: {
				codigo: {
					type: 'string',
					pattern: '^[A-Z0-9-]{3,20}$',
					description: 'En mayúsculas, de 3 a 20 letras, números o guiones. Ej.: VERANO15.'
				},
				tipo: { type: 'string', enum: ['porcentaje', 'monto_fijo', 'envio_gratis'] },
				valor: {
					type: 'integer',
					minimum: 0,
					description: 'Porcentaje (1 a 90) o pesos de descuento (desde 100). 0 en envío gratis.'
				},
				minimo: {
					type: 'integer',
					minimum: 0,
					default: 0,
					description: 'Compra mínima en pesos.'
				},
				limite_usos: {
					type: 'integer',
					minimum: 0,
					default: 0,
					description: 'Usos en total (0: sin límite).'
				},
				vigencia: {
					type: 'string',
					enum: ['7', '30', 'finde', 'sin'],
					default: '7',
					description: '7 o 30 días, este fin de semana, o sin vencimiento.'
				}
			},
			required: ['codigo', 'tipo'],
			additionalProperties: false
		}
	},
	{
		name: 'crear_link_de_pago',
		title: 'Crear un link de pago',
		description:
			'Crea un link de pago por un monto en pesos, para mandarle al cliente. Si cobra un pedido, al pagarse el pedido pasa a preparar.',
		accion: true,
		inputSchema: {
			type: 'object',
			properties: {
				monto: { type: 'integer', minimum: 1, description: 'Monto en pesos.' },
				concepto: { type: 'string', description: 'Qué se cobra (lo ve quien paga).' },
				cuotas: { type: 'boolean', default: true, description: 'Permitir pagar en cuotas.' },
				pedido: { type: 'integer', minimum: 1, description: 'Número de pedido que cobra.' }
			},
			required: ['monto', 'concepto'],
			additionalProperties: false
		}
	}
];

export function herramientasPara(permisos: Permiso[]) {
	return HERRAMIENTAS.filter((h) => !h.accion || permisos.includes('acciones')).map((h) => ({
		name: h.name,
		title: h.title,
		description: h.description,
		inputSchema: h.inputSchema,
		annotations: {
			title: h.title,
			readOnlyHint: !h.accion,
			destructiveHint: !!h.destructiva,
			idempotentHint: !h.accion,
			openWorldHint: false
		}
	}));
}

// ───── Ejecución ─────

export type Resultado = {
	content: { type: 'text'; text: string }[];
	structuredContent?: Record<string, unknown>;
	isError?: boolean;
};

class ErrorHerramienta extends Error {}
const falla = (m: string): never => {
	throw new ErrorHerramienta(m);
};
const conError = (text: string): Resultado => ({
	content: [{ type: 'text', text }],
	isError: true
});

export type Contexto = {
	datos: Datos;
	permisos: Permiso[];
	sitio: string;
	/** Los interruptores del panel, si ya se conocen (con la API real los trae la conexión). */
	permisosPanel?: Record<string, boolean>;
};
type Args = Record<string, unknown>;
type Manejador = (c: Contexto, a: Args) => Promise<Record<string, unknown>>;

/** Ejecuta una herramienta respetando los permisos concedidos y los del panel. */
export async function llamar(c: Contexto, nombre: string, args: Args): Promise<Resultado> {
	const h = HERRAMIENTAS.find((x) => x.name === nombre);
	if (!h) return conError(`No existe la herramienta "${nombre}".`);
	if (h.accion && !c.permisos.includes('acciones'))
		return conError(
			'Esta conexión es de sólo lectura. Para hacer cambios, desconectala en Tu Local Web › Integraciones y volvé a conectarla permitiendo acciones.'
		);
	try {
		if (h.panel) {
			const permisosIa = c.permisosPanel ?? (await c.datos.integraciones()).permisosIa;
			if (!permisosIa[h.panel])
				falla(
					`En Tu Local Web › Integraciones está apagado "${PERMISO_PANEL[h.panel]}". Prendelo para que la IA pueda hacer esto.`
				);
		}
		const data = await MANEJADORES[nombre](c, args ?? {});
		return {
			content: [{ type: 'text', text: JSON.stringify(data, null, 1) }],
			structuredContent: data
		};
	} catch (e) {
		if (e instanceof ErrorHerramienta) return conError(e.message);
		// Redirecciones o errores de la fuente de datos: nunca el detalle interno.
		console.error('mcp', nombre, e instanceof Error ? e.message : 'error');
		return conError('No pudimos completar el pedido. Probá de nuevo en un rato.');
	}
}

// ───── Validación de argumentos ─────

const texto = (v: unknown, max = 120) =>
	typeof v === 'string' && v.trim() ? v.trim().slice(0, max) : undefined;
const entero = (v: unknown, campo: string, min = 0, max = 1_000_000_000) => {
	if (v === undefined || v === null) return undefined;
	if (typeof v !== 'number' || !Number.isInteger(v) || v < min || v > max)
		falla(`"${campo}" tiene que ser un número entero entre ${min} y ${max}.`);
	return v as number;
};
function opcion<T extends string>(v: unknown, campo: string, valores: readonly T[]): T | undefined {
	if (v === undefined || v === null) return undefined;
	if (!valores.includes(v as T)) falla(`"${campo}" tiene que ser uno de: ${valores.join(', ')}.`);
	return v as T;
}
const monto = (n: number) => ({ pesos: n, texto: pesos(n) });
/** Un monto del tablero, que no viene si el rol no ve montos (Ventas, Local): tampoco va acá. */
const montoSi = (n: number | undefined) => (n === undefined ? undefined : monto(n));
/** Lo que se le dice a la IA cuando el rol de quien la conectó no ve montos. */
const SIN_MONTOS = { montos: 'Tu rol no ve montos de dinero: van sólo cantidades.' };
const fecha = (iso: string | null | undefined) => iso ?? null;

const pedidoBreve = (
	p: PedidoDetalle | Awaited<ReturnType<Datos['listarPedidos']>>['filas'][0]
) => ({
	numero: p.numero,
	cliente: p.cliente,
	ciudad: p.ciudad,
	canal: p.canal,
	pago: p.pago,
	envio: p.envio,
	estado: p.estado,
	total: monto(p.total),
	creado: p.creado,
	productos: p.productos.map((x) => x.nombre),
	cantidad_productos: p.cantidadProductos
});

const SEG_PRODUCTO = {
	activo: 'active',
	sin_stock: 'out',
	pausado: 'paused',
	borrador: 'draft'
} as const;
const ORDEN_PRODUCTO = {
	mas_vendidos: 'sold',
	menos_stock: 'stock',
	mayor_precio: 'price',
	mayor_margen: 'margin'
} as const;
const SEG_CLIENTE = {
	vip: 'vip',
	frecuentes: 'freq',
	nuevos: 'new',
	en_riesgo: 'risk',
	dormidos: 'sleep'
} as const;
const ORDEN_CLIENTE = {
	mas_gastaron: 'ltv',
	compra_reciente: 'last',
	mas_pedidos: 'orders',
	alfabetico: 'az'
} as const;
const nombreSeg = <T extends Record<string, string>>(m: T, v: unknown) =>
	(Object.entries(m).find(([, x]) => x === v)?.[0] ?? v) as string;

const fila = (r: Record<string, unknown>, claves: string[]) =>
	Object.fromEntries(claves.filter((k) => k in r).map((k) => [k, r[k]]));

async function informePeriodo(d: Datos, periodo: 'hoy' | '7d' | '30d') {
	const t = await d.tablero(periodo);
	return {
		periodo,
		desde: t.desde,
		hasta: t.hasta,
		...(t.montos ? {} : SIN_MONTOS),
		ventas: t.ventas && { ...monto(t.ventas.actual), periodo_anterior: monto(t.ventas.anterior) },
		pedidos: { cantidad: t.kpis.pedidos.actual, periodo_anterior: t.kpis.pedidos.anterior },
		ticket_promedio: t.kpis.ticket && {
			...monto(t.kpis.ticket.actual),
			periodo_anterior: monto(t.kpis.ticket.anterior)
		},
		clientes_nuevos: {
			cantidad: t.kpis.clientesNuevos.actual,
			periodo_anterior: t.kpis.clientesNuevos.anterior
		},
		canales: t.canales?.map((c) => ({
			canal: c.canal,
			parte_pct: Math.round(c.parte * 1000) / 10
		})),
		mas_vendidos_7d: t.masVendidos.map((m) => ({
			id: m.productoId,
			nombre: m.nombre,
			unidades: m.unidades
		}))
	};
}

const MANEJADORES: Record<string, Manejador> = {
	async resumen_del_dia({ datos }) {
		const t = await datos.tablero('hoy');
		return {
			desde: t.desde,
			...(t.montos ? {} : SIN_MONTOS),
			ventas_hoy: montoSi(t.ventas?.actual),
			ventas_ayer_mismo_tramo: montoSi(t.ventas?.anterior),
			pedidos_hoy: t.kpis.pedidos.actual,
			pedidos_ayer: t.kpis.pedidos.anterior,
			ticket_promedio: montoSi(t.kpis.ticket?.actual),
			para_hacer: {
				pedidos_para_preparar: t.paraHacer.preparar,
				chats_sin_responder: t.paraHacer.chats,
				transferencias_por_confirmar: t.paraHacer.transferencias,
				transferencias_monto: montoSi(t.paraHacer.transferenciasMonto),
				productos_sin_stock: t.paraHacer.sinStock
			},
			meta_del_mes: t.meta?.objetivo
				? {
						objetivo: monto(t.meta.objetivo),
						vendido: monto(t.meta.actual),
						proyeccion: monto(t.meta.proyeccion)
					}
				: null,
			ultimos_pedidos: t.ultimosPedidos.slice(0, 5).map((p) => ({
				numero: p.numero,
				cliente: p.cliente,
				canal: p.canal,
				estado: p.estado,
				total: montoSi(p.total),
				creado: p.creado
			}))
		};
	},

	async listar_pedidos({ datos }, a) {
		const estado = opcion(a.estado, 'estado', ESTADOS);
		const canal = texto(a.canal, 10);
		const f: FiltrosPedidos = {
			q: texto(a.texto),
			estado: estado ? [estado] : undefined,
			canal: canal ? [canal] : undefined,
			fecha: opcion(a.fecha, 'fecha', ['hoy', '7d', '30d', 'todo'] as const) ?? 'todo',
			minimo: entero(a.minimo, 'minimo'),
			orden: opcion(a.orden, 'orden', ['nuevos', 'urgentes', 'monto', 'viejos'] as const),
			cuantos: entero(a.limite, 'limite', 1, 50) ?? 20
		};
		const l = await datos.listarPedidos(f);
		return {
			total: l.total,
			suma: monto(l.suma),
			por_estado: l.etapas,
			pedidos: l.filas.map(pedidoBreve)
		};
	},

	async ver_pedido({ datos }, a) {
		const numero = entero(a.numero, 'numero', 1) ?? falla('Falta el número de pedido.');
		const p = await datos.pedido(numero);
		if (!p) falla(`No existe el pedido #${numero}.`);
		const d = p!;
		return {
			...pedidoBreve(d),
			items: d.items.map((i) => ({
				nombre: i.nombre,
				cantidad: i.cantidad,
				precio: monto(i.precio)
			})),
			subtotal: monto(d.subtotal),
			costo_envio: monto(d.costoEnvio),
			comprador: {
				nombre: d.comprador.nombre,
				telefono: d.comprador.telefono,
				email: d.comprador.email,
				cliente_desde: d.comprador.desde,
				pedidos: d.comprador.pedidos,
				gastado: monto(d.comprador.gastado)
			},
			estado_pago: d.estadoPago,
			factura: d.factura,
			seguimiento: d.seguimiento,
			llega: d.llega,
			nota: d.nota,
			historial: d.historial.map((h) => ({
				texto: h.texto,
				cuando: fecha(h.cuando),
				hecho: h.hecho
			}))
		};
	},

	async listar_productos({ datos }, a) {
		const seg = opcion(
			a.estado,
			'estado',
			Object.keys(SEG_PRODUCTO) as (keyof typeof SEG_PRODUCTO)[]
		);
		const orden =
			opcion(a.orden, 'orden', Object.keys(ORDEN_PRODUCTO) as (keyof typeof ORDEN_PRODUCTO)[]) ??
			'mas_vendidos';
		const f: FiltrosLista = {
			q: texto(a.texto),
			seg: seg ? [SEG_PRODUCTO[seg]] : undefined,
			orden: ORDEN_PRODUCTO[orden],
			cuantos: entero(a.limite, 'limite', 1, 50) ?? 20
		};
		const l = await datos.listar('productos', f);
		return {
			total: l.total,
			productos: l.filas.map((r) => {
				const p = fila(r, ['id', 'nombre', 'sku', 'categoria', 'stock', 'margen', 'canales']);
				return {
					...p,
					precio: monto(Number(r.precio ?? 0)),
					costo: monto(Number(r.costo ?? 0)),
					vendidos_30d: r.vendidos30 ?? r.vendidos30d ?? null,
					estado: nombreSeg(SEG_PRODUCTO, r.estado)
				};
			})
		};
	},

	async listar_clientes({ datos }, a) {
		const seg = opcion(
			a.segmento,
			'segmento',
			Object.keys(SEG_CLIENTE) as (keyof typeof SEG_CLIENTE)[]
		);
		const orden =
			opcion(a.orden, 'orden', Object.keys(ORDEN_CLIENTE) as (keyof typeof ORDEN_CLIENTE)[]) ??
			'mas_gastaron';
		const l = await datos.listar('clientes', {
			q: texto(a.texto),
			seg: seg ? [SEG_CLIENTE[seg]] : undefined,
			orden: ORDEN_CLIENTE[orden],
			cuantos: entero(a.limite, 'limite', 1, 50) ?? 20
		});
		return {
			total: l.total,
			clientes: l.filas.map((r) => ({
				...fila(r, ['id', 'nombre', 'ciudad', 'canal', 'telefono', 'pedidos', 'etiquetas']),
				gastado: monto(Number(r.gastado ?? 0)),
				dias_desde_la_ultima_compra: r.ultima ?? null,
				segmento: nombreSeg(SEG_CLIENTE, r.seg)
			}))
		};
	},

	async informe_de_ventas({ datos }, a) {
		const periodo = opcion(a.periodo, 'periodo', ['hoy', '7d', '30d', '12m'] as const) ?? '30d';
		if (periodo !== '12m') return informePeriodo(datos, periodo);
		const i = await datos.informes();
		return {
			periodo,
			meses: i.meses.map((mes, k) => ({
				mes,
				ventas: monto(i.series.ventas[k] ?? 0),
				unidades: i.series.productos[k] ?? 0,
				clientes_nuevos: i.series.clientes[k] ?? 0,
				margen_pct: i.series.margen[k] ?? null
			})),
			mes_actual: {
				ventas: monto(i.kpis.ventas),
				variacion_pct: i.kpis.ventasVs,
				ticket_promedio: monto(i.kpis.ticket),
				margen_pct: i.kpis.margen,
				mes_anterior: i.kpis.mesAnterior
			},
			observaciones: i.insights
		};
	},

	async cambiar_estado_pedido({ datos }, a) {
		const numero = entero(a.numero, 'numero', 1) ?? falla('Falta el número de pedido.');
		const estado = opcion(a.estado, 'estado', ESTADOS) ?? falla('Falta el estado nuevo.');
		const antes = await datos.pedido(numero);
		if (!antes) falla(`No existe el pedido #${numero}.`);
		const r = await datos.cambiarEstado(numero, estado);
		if (r === 'no-existe') falla(`No existe el pedido #${numero}.`);
		if (r === 'no-se-puede')
			falla(
				`El pedido #${numero} está "${antes!.estado}": no puede pasar a "${estado}". Sólo avanza una etapa, vuelve una o se cancela.`
			);
		return { numero, antes: antes!.estado, ahora: estado };
	},

	async ajustar_producto({ datos }, a) {
		const id = texto(a.id, 80) ?? falla('Falta el id del producto (usá listar_productos).');
		const precio = entero(a.precio, 'precio', 0, 100_000_000);
		const stock = entero(a.stock, 'stock', 0, 1_000_000);
		const sumar = entero(a.sumar_stock, 'sumar_stock', -1_000_000, 1_000_000);
		if (precio === undefined && stock === undefined && sumar === undefined)
			falla('Decime el precio nuevo, el stock nuevo o cuánto sumar al stock.');
		if (stock !== undefined && sumar !== undefined)
			falla('Pasá "stock" o "sumar_stock", no los dos.');
		const p = await datos.producto(id);
		if (!p) falla(`No existe el producto "${id}". Buscalo con listar_productos.`);
		const antes = p!;
		if ((stock !== undefined || sumar !== undefined) && antes.opciones?.length)
			falla(
				`"${antes.nombre}" tiene variantes (${antes.opciones.map((o) => o.nombre.toLowerCase()).join(', ')}): su stock se ajusta por variante en Inventario.`
			);
		const nuevoStock = sumar !== undefined ? (antes.stock ?? 0) + sumar : stock;
		if (nuevoStock !== undefined && nuevoStock < 0)
			falla(`El stock no puede quedar negativo (hay ${antes.stock ?? 0}).`);
		const g = await datos.guardarProducto(id, {
			nombre: antes.nombre,
			precio: precio ?? antes.precio,
			...(nuevoStock !== undefined ? { stock: nuevoStock } : {})
		});
		if (!g) falla(`No existe el producto "${id}".`);
		return {
			id,
			nombre: antes.nombre,
			precio: { antes: monto(antes.precio), ahora: monto(g!.precio) },
			stock: { antes: antes.stock ?? 0, ahora: g!.stock ?? 0 }
		};
	},

	async crear_cupon({ datos }, a) {
		// Sin recortar: un código de más de 20 lo rechaza la regex, no se acorta en silencio.
		const codigo = (typeof a.codigo === 'string' ? a.codigo.trim() : '').toUpperCase();
		if (!/^[A-Z0-9-]{3,20}$/.test(codigo))
			falla('El código va en mayúsculas, de 3 a 20 letras, números o guiones.');
		const tipo =
			opcion(a.tipo, 'tipo', ['porcentaje', 'monto_fijo', 'envio_gratis'] as const) ??
			falla('Falta el tipo de descuento.');
		const valor = tipo === 'envio_gratis' ? 0 : (entero(a.valor, 'valor', 0, 100_000) ?? 0);
		if (tipo === 'porcentaje' && (valor < 1 || valor > 90)) falla('El porcentaje va de 1 a 90.');
		if (tipo === 'monto_fijo' && valor < 100) falla('El descuento fijo es de $ 100 para arriba.');
		const n: CuponNuevo = {
			codigo,
			tipo: tipo === 'porcentaje' ? 'pct' : tipo === 'monto_fijo' ? 'fijo' : 'envio',
			valor,
			alcance: 'todo',
			quien: 'todos',
			minimo: entero(a.minimo, 'minimo', 0, 10_000_000) ?? 0,
			limite: entero(a.limite_usos, 'limite_usos', 0, 100_000) ?? 0,
			porCliente: '1',
			vigencia: opcion(a.vigencia, 'vigencia', ['7', '30', 'finde', 'sin'] as const) ?? '7',
			canales: ['web', 'wa', 'ig', 'pos'],
			opciones: { hornero: false, combinable: false, unicos: false, link: true },
			programar: false
		};
		const c = await datos.crearCupon(n);
		if (c === 'repetido') falla(`Ya hay un cupón ${codigo}. Elegí otro código.`);
		const cupon = c as Exclude<typeof c, 'repetido'>;
		return {
			id: cupon.id,
			codigo: cupon.codigo,
			descripcion: cupon.descripcion,
			estado: cupon.estado,
			minimo: monto(cupon.minimo),
			limite_usos: cupon.limite,
			termina: cupon.termina
		};
	},

	async crear_link_de_pago({ datos, sitio }, a) {
		const m = entero(a.monto, 'monto', 1, 100_000_000) ?? falla('Falta el monto en pesos.');
		const concepto = texto(a.concepto, 140);
		if (!concepto || concepto.length < 2) falla('Escribí qué se cobra.');
		const pedido = entero(a.pedido, 'pedido', 1) ?? null;
		if (pedido && !(await datos.pedido(pedido))) falla(`No existe el pedido #${pedido}.`);
		const codigo = await datos.crearLink({
			monto: m,
			concepto: concepto!,
			cuotas: a.cuotas !== false,
			pedido
		});
		return { codigo, url: `${sitio}/p/${codigo}`, monto: monto(m), concepto, pedido };
	}
};
