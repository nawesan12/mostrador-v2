/**
 * Datos del onboarding (design/Onboarding Mostrador.dc.html). En la demo el recorrido usa el negocio
 * de ejemplo del handoff (Café La Esquina); con la API, los pasos que dependen de una integración
 * muestran lo que esa integración trae cuando está conectada.
 */
import * as I from '#lib/icons.ts';
import type { IconData } from '#lib/icons.ts';

export const PASOS = [
	'Cuenta',
	'Tu negocio',
	'Canales',
	'Importación',
	'Tu tienda',
	'Cobros y envíos',
	'Tincho',
	'¡Listo!'
] as const;
export const PORCENTAJE = [6, 28, 52, 71, 80, 88, 94, 100] as const;
export const CONSEJOS = [
	'Si entrás con Mercado Pago, me ahorro preguntarte la mitad de las cosas.',
	'Saqué esto de ARCA, de tu Instagram y de Google Maps. Si algo está mal, tocalo y lo corregís.',
	'Cuantas más cuentas conectes, menos vas a cargar a mano. Nada se publica sin que lo revises.',
	'Mientras importo, junto clientes repetidos y ordeno tus productos. Podés seguir, te aviso al terminar.',
	'Usé los colores y las fotos de tu Instagram. Probá otra paleta si querés.',
	'Configuré todo con lo que ya tenías. Solo falta el certificado de ARCA, te guío cuando quieras.',
	'Aprendí tus precios, stock, horarios y políticas. Elegí cómo querés que hable con tus clientes.',
	'¡Listo! Te dejo el panel ordenado y te aviso apenas entre la primera venta.'
] as const;
/** Los consejos de Tincho con una cuenta real: nada de lo que "encontró" la demo. */
export const CONSEJOS_REALES = [
	'Te pido lo justo. Todo lo podés cambiar después desde el panel.',
	'Contame qué querés resolver y ordeno tu panel según eso.',
	'Conectá lo que ya usás cuando quieras: cada cosa se conecta con tu cuenta, desde el panel.',
	'Si venís de otra plataforma, traigo tus productos. Empieza al terminar y te aviso.',
	'Elegí la plantilla; los colores y las secciones los cambiás cuando quieras.',
	'Con un medio de pago ya podés vender. Te dejo el link a cada cosa.',
	'Elegí cómo querés que hable. Nunca inventa precios ni promete lo que no tenés.',
	'¡Listo! Te aviso apenas entre la primera venta.'
] as const;
export const NOTA_PIE_REAL: Record<number, string> = {
	3: 'La importación empieza al terminar y la seguís desde Productos.',
	5: 'Podés cambiar todo después desde Configuración.'
};
export const ILUSTRACION = [
	'saludo',
	'lupa',
	'whatsapp',
	'caja',
	'foto',
	'ticket',
	'hablando',
	'celebrando'
] as const;

/** Marca de cada servicio: ícono o sigla, con su color y fondo. */
export type Logo = { icon?: IconData; word?: string; col: string; bg: string };
export const LOGOS: Record<string, Logo> = {
	mp: { icon: I.siMercadopago, col: '#00B1EA', bg: '#E6F7FD' },
	ml: { word: 'ML', col: '#2D3277', bg: '#FFE600' },
	ig: { icon: I.logosInstagramIcon, col: 'inherit', bg: '#FCEEF3' },
	wa: { icon: I.logosWhatsappIcon, col: 'inherit', bg: '#E9F8EE' },
	tn: { icon: I.phStorefrontFill, col: '#fff', bg: '#2C3E8F' },
	et: { word: 'ET', col: '#fff', bg: '#2E2A6B' },
	gm: { icon: I.logosGoogleMaps, col: 'inherit', bg: '#EEF4FD' },
	and: { word: 'AND', col: '#fff', bg: '#E2001A' },
	arca: { word: 'ARCA', col: '#fff', bg: '#1F3A6E' },
	transf: { icon: I.phBankFill, col: '#6B5244', bg: '#F1E9DF' },
	cash: { icon: I.phMoneyFill, col: '#2F8F63', bg: '#E9F5EE' },
	local: { icon: I.phStorefrontFill, col: '#8A5A3E', bg: '#F3E9DF' },
	free: { icon: I.phGiftFill, col: '#C24E25', bg: '#FCEDE5' },
	pol: { icon: I.phFileTextFill, col: '#6A4FA0', bg: '#EFE9F7' },
	moto: { icon: I.phMopedFill, col: '#9A6A10', bg: '#FDF3DD' }
};

export const TRAE_MP = [
	[I.phIdentificationCardFill, 'Tu nombre, email y CUIT'],
	[I.phUsersThreeFill, 'Los clientes que ya te compraron'],
	[I.phChartLineUpFill, 'Tus ventas, para los informes'],
	[I.phWalletFill, 'Cobros, cuotas y tu alias para transferencias'],
	[I.phQrCodeFill, 'Tu QR del local y el Point']
] as const;

export const DATOS_NEGOCIO = [
	['CUIT · desde Mercado Pago', '27-31458902-0'],
	['Instagram', '@cafelaesquina'],
	['Tienda actual', 'laesquina.mitiendanube.com']
] as const;

export const HALLAZGOS = [
	{
		src: 'ARCA · desde tu CUIT',
		k: 'arca',
		rows: [
			['Razón social', 'Valeria Sosa'],
			['Nombre de fantasía', 'Café La Esquina'],
			['Condición', 'Monotributo · categoría D'],
			['Domicilio fiscal', 'Bv. Oroño 1240, Rosario, Santa Fe']
		]
	},
	{
		src: 'Instagram · @cafelaesquina',
		k: 'ig',
		colores: true,
		rows: [
			['Logo y foto de portada', 'Listos para tu tienda'],
			['Descripción', 'Café de especialidad y todo para el mate, desde 2019'],
			['Productos detectados', '86 fotos con precio en la descripción']
		]
	},
	{
		src: 'Google Maps',
		k: 'gm',
		rows: [
			['Dirección', 'Bv. Oroño 1240, Rosario'],
			['Horario', 'Lunes a sábado de 8 a 20 h'],
			['Teléfono', '+54 341 555-0142'],
			['Reseñas', '4,8 ★ · 312 reseñas']
		]
	}
] as const;

export const OBJETIVOS = [
	[I.phStorefrontFill, 'Vender online por primera vez'],
	[I.phWhatsappLogoFill, 'Ordenar los pedidos de WhatsApp'],
	[I.phCashRegisterFill, 'Unificar el local y la web'],
	[I.phReceiptFill, 'Facturar sin complicarme'],
	[I.phUsersThreeFill, 'Vender a mayoristas'],
	[I.phTrendUpFill, 'Dejar de pagar comisiones']
] as const;

export const CONEXIONES = [
	[
		'mp',
		'Mercado Pago',
		'Cobros, devoluciones y clientes',
		'Conectado al crear tu cuenta · 1.920 clientes · 14 meses de ventas',
		true
	],
	[
		'tn',
		'Tiendanube',
		'Migración completa de tu tienda actual',
		'1.204 productos · 3.880 clientes · 9.112 pedidos · 14 cupones',
		true
	],
	[
		'et',
		'Empretienda',
		'Productos, fotos y clientes de tu tienda',
		'Desde el link de tu tienda y el Excel de clientes, sin claves',
		false
	],
	[
		'ml',
		'Mercado Libre',
		'Publicaciones, stock y preguntas',
		'86 publicaciones · 412 preguntas · reputación verde',
		true
	],
	[
		'wa',
		'WhatsApp Business',
		'Tu número, catálogo y contactos',
		'+54 341 555-0142 · 380 contactos · catálogo de 64 productos',
		true
	],
	[
		'ig',
		'Instagram y Facebook',
		'Fotos, catálogo y mensajes',
		'@cafelaesquina · 86 fotos de productos · 12,4 mil seguidores',
		true
	],
	[
		'gm',
		'Google Maps',
		'Dirección, horarios y reseñas',
		'Bv. Oroño 1240 · lun a sáb 8–20 h · 4,8 ★ (312 reseñas)',
		false
	],
	[
		'arca',
		'ARCA',
		'Punto de venta para facturar',
		'Monotributo D · punto de venta 0004 creado',
		false
	]
] as const;
export type Conexion = (typeof CONEXIONES)[number][0];

/**
 * Con una cuenta real (la API): lo que conviene conectar y configurar, con el link a la pantalla del
 * panel donde se hace. Sin datos de ejemplo.
 */
export const CANALES_REALES = [
	{
		k: 'mp',
		t: 'Mercado Pago',
		d: 'Cobrá con tarjeta, QR y cuotas en tu tienda',
		href: '/panel/integraciones'
	},
	{
		k: 'ml',
		t: 'Mercado Libre',
		d: 'Publicá tus productos con el mismo stock',
		href: '/panel/integraciones'
	},
	{
		k: 'wa',
		t: 'WhatsApp',
		d: 'Atendé los mensajes desde la bandeja del panel',
		href: '/panel/whatsapp'
	},
	{
		k: 'ig',
		t: 'Instagram y Facebook',
		d: 'Los mensajes, en la misma bandeja',
		href: '/panel/integraciones'
	}
];
export const COBROS_REALES = [
	{
		k: 'transf',
		t: 'Transferencia',
		d: 'Cargá tu alias o CBU para ofrecerla al pagar',
		href: '/panel/configuracion'
	},
	{
		k: 'local',
		t: 'Envíos y retiro',
		d: 'Elegí cómo entregás: retiro, Correo Argentino o moto',
		href: '/panel/envios/configurar'
	},
	{
		k: 'arca',
		t: 'Factura electrónica',
		d: 'Conectá ARCA para facturar cada venta',
		href: '/panel/facturacion'
	},
	{
		k: 'pol',
		t: 'Cambios y devoluciones',
		d: 'Tu política, a la vista en la tienda',
		href: '/panel/mi-tienda'
	}
];

export const IMPORTACION = [
	[
		I.phTagFill,
		'Productos',
		'Tiendanube + Mercado Libre, sin duplicados',
		1204,
		'#FCEDE5',
		'#C24E25'
	],
	[I.phImagesFill, 'Fotos', 'Con fondo mejorado y tamaño optimizado', 3412, '#EFE9F7', '#6A4FA0'],
	[
		I.phUsersThreeFill,
		'Clientes',
		'Unificados de Mercado Pago, ML y WhatsApp',
		3880,
		'rgba(47,143,99,.13)',
		'#2F8F63'
	],
	[
		I.phReceiptFill,
		'Pedidos históricos',
		'Para que tus informes arranquen con datos',
		9112,
		'#FDF3DD',
		'#9A6A10'
	],
	[I.phStarFill, 'Reseñas', 'De Google y Mercado Libre', 312, '#FDF3DD', '#C98A12'],
	[I.phTicketFill, 'Cupones', 'Activos en Tiendanube', 14, '#FCEDE5', '#C24E25'],
	[
		I.phArrowsSplitFill,
		'Redirecciones SEO',
		'Para no perder tu lugar en Google',
		1204,
		'#EEF4FD',
		'#3E6BFD'
	]
] as const;
export const REGISTRO_HORNERO = [
	'Unifiqué 412 clientes repetidos entre Mercado Pago y WhatsApp.',
	'Agrupé tus productos en 8 categorías.',
	'Mejoré 128 títulos para que aparezcan mejor en Google.',
	'Detecté 4 productos sin stock en Mercado Libre y los marqué.',
	'Calculé precios con IVA incluido para Factura B.'
] as const;
export const CONFLICTO = [
	['tn', 'Usar Tiendanube', 'Listo: usamos los precios de Tiendanube en todos tus canales.'],
	[
		'ml',
		'Usar Mercado Libre (+10%)',
		'Listo: en Mercado Libre mantenemos tu recargo del 10% automáticamente.'
	],
	['one', 'Revisar uno por uno', 'Te dejamos los 12 en una lista para revisar desde el panel.']
] as const;

/** Lo que trae Instagram para la vista previa de la tienda (colores, foto y frase). */
export const DE_INSTAGRAM = {
	bg: '#F3E9DC',
	ink: '#3B2416',
	acc: '#6B3A22',
	img: '1687920492018-e21cad9023d9',
	tag: 'El mate de cada mañana'
};
export const RECOMENDADAS = ['litoral', 'mostaza', 'yerbal'] as const;
/** Para una tienda real sin rubro conocido: moda, gastronomía y almacén, tres estilos distintos. */
export const PARA_EMPEZAR = ['ceibo', 'mostaza', 'yerbal'] as const;
export const SECCIONES = [
	['dest', I.phStar, 'Destacados'],
	['cat', I.phSquaresFour, 'Categorías'],
	['ig', I.phInstagramLogo, 'Feed de Instagram'],
	['res', I.phChatCenteredText, 'Reseñas']
] as const;
export type Seccion = (typeof SECCIONES)[number][0];

type Opcion = { clave: string; logo: string; t: string; d: string; falta?: boolean };
export const CONFIGURACION: { i: IconData; t: string; items: Opcion[] }[] = [
	{
		i: I.phWalletFill,
		t: 'Cobros',
		items: [
			{
				clave: 'cuotas',
				logo: 'mp',
				t: 'Mercado Pago',
				d: 'Tarjetas, QR y hasta 3 cuotas sin interés'
			},
			{
				clave: 'transf',
				logo: 'transf',
				t: 'Transferencia',
				d: 'Alias cafe.esquina.mp, sacado de tu cuenta'
			},
			{
				clave: 'cash',
				logo: 'cash',
				t: 'Efectivo en el local',
				d: 'Con caja diaria en el punto de venta'
			}
		]
	},
	{
		i: I.phTruckFill,
		t: 'Envíos',
		items: [
			{
				clave: 'andreani',
				logo: 'and',
				t: 'Andreani',
				d: 'Paquete promedio 20×15×10 cm, 1,2 kg (calculado de tus productos)'
			},
			{
				clave: 'retiro',
				logo: 'local',
				t: 'Retiro en el local',
				d: 'Bv. Oroño 1240, en tu horario de Google'
			},
			{
				clave: 'free',
				logo: 'free',
				t: 'Envío gratis desde $ 60.000',
				d: 'Sugerido por Tincho: 1,8× tu ticket promedio'
			},
			{
				clave: 'moto',
				logo: 'moto',
				t: 'Moto propia en Rosario',
				d: 'Activalo si hacés entregas en el día'
			}
		]
	},
	{
		i: I.phReceiptFill,
		t: 'Facturas y políticas',
		items: [
			{
				clave: 'arca',
				logo: 'arca',
				t: 'Factura electrónica ARCA',
				d: 'Punto de venta 0004 creado · Factura C por ser monotributo',
				falta: true
			},
			{
				clave: 'pol',
				logo: 'pol',
				t: 'Cambios y devoluciones',
				d: '30 días, redactado según la ley de defensa del consumidor'
			}
		]
	}
];

export const TONOS = {
	cercano: [
		'Cercano',
		'¡Hola Lucía! Sí, nos quedan 3 en negro. ¿Te lo separo? Te llega el jueves a Córdoba.'
	],
	profesional: [
		'Profesional',
		'Hola Lucía, buenas tardes. Sí, tenemos 3 unidades en negro. ¿Desea que se lo reserve? El envío a Córdoba demora 48 h.'
	],
	canchero: [
		'Canchero',
		'¡Holaa Lu! Sí, quedan 3 negros y vuelan. ¿Te guardo uno? El jueves lo tenés en Córdoba.'
	]
} as const;
export type Tono = keyof typeof TONOS;

export const RESUMEN = [
	['1.204', 'productos'],
	['3.880', 'clientes'],
	['5', 'canales conectados'],
	['0', 'cosas cargadas a mano']
] as const;

/** El QR de muestra del paso final (7×7, el patrón del handoff). */
export const QR = Array.from({ length: 49 }, (_, i) => {
	const r = Math.floor(i / 7);
	const c = i % 7;
	const f = (r < 2 && c < 2) || (r < 2 && c > 4) || (r > 4 && c < 2);
	return f || (i * 37 + 11) % 5 < 2;
});

export const NOTA_PIE: Record<number, string> = {
	2: 'Nada se publica sin que lo revises.',
	3: 'Si cerrás, la importación sigue y te avisamos por email.',
	5: 'Podés cambiar todo después desde Configuración.'
};
