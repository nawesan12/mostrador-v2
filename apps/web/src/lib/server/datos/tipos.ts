/**
 * Lo que las páginas pueden pedir. Las dos fuentes (sandbox de la demo y API Go) implementan esta
 * misma interfaz con los tipos del contrato (packages/contrato/openapi.yaml), y los casos de
 * packages/contrato/casos corren contra las dos.
 */
import type { components } from '@mostrador/contrato/esquema';

export type Rubro = components['schemas']['Rubro'];
export type Tienda = components['schemas']['Tienda'];
export type Plan = components['schemas']['Plan'];

export type Onboarding = components['schemas']['Onboarding'];
export type ResumenPanel = components['schemas']['ResumenPanel'];
export type Notificacion = components['schemas']['Notificacion'];
export type Tablero = components['schemas']['Tablero'];
export type Periodo = Tablero['periodo'];
export type ContextoHornero = components['schemas']['ContextoHornero'];
export type Idea = 'carritos' | 'stock' | 'promo';
export type ListaPedidos = components['schemas']['ListaPedidos'];
export type PedidoDetalle = components['schemas']['PedidoDetalle'];
export type EstadoPedido = components['schemas']['EstadoPedido'];
export type AccionLote = 'preparar' | 'etiquetas' | 'facturar' | 'avisar';
export type FiltrosPedidos = import('../sandbox/pedidos.ts').FiltrosPedidos;
/** Resultado de una acción sobre un pedido. */
export type Resultado = 'ok' | 'no-existe' | 'no-se-puede';
export type ListaGenerica = components['schemas']['ListaGenerica'];
export type NombreLista = components['schemas']['NombreLista'];
export type Producto = components['schemas']['Producto'];
export type ProductoEditable = components['schemas']['ProductoEditable'];
export type FiltrosLista = import('../sandbox/listas.ts').FiltrosLista;
export type MovimientoStock = components['schemas']['MovimientoStock'];
export type ConversacionBreve = components['schemas']['ConversacionBreve'];
export type Conversacion = components['schemas']['Conversacion'];
export type Carritos = components['schemas']['Carritos'];
export type AutomaticosCarritos = components['schemas']['AutomaticosCarritos'];
export type PedidoNuevo = components['schemas']['PedidoNuevo'];
export type ListasPrecio = components['schemas']['ListasPrecio'];
export type Cupon = components['schemas']['Cupon'];
export type Cupones = components['schemas']['Cupones'];
export type CuponNuevo = components['schemas']['CuponNuevo'];
export type AccionCupon = 'pausar' | 'activar' | 'duplicar' | 'minimo' | 'deshacer-minimo';
export type AccionRelampago = 'sumar-hora' | 'liberar' | 'terminar' | 'reanudar';
export type Devolucion = components['schemas']['Devolucion'];
export type ResenasYPreguntas = components['schemas']['ResenasYPreguntas'];

export interface Datos {
	rubros(): Promise<Rubro[]>;
	/** La tienda de la sesión. */
	tienda(): Promise<Tienda>;
	/** Lo elegido en el onboarding (y si terminó). */
	guardarOnboarding(o: Onboarding): Promise<void>;
	/** Insignias del menú, notificaciones, canales y prueba: lo que muestra el armazón del panel. */
	resumen(): Promise<ResumenPanel>;
	marcarNotificacionesLeidas(): Promise<void>;
	/** El Resumen (dashboard) del período. */
	tablero(periodo: Periodo): Promise<Tablero>;
	/** La meta de ventas del mes, en pesos (0 la saca). */
	guardarMeta(objetivo: number): Promise<void>;
	ocultarAviso(aviso: 'certificado'): Promise<void>;
	/** Manda el recordatorio a los carritos abandonados; devuelve cuántos mensajes salieron. */
	recordarCarritos(): Promise<number>;
	/** Lo que Hornero sabe de la tienda para responder. */
	contextoHornero(): Promise<ContextoHornero>;
	/** Aprueba (o deshace) una idea de Hornero. */
	marcarIdea(idea: Idea, hecha: boolean): Promise<void>;
	listarPedidos(f: FiltrosPedidos): Promise<ListaPedidos>;
	/** null si no existe. */
	pedido(numero: number): Promise<PedidoDetalle | null>;
	cambiarEstado(numero: number, estado: EstadoPedido): Promise<Resultado>;
	pedidosEnLote(
		numeros: number[],
		accion: AccionLote
	): Promise<{ hechos: number; cambios: { numero: number; antes: EstadoPedido }[] }>;
	/** Recordatorio de pago por WhatsApp. */
	recordarPago(numero: number): Promise<Resultado>;
	/** Motor de listas: Clientes, Productos, Inventario, Facturación y Envíos. */
	listar(lista: NombreLista, f: FiltrosLista): Promise<ListaGenerica>;
	/** Una acción sobre filas de una lista (o la sugerencia de Hornero); devuelve cuántas cambiaron. */
	accionLista(lista: NombreLista, ids: string[], accion: string, valor?: unknown): Promise<number>;
	producto(id: string): Promise<Producto | null>;
	/** Crea (id null) o guarda un producto. null si el id no existe. */
	guardarProducto(id: string | null, p: ProductoEditable): Promise<Producto | null>;
	/** Movimientos de stock de hoy (ventas, ajustes, conteos, ingresos). */
	movimientosHoy(): Promise<MovimientoStock[]>;
	bandeja(): Promise<{ conversaciones: ConversacionBreve[]; atendidasHoy: number }>;
	conversacion(id: string): Promise<Conversacion | null>;
	/** Mandar un mensaje, tomar o devolver la conversación, o marcarla leída. */
	responder(
		id: string,
		c: { texto?: string; tomada?: boolean; leida?: boolean }
	): Promise<Resultado>;
	carritos(): Promise<Carritos>;
	guardarAutomaticos(a: AutomaticosCarritos): Promise<void>;
	recordarCarrito(id: string): Promise<Resultado>;
	/** Un pedido nuevo (punto de venta o cargado a mano); descuenta stock. */
	crearPedido(
		n: PedidoNuevo
	): Promise<{ numero: number; total: number } | 'sin-stock' | 'invalido'>;
	devoluciones(): Promise<{ devoluciones: Devolucion[]; tasa: number }>;
	resolverDevolucion(id: string, accion: 'aprobar' | 'cambio' | 'deshacer'): Promise<Resultado>;
	resenas(): Promise<ResenasYPreguntas>;
	responderPregunta(id: string, respuesta: string): Promise<Resultado>;
	listasPrecio(): Promise<ListasPrecio>;
	cupones(): Promise<Cupones>;
	crearCupon(n: CuponNuevo): Promise<Cupon | 'repetido'>;
	accionCupon(
		id: string,
		accion: AccionCupon,
		minimo?: number
	): Promise<Cupon | 'no-existe' | 'no-se-puede'>;
	descartarConsejo(): Promise<void>;
	fijarAutomatica(id: string, activa: boolean): Promise<Resultado>;
	accionRelampago(accion: AccionRelampago): Promise<Resultado>;
	/** Sube todos los precios ya, o lo programa para el lunes. */
	aumentarPrecios(
		pct: number,
		redondeo: 'no' | '90' | '500',
		programar: boolean
	): Promise<{ productos: number; para: string | null }>;
	deshacerAumento(): Promise<Resultado>;
	cancelarAumento(): Promise<Resultado>;
	/** Un link de pago; devuelve el código. */
	crearLink(l: {
		monto: number;
		concepto: string;
		cuotas: boolean;
		pedido: number | null;
	}): Promise<string>;
}
