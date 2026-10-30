/**
 * Las integraciones del panel (design/PanelOperacion.dc.html, sInt: el orden y lo conectado de la
 * tienda de ejemplo salen de `INT`; TikTok Shop queda afuera porque no opera en Argentina, y Andreani,
 * OCA, MODO y Gemini porque no se pueden conectar sin contrato o no admiten conectores). Cada una
 * dice qué sincroniza y qué opciones tiene en "Configurar" (el handoff no la diseña: se arma con sus
 * piezas).
 */
import * as I from '#lib/icons.ts';

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
			'Las preguntas llegan a Reseñas y preguntas para que las respondas'
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
		// Sin conexión real todavía: no se ofrece como disponible.
		pronto: true,
		nombre: 'WhatsApp Business',
		detalle: 'Bandeja y campañas',
		categoria: 'canales',
		marca: 'wa',
		hace: [
			'Los chats llegan a la bandeja del panel',
			'Tincho responde y manda links de pago',
			'Campañas con el costo exacto antes de enviar'
		],
		opciones: [
			{ id: 'hornero', t: 'Tincho responde', d: 'Si está apagado, sólo sugiere', def: true },
			{ id: 'horario', t: 'Sólo fuera de horario', d: 'En horario atiende tu equipo', def: false }
		],
		demo: true
	},
	{
		id: 'ig',
		// Sin conexión real todavía: no se ofrece como disponible.
		pronto: true,
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
		detalle: 'Tincho y avisos',
		categoria: 'canales',
		marca: 'tg',
		hace: [
			'Preguntale a Tincho por ventas, pedidos y stock desde Telegram',
			'Te avisa cada venta al momento',
			'Sólo ve la tienda que vinculaste'
		],
		opciones: [],
		demo: false
	},
	{
		id: 'arca',
		nombre: 'ARCA',
		detalle: 'Factura electrónica',
		categoria: 'gestion',
		marca: 'arca',
		hace: [
			'Factura A, B o C sola en cada venta',
			'Nota de crédito al anular una factura',
			'Te avisa antes de que venza el certificado'
		],
		opciones: [
			{ id: 'auto', t: 'Facturar al cobrar', d: 'Si está apagado, facturás a mano', def: true }
		],
		demo: true
	},
	{
		id: 'payway',
		// Sin conexión real todavía: no se ofrece como disponible.
		pronto: true,
		nombre: 'Payway',
		detalle: 'Tarjetas y cuotas',
		categoria: 'cobros',
		marca: 'payway',
		hace: ['Tarjetas de crédito y débito', 'Cuotas con las promos de cada banco'],
		opciones: [],
		demo: false
	},
	{
		// Con la cuenta de MiCorreo de la tienda (API MiCorreo, docs/integraciones/correo-argentino.md).
		id: 'correo',
		nombre: 'Correo Argentino',
		detalle: 'Sucursal y domicilio',
		categoria: 'envios',
		marca: 'correo',
		hace: [
			'Cotiza a domicilio y a sucursal en el checkout, con el precio de Correo',
			'Carga el envío de cada pedido en tu MiCorreo',
			'Quien compra ve el número de seguimiento'
		],
		opciones: [],
		// La demo no tiene cuenta de MiCorreo: no se muestra conectada.
		demo: false
	},
	{
		id: 'ga',
		nombre: 'Google Analytics',
		detalle: 'Métricas',
		categoria: 'gestion',
		marca: 'ga',
		hace: [
			'Visitas, productos vistos, carritos y ventas de tu tienda en tu GA4',
			'Sólo se carga si quien compra acepta las cookies de medición'
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
		hace: [
			'Tus productos en Google Shopping, gratis',
			'Precio, oferta y stock al día con un feed que Google lee solo'
		],
		opciones: [],
		demo: true
	},
	{
		id: 'meta',
		nombre: 'Meta Ads',
		detalle: 'Píxel y conversiones',
		categoria: 'gestion',
		marca: 'meta',
		hace: [
			'El píxel mide visitas, carritos y ventas de tus anuncios',
			'Tu catálogo para anuncios dinámicos, con el mismo feed'
		],
		opciones: [],
		demo: false
	}
];

/** "Manejá Tu Local Web desde tu IA" (aiProv del handoff). */
export const IAS = [
	{ id: 'claude', nombre: 'Claude', icon: I.siClaude, col: '#D97757' },
	{ id: 'chatgpt', nombre: 'ChatGPT', icon: I.siOpenai, col: '#10A37F' },
	{ id: 'mcp', nombre: 'Otra IA (MCP)', icon: I.phPlugsConnectedBold, col: '#8A5A3E' }
] as const;

export type IdIa = (typeof IAS)[number]['id'];

/** La que se muestra cuando el cliente MCP no es de una IA conocida. */
export const OTRA_IA = IAS[2];

/** A qué IA corresponde un cliente MCP, por el nombre con que se registró (el resto: "Otra IA"). */
export function iaDe(nombre: string): (typeof IAS)[number] {
	const id: IdIa = /claude|anthropic/i.test(nombre)
		? 'claude'
		: /chatgpt|openai/i.test(nombre)
			? 'chatgpt'
			: 'mcp';
	return IAS.find((p) => p.id === id) ?? OTRA_IA;
}

/** Qué puede hacer la IA (aiPerms del handoff). "Ver" está siempre. */
export const PERMISOS_IA = [
	{
		id: 'leer',
		icon: I.phEyeFill,
		t: 'Ver ventas, stock y clientes',
		d: 'Solo lectura, siempre activo',
		def: true
	},
	{
		id: 'escribir',
		icon: I.phPencilSimpleFill,
		t: 'Cargar y editar productos',
		d: 'Precios, stock y fotos',
		def: true
	},
	{
		id: 'responder',
		icon: I.phChatCircleDotsFill,
		t: 'Responder mensajes',
		d: 'WhatsApp e Instagram, con tu tono',
		def: false
	},
	{
		id: 'plata',
		icon: I.phCurrencyCircleDollarFill,
		t: 'Reembolsos y descuentos',
		d: 'Crear cupones de descuento',
		def: false
	}
] as const;

export const CATEGORIAS = [
	['todas', 'Todas'],
	['canales', 'Canales'],
	['cobros', 'Cobros'],
	['envios', 'Envíos'],
	['gestion', 'Gestión']
] as const;

/** Estado de las integraciones de la tienda (en la demo se simula; en la real, cada una por su PR). */
export type EstadoIntegraciones = {
	conectadas: Record<string, boolean>;
	opciones: Record<string, Record<string, boolean>>;
	ia: Record<string, boolean>;
	permisosIa: Record<string, boolean>;
};

export type AjusteIntegracion =
	| { tipo: 'conexion'; id: string; valor: boolean }
	| { tipo: 'opcion'; id: string; clave: string; valor: boolean }
	| { tipo: 'ia'; id: string; valor: boolean }
	| { tipo: 'permisoIa'; id: string; valor: boolean };

/** Las integraciones que se configuran con la medición (lib/tienda/medicion.ts). */
export const DE_MEDICION = ['ga', 'gm', 'meta'] as const;

/** Conectada = el ID cargado (GA4, píxel) o el feed marcado como cargado (Shopping, catálogo). */
export function conectadasDeMedicion(m: {
	ga4: string | null;
	pixel: string | null;
	feeds: { gm: boolean; meta: boolean };
}) {
	return { ga: !!m.ga4, gm: m.feeds.gm, meta: !!m.pixel || m.feeds.meta };
}

/** El estado de fábrica (demo: lo conectado en la tienda de ejemplo; real: nada conectado). */
export function estadoInicial(demo: boolean): EstadoIntegraciones {
	return {
		conectadas: Object.fromEntries(INTEGRACIONES.map((i) => [i.id, demo && i.demo])),
		opciones: Object.fromEntries(
			INTEGRACIONES.map((i) => [i.id, Object.fromEntries(i.opciones.map((o) => [o.id, o.def]))])
		),
		ia: Object.fromEntries(IAS.map((p) => [p.id, false])),
		permisosIa: Object.fromEntries(PERMISOS_IA.map((p) => [p.id, p.def]))
	};
}
