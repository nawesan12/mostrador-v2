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
	/** Con opciones, la suma del stock de sus variantes. */
	stock: number;
	/** Galería (la primera es `foto`). */
	fotos: string[];
	/** Talle, Color, Presentación… (vacío: se vende sin elegir nada). */
	opciones: OpcionProducto[];
	/** Sólo las activas. `precio` null: el del producto. */
	variantes: VariantePublica[];
};

export type OpcionProducto = { nombre: string; valores: string[] };

export type VariantePublica = {
	id: string;
	valores: Record<string, string>;
	precio: number | null;
	stock: number;
	foto: string | null;
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
