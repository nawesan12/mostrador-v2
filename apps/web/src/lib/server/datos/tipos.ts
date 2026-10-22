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
}
