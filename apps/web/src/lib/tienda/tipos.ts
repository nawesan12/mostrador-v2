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

export type ResenaPublica = {
	texto: string;
	autor: string;
	/** El nombre del producto ('' si es de la tienda en general). */
	producto: string;
	estrellas: number;
	/** La dejó quien compró (con el código de un pedido cobrado). */
	conCompra: boolean;
	/** La respuesta pública de la tienda. */
	respuesta: string | null;
};

export type PreguntaPublica = { productoId: string; pregunta: string; respuesta: string };

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
	/** Las visibles, de la más nueva a la más vieja. */
	resenas: ResenaPublica[];
	promedio: number | null;
	/** Preguntas ya respondidas por la tienda (las demás no se muestran). */
	preguntas: PreguntaPublica[];
	fotos: string[];
	cobros: { transferencia: number; cuotas: number; mercadoPago: boolean };
	envios: { gratisDesde: number | null; retiro: boolean; moto: boolean; correo: boolean };
	/** Sólo mirar: la vitrina de una plantilla (nadie la administra, no se compra). */
	vitrina: boolean;
};

/** Una opción de envío cotizada para un código postal. */
export type OpcionEnvio = {
	id: 'andreani' | 'correo' | 'moto' | 'retiro';
	nombre: string;
	detalle: string;
	precio: number;
};
