/**
 * La dirección de la tienda (tulocalweb.com.ar/t/<slug>), editable desde el onboarding y desde
 * Configuración › Negocio (decisiones §38). La regla es la de Go (tiendas.SlugValido) y los
 * reservados salen del contrato (packages/contrato/direccion.json), igual que en Go.
 */
import CONTRATO from '@mostrador/contrato/direccion.json';
import { SITE } from '#lib/site.ts';

export const SLUG_VALIDO = /^[a-z0-9][a-z0-9-]{0,58}[a-z0-9]$/;
export const RESERVADOS: readonly string[] = CONTRATO.reservados;
export const MAX_CAMBIOS_POR_DIA = 5;
export const HOST = SITE.replace(/^https?:\/\//, '');

export type MotivoDireccion = 'libre' | 'actual' | 'propio' | 'invalido' | 'reservado' | 'usado';
export type Disponibilidad = {
	slug: string;
	disponible: boolean;
	motivo: MotivoDireccion;
	mensaje: string;
};

export const MENSAJES: Record<MotivoDireccion, string> = {
	libre: 'Está libre.',
	actual: 'Es la dirección que tenés ahora.',
	propio: 'Es una dirección que ya usaste: podés volver a ella.',
	invalido:
		'Usá de 2 a 60 letras minúsculas, números o guiones (sin guion al principio ni al final).',
	reservado: 'Esa dirección está reservada. Probá con otra.',
	usado: 'Esa dirección ya la usa otra tienda. Probá con otra.'
};

/** Lo que se puede decir sin preguntarle a nadie: formato y reservados (null si pasa). */
export function motivoDeFormato(slug: string): 'invalido' | 'reservado' | null {
	if (!SLUG_VALIDO.test(slug)) return 'invalido';
	if (RESERVADOS.includes(slug)) return 'reservado';
	return null;
}

/**
 * Lo que se escribe en el campo de la dirección, como slug: minúsculas, sin tildes, espacios y
 * signos como guion, sin guiones repetidos ni al principio. El guion del final queda mientras se
 * escribe (la validación lo marca); `slugFinal` lo saca.
 */
export function normalizarSlug(s: string): string {
	return s
		.normalize('NFD')
		.replace(/[\u0300-\u036f]/g, '')
		.toLowerCase()
		.replace(/&/g, 'y')
		.replace(/[^a-z0-9-]+/g, '-')
		.replace(/-{2,}/g, '-')
		.replace(/^-+/, '')
		.slice(0, 60);
}

export const slugFinal = (s: string) => normalizarSlug(s).replace(/-+$/, '');

/** La dirección que propone el nombre: «Café de Ana» → «cafe-de-ana». */
export const slugDeNombre = (nombre: string) => slugFinal(nombre).slice(0, 48).replace(/-+$/, '');

/** El nombre que le pone el registro con Google, Apple o Mercado Pago («La tienda de Ana»). */
export const esNombreAutomatico = (nombre: string) => /^La tienda de \S+$/.test(nombre.trim());
