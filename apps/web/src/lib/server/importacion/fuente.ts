/**
 * La fuente de una importación, como en Go (apps/api/internal/dominio/importacion/fuente.go): lee la
 * tienda vieja de a una página por etapa. La implementa cada plataforma; en la demo la usa
 * `DatosSandbox.avanzarImportacion`, un paso por pedido del navegador (no hay outbox).
 */
import type { components } from '@mostrador/contrato/esquema';
import type { ProductoEditable } from '../datos/tipos.ts';

export type Plataforma = components['schemas']['Plataforma'];
export type Importacion = components['schemas']['Importacion'];
export type ImportacionNueva = components['schemas']['ImportacionNueva'];
export type PedidoHistorial = components['schemas']['PedidoHistorial'];
export type Progreso = components['schemas']['ProgresoImportacion'];
export type ErrorImportacion = components['schemas']['ErrorImportacion'];
export type Etapa = Importacion['etapa'];

export const PLATAFORMAS: Plataforma[] = [
	'tiendanube',
	'shopify',
	'woocommerce',
	'mercadolibre',
	'empretienda',
	'csv'
];
/** En orden: los productos primero (los pedidos viejos los nombran). */
export const ETAPAS: Etapa[] = ['productos', 'clientes', 'pedidos', 'redirecciones'];
export const MAX_ERRORES = 50;
/** Corta una fuente que nunca termina (un cursor que se repite). */
export const MAX_PASOS = 20_000;

export type ProductoViejo = { externo: string; url: string; producto: ProductoEditable };
export type ClienteViejo = {
	externo: string;
	nombre: string;
	email?: string;
	telefono?: string;
	ciudad?: string;
	provincia?: string;
	notas?: string;
};
export type ItemViejo = { nombre: string; variante: string; cantidad: number; precio: number };
export type PedidoViejo = {
	externo: string;
	numero: string;
	/** ms desde 1970. */
	fecha: number;
	email?: string;
	nombre?: string;
	/** En pesos enteros. */
	total: number;
	items: ItemViejo[];
	estado: string;
};
/** Una dirección vieja que no es de un producto y adónde va (path relativo a la tienda). */
export type UrlVieja = { desde: string; hacia: string };

/** Lo de la etapa pedida; `siguiente` null: la etapa terminó. */
export type Pagina = {
	productos?: ProductoViejo[];
	clientes?: ClienteViejo[];
	pedidos?: PedidoViejo[];
	redirecciones?: UrlVieja[];
	siguiente: unknown;
};

/** Error que no se arregla reintentando (la credencial ya no sirve): corta la importación. */
export class ErrorDefinitivo extends Error {}

export interface Fuente {
	/** Una página de la etapa desde el cursor (null: desde el principio). */
	pagina(etapa: Etapa, cursor: unknown): Promise<Pagina>;
}

export type Origen = { plataforma: Plataforma; url: string };
export type Abrir = (o: Origen) => Fuente | Promise<Fuente>;

/**
 * Las fuentes que hay, por plataforma. Cada plataforma se suma acá cuando esté; mientras falte,
 * crear una importación desde ella responde 'sin-fuente'.
 */
export const FUENTES: Partial<Record<Plataforma, Abrir>> = {};

export const progresoVacio = (): Progreso => ({
	productos: 0,
	clientes: 0,
	pedidos: 0,
	redirecciones: 0,
	repetidos: 0,
	recortados: 0
});
