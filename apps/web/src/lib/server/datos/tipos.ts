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
}
