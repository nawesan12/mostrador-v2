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

export interface Datos {
	rubros(): Promise<Rubro[]>;
	/** La tienda de la sesión. */
	tienda(): Promise<Tienda>;
	/** Lo elegido en el onboarding (y si terminó). */
	guardarOnboarding(o: Onboarding): Promise<void>;
	/** Insignias del menú, notificaciones, canales y prueba: lo que muestra el armazón del panel. */
	resumen(): Promise<ResumenPanel>;
	marcarNotificacionesLeidas(): Promise<void>;
}
