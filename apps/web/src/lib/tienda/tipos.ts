import type { components } from '@mostrador/contrato/esquema';

/** Los turnos de la tienda (servicios y profesionales activos, seña, horario); null si no toma. */
export type TurnosTienda = components['schemas']['TurnosPublicos'];
export type TurnoPublico = components['schemas']['TurnoPublico'];
export type HorarioLibre = components['schemas']['HorarioLibre'];

/**
 * La tienda online tal como la ve quien compra: lo publicado, sin costos, márgenes ni datos de
 * clientes. La arma el servidor (demo o API) y la dibuja la plantilla elegida en Mi tienda.
 */
export type ProductoPublico = {
	id: string;
	nombre: string;
	descripcion: string;
	categoria: string | null;
	precio: number;
	/** Precio tachado (oferta). */
	anterior: number | null;
	/** URL de la foto (o id de Unsplash en la demo). */
	foto: string | null;
	etiqueta: string | null;
	stock: number;
};

export type TiendaPublica = {
	slug: string;
	nombre: string;
	plantilla: string;
	/** 0–3: la paleta elegida en Mi tienda. */
	paleta: number;
	/** Secciones visibles de la portada, en orden. */
	secciones: ('hero' | 'cats' | 'prods' | 'revs' | 'ig' | 'wa')[];
	portada: { titulo: string; lema: string; foto: string | null };
	ciudad: string;
	direccion: string;
	whatsapp: string;
	categorias: { nombre: string; cantidad: number; foto: string | null }[];
	productos: ProductoPublico[];
	/** Ids de los más vendidos, del más al menos. */
	masVendidos: string[];
	resenas: { texto: string; autor: string; producto: string; estrellas: number }[];
	promedio: number | null;
	fotos: string[];
	cobros: { transferencia: number; cuotas: number; mercadoPago: boolean };
	envios: { gratisDesde: number | null; retiro: boolean; moto: boolean; correo: boolean };
	/** Turnos con seña; null si la tienda no los tomó. `sena` es el % que se cobra de verdad. */
	turnos: TurnosTienda | null;
	/** Sólo mirar: la vitrina de una plantilla (nadie la administra, no se compra). */
	vitrina: boolean;
};

/** Lo que muestra la portada de los turnos: los libres de hoy y el próximo libre de cada uno. */
export type ProximosTurnos = {
	/** Inicios (ISO −03:00) libres hoy, con cualquiera, para el primer servicio. */
	hoy: string[];
	/** Por profesional: su próximo horario libre (ISO) o null. */
	porProfesional: Record<string, string | null>;
};

/** Una opción de envío cotizada para un código postal. */
export type OpcionEnvio = {
	id: 'andreani' | 'correo' | 'moto' | 'retiro';
	nombre: string;
	detalle: string;
	precio: number;
};
