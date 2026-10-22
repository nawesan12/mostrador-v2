/** Qué puede hacer cada rol (design/PanelCuenta.dc.html). El dueño puede todo; "Plan, usuarios y datos" sólo el dueño. */
export const ROLES = ['duenio', 'admin', 'ventas', 'deposito'] as const;
export type Rol = (typeof ROLES)[number];
export const NOMBRE_ROL: Record<Rol, string> = {
	duenio: 'Dueño',
	admin: 'Admin',
	ventas: 'Ventas',
	deposito: 'Depósito'
};

export const PERMISOS = [
	{
		id: 'ventas',
		titulo: 'Ver ventas y pedidos',
		detalle: 'Lista de pedidos y clientes',
		def: [1, 1, 1, 1]
	},
	{
		id: 'productos',
		titulo: 'Editar productos y precios',
		detalle: 'Incluye stock e importaciones',
		def: [1, 1, 0, 1]
	},
	{
		id: 'whatsapp',
		titulo: 'Responder WhatsApp',
		detalle: 'Y aprobar respuestas de Hornero',
		def: [1, 1, 1, 0]
	},
	{
		id: 'despachar',
		titulo: 'Preparar y despachar',
		detalle: 'Etiquetas y cambios de estado',
		def: [1, 1, 1, 1]
	},
	{
		id: 'finanzas',
		titulo: 'Ver finanzas',
		detalle: 'Saldo, liquidaciones y Caja',
		def: [1, 1, 0, 0]
	},
	{
		id: 'facturas',
		titulo: 'Emitir facturas',
		detalle: 'Factura A, B y C ante ARCA',
		def: [1, 1, 1, 0]
	},
	{
		id: 'diseno',
		titulo: 'Cambiar diseño de la tienda',
		detalle: 'Plantilla, colores y secciones',
		def: [1, 1, 0, 0]
	},
	{
		id: 'cuenta',
		titulo: 'Plan, usuarios y datos',
		detalle: 'Solo el dueño puede cambiarlo',
		def: [1, 0, 0, 0]
	}
] as const;

/** Lo que no se puede tocar: todo lo del dueño y "Plan, usuarios y datos". */
export const fijo = (permiso: string, rol: Rol) => rol === 'duenio' || permiso === 'cuenta';

export const AREAS: Record<Rol, string[]> = {
	duenio: ['Todo'],
	admin: ['Ventas', 'Catálogo', 'Finanzas'],
	ventas: ['Pedidos', 'WhatsApp'],
	deposito: ['Stock', 'Envíos']
};

/** Los motivos y la oferta de cada uno en la cancelación en 3 pasos. */
export const MOTIVOS = [
	'Es caro para lo que vendo',
	'Me falta una función',
	'Voy a pausar el negocio',
	'Me voy a otra plataforma',
	'Otro motivo'
] as const;
