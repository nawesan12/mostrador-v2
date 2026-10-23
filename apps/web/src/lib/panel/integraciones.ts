/**
 * Las integraciones del panel (design/PanelOperacion.dc.html, sInt). Cada una dice qué sincroniza y
 * qué opciones tiene en "Configurar" (el handoff no la diseña: se arma con sus piezas).
 */
export type Categoria = 'canales' | 'cobros' | 'envios' | 'gestion';

export type Integracion = {
	id: string;
	nombre: string;
	detalle: string;
	categoria: Categoria;
	/** Clave de lib/marcas.ts para el ícono, o un logo propio. */
	marca: string;
	/** Lo que hace una vez conectada. */
	hace: string[];
	opciones: { id: string; t: string; d: string; def: boolean }[];
	/** Conectada de entrada en la tienda de ejemplo. */
	demo: boolean;
	/** Todavía no se puede conectar. */
	pronto?: boolean;
};

export const INTEGRACIONES: Integracion[] = [
	{
		id: 'mp',
		nombre: 'Mercado Pago',
		detalle: 'Checkout, QR y Point',
		categoria: 'cobros',
		marca: 'mp',
		hace: [
			'Cobra en la tienda, por link y con QR en el local',
			'Las ventas se acreditan en tu cuenta',
			'Concilia solo cada pago con su pedido'
		],
		opciones: [
			{ id: 'cuotas', t: 'Cuotas sin interés', d: '3 y 6 cuotas desde $ 30.000', def: true },
			{ id: 'qr', t: 'QR en el local', d: 'Para cobrar en el punto de venta', def: true }
		],
		demo: true
	},
	{
		id: 'ml',
		nombre: 'Mercado Libre',
		detalle: 'Publicaciones y stock',
		categoria: 'canales',
		marca: 'ml',
		hace: [
			'Publica los productos que elijas',
			'El stock es uno solo con la tienda y el local',
			'Las preguntas llegan con la respuesta de Hornero'
		],
		opciones: [
			{
				id: 'precio',
				t: 'Usar la lista Mercado Libre',
				d: '+10 % para cubrir la comisión',
				def: true
			},
			{ id: 'pausar', t: 'Pausar sin stock', d: 'Las publicaciones se pausan solas', def: true }
		],
		demo: true
	},
	{
		id: 'wa',
		nombre: 'WhatsApp Business',
		detalle: 'Bandeja y campañas',
		categoria: 'canales',
		marca: 'wa',
		hace: [
			'Los chats llegan a la bandeja del panel',
			'Hornero responde y manda links de pago',
			'Campañas con el costo exacto antes de enviar'
		],
		opciones: [
			{ id: 'hornero', t: 'Hornero responde', d: 'Si está apagado, sólo sugiere', def: true },
			{ id: 'horario', t: 'Sólo fuera de horario', d: 'En horario atiende tu equipo', def: false }
		],
		demo: true
	},
	{
		id: 'ig',
		nombre: 'Instagram',
		detalle: 'Shopping y mensajes',
		categoria: 'canales',
		marca: 'ig',
		hace: ['Etiquetá productos en posts e historias', 'Los mensajes directos llegan a la bandeja'],
		opciones: [
			{
				id: 'catalogo',
				t: 'Sincronizar el catálogo',
				d: 'Cada vez que cambia un producto',
				def: true
			}
		],
		demo: true
	},
	{
		id: 'tg',
		nombre: 'Telegram',
		detalle: 'Bot de pedidos',
		categoria: 'canales',
		marca: 'tg',
		hace: ['Un bot con tu catálogo que toma pedidos', 'Los chats llegan a la bandeja'],
		opciones: [{ id: 'hornero', t: 'Hornero responde', d: 'Con el tono de tu tienda', def: true }],
		demo: true
	},
	{
		id: 'arca',
		nombre: 'ARCA',
		detalle: 'Factura electrónica',
		categoria: 'gestion',
		marca: 'arca',
		hace: [
			'Factura A, B o C sola en cada venta',
			'Notas de crédito al aprobar devoluciones',
			'Te avisa antes de que venza el certificado'
		],
		opciones: [
			{ id: 'auto', t: 'Facturar al cobrar', d: 'Si está apagado, facturás a mano', def: true }
		],
		demo: true
	},
	{
		id: 'andreani',
		nombre: 'Andreani',
		detalle: 'Etiquetas y seguimiento',
		categoria: 'envios',
		marca: 'andreani',
		hace: [
			'Cotiza el envío por código postal en el checkout',
			'Imprime las etiquetas de todos los pedidos',
			'El cliente recibe el seguimiento'
		],
		opciones: [
			{ id: 'retiro', t: 'Retiro en el local', d: 'Andreani pasa a buscar los paquetes', def: true }
		],
		demo: true
	},
	{
		id: 'modo',
		nombre: 'MODO',
		detalle: 'Apps bancarias',
		categoria: 'cobros',
		marca: 'modo',
		hace: ['Cobra con las apps de los bancos', 'Promociones bancarias en el checkout'],
		opciones: [],
		demo: true
	},
	{
		id: 'payway',
		nombre: 'Payway',
		detalle: 'Tarjetas y cuotas',
		categoria: 'cobros',
		marca: 'payway',
		hace: ['Tarjetas de crédito y débito', 'Cuotas con las promos de cada banco'],
		opciones: [],
		demo: false
	},
	{
		id: 'uala',
		nombre: 'Ualá Bis',
		detalle: 'Cobros y link de pago',
		categoria: 'cobros',
		marca: 'uala',
		hace: ['Cobra con tarjeta y link de pago de Ualá'],
		opciones: [],
		demo: false
	},
	{
		id: 'oca',
		nombre: 'OCA',
		detalle: 'Envíos a todo el país',
		categoria: 'envios',
		marca: 'oca',
		hace: ['Cotiza y genera etiquetas de OCA', 'Retiro en sucursal o a domicilio'],
		opciones: [],
		demo: false
	},
	{
		id: 'correo',
		nombre: 'Correo Argentino',
		detalle: 'Sucursal y domicilio',
		categoria: 'envios',
		marca: 'correo',
		hace: ['Cotiza y genera etiquetas de Correo Argentino'],
		opciones: [],
		demo: false
	},
	{
		id: 'caja',
		nombre: 'Caja',
		detalle: 'Tu plata, ordenada',
		categoria: 'gestion',
		marca: 'caja',
		hace: [
			'Ventas, liquidaciones y gastos pasan solos a Caja',
			'El monotributo y la ganancia del mes, al día'
		],
		opciones: [],
		demo: true
	},
	{
		id: 'ga',
		nombre: 'Google Analytics',
		detalle: 'Métricas',
		categoria: 'gestion',
		marca: 'ga',
		hace: [
			'Visitas y conversiones de la tienda en GA4',
			'Sólo con el consentimiento de cookies del visitante'
		],
		opciones: [],
		demo: false
	},
	{
		id: 'gm',
		nombre: 'Google Shopping',
		detalle: 'Merchant Center',
		categoria: 'canales',
		marca: 'gg',
		hace: ['Tus productos en Google gratis', 'Precio y stock al día'],
		opciones: [],
		demo: true
	},
	{
		id: 'meta',
		nombre: 'Meta Ads',
		detalle: 'Píxel y conversiones',
		categoria: 'gestion',
		marca: 'meta',
		hace: ['Píxel y API de conversiones para tus anuncios', 'Catálogo para anuncios dinámicos'],
		opciones: [],
		demo: false
	},
	{
		id: 'zap',
		nombre: 'Zapier',
		detalle: '+6.000 apps',
		categoria: 'gestion',
		marca: 'zap',
		hace: ['Conectá Mostrador con otras apps'],
		opciones: [],
		demo: false,
		pronto: true
	}
];

export const IAS = [
	{ id: 'claude', nombre: 'Claude' },
	{ id: 'chatgpt', nombre: 'ChatGPT' },
	{ id: 'gemini', nombre: 'Gemini' },
	{ id: 'mcp', nombre: 'Otra IA (MCP)' }
] as const;

export const PERMISOS_IA = [
	{
		id: 'leer',
		t: 'Ver ventas, stock y clientes',
		d: 'Solo lectura, siempre activo',
		def: true,
		fijo: true
	},
	{ id: 'escribir', t: 'Cargar y editar productos', d: 'Precios, stock y fotos', def: true },
	{ id: 'responder', t: 'Responder mensajes', d: 'WhatsApp e Instagram, con tu tono', def: false },
	{ id: 'plata', t: 'Reembolsos y descuentos', d: 'Siempre te pide confirmación', def: false }
] as const;
