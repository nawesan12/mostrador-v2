// Generado por scripts/datos-landing.ts desde design/Mostrador v4.dc.html — no editar a mano.
// Los íconos son referencias a #lib/icons.ts: el bundle lleva sólo los que se usan.
import * as I from '#lib/icons.ts';

export const FEED = [
	['wa', 'Pedido por WhatsApp', 'Lucía · Termo 1 L negro', 42900],
	['mp', 'Cobro aprobado', 'Mercado Pago · 3 cuotas', 128500],
	['ml', 'Venta en Mercado Libre', 'Mate de calabaza x2', 37000],
	['ig', 'Pedido desde Instagram', 'Sofía · Kit matero', 58400],
	['arca', 'Factura B emitida', 'Venta en el local #0412', 8500],
	['tg', 'Pedido por Telegram', 'Martín · Yerba 1 kg x3', 20400],
	['andreani', 'Envío despachado', 'Pedido #1046 · Córdoba', 21990]
] as const;

export const INTEG = [
	[
		'WhatsApp',
		'canales',
		{
			icon: I.logosWhatsappIcon
		},
		'Bandeja, catálogo y campañas'
	],
	[
		'Telegram',
		'canales',
		{
			icon: I.logosTelegram
		},
		'Bot de pedidos y avisos'
	],
	[
		'Instagram',
		'canales',
		{
			icon: I.logosInstagramIcon
		},
		'Shopping y mensajes'
	],
	[
		'Facebook',
		'canales',
		{
			icon: I.logosFacebook
		},
		'Catálogo y tienda'
	],
	[
		'Messenger',
		'canales',
		{
			icon: I.logosMessenger
		},
		'Mensajes en la bandeja'
	],
	[
		'Mercado Libre',
		'canales',
		{
			w: 'mercado libre',
			abbr: 'ML',
			bg: '#FFE600',
			fg: '#2D3277'
		},
		'Stock y pedidos sincronizados'
	],
	[
		'TikTok Shop',
		'canales',
		{
			icon: I.logosTiktokIcon
		},
		'Catálogo y píxel'
	],
	[
		'Google',
		'canales',
		{
			icon: I.logosGoogleIcon
		},
		'Shopping y Merchant Center'
	],
	[
		'Google Maps',
		'canales',
		{
			icon: I.logosGoogleMaps
		},
		'Tu local en el mapa'
	],
	[
		'PedidosYa',
		'canales',
		{
			w: 'PedidosYa',
			abbr: 'PYa',
			bg: '#FA0050',
			fg: '#fff'
		},
		'Pedidos del delivery'
	],
	[
		'Rappi',
		'canales',
		{
			w: 'Rappi',
			abbr: 'Rappi',
			bg: '#FF441F',
			fg: '#fff'
		},
		'Pedidos del delivery'
	],
	[
		'Mercado Pago',
		'cobros',
		{
			icon: I.siMercadopago,
			col: '#00B1EA'
		},
		'Checkout, QR y Point'
	],
	[
		'Payway',
		'cobros',
		{
			w: 'payway',
			abbr: 'PW',
			bg: '#0E1E4A',
			fg: '#fff'
		},
		'Tarjetas y cuotas'
	],
	[
		'MODO',
		'cobros',
		{
			w: 'MODO',
			abbr: 'MODO',
			bg: '#00A88F',
			fg: '#fff'
		},
		'Pago con apps bancarias'
	],
	[
		'Ualá',
		'cobros',
		{
			w: 'ualá',
			abbr: 'ualá',
			bg: '#3E6BFD',
			fg: '#fff'
		},
		'Ualá Bis y link de pago'
	],
	[
		'Naranja X',
		'cobros',
		{
			w: 'naranja X',
			abbr: 'NX',
			bg: '#FE5000',
			fg: '#fff'
		},
		'Plan Z y cuotas'
	],
	[
		'Cuenta DNI',
		'cobros',
		{
			w: 'Cuenta DNI',
			abbr: 'DNI',
			bg: '#00A650',
			fg: '#fff'
		},
		'QR interoperable'
	],
	[
		'Visa',
		'cobros',
		{
			icon: I.logosVisa
		},
		'Crédito y débito'
	],
	[
		'Mastercard',
		'cobros',
		{
			icon: I.logosMastercard
		},
		'Crédito y débito'
	],
	[
		'American Express',
		'cobros',
		{
			icon: I.logosAmex
		},
		'Crédito'
	],
	[
		'PayPal',
		'cobros',
		{
			icon: I.logosPaypal
		},
		'Ventas al exterior'
	],
	[
		'Andreani',
		'envios',
		{
			w: 'Andreani',
			abbr: 'AND',
			bg: '#E2001A',
			fg: '#fff'
		},
		'Etiquetas y seguimiento'
	],
	[
		'OCA',
		'envios',
		{
			w: 'OCA',
			abbr: 'OCA',
			bg: '#4B2A85',
			fg: '#fff'
		},
		'Envíos a todo el país'
	],
	[
		'Correo Argentino',
		'envios',
		{
			w: 'Correo Arg.',
			abbr: 'CA',
			bg: '#FFCE00',
			fg: '#003DA5'
		},
		'Sucursal y domicilio'
	],
	[
		'DHL',
		'envios',
		{
			icon: I.siDhl,
			col: '#D40511'
		},
		'Envíos internacionales'
	],
	[
		'FedEx',
		'envios',
		{
			icon: I.siFedex,
			col: '#4D148C'
		},
		'Envíos internacionales'
	],
	[
		'Uber Flash',
		'envios',
		{
			icon: I.siUber,
			col: '#000'
		},
		'Envíos en el día'
	],
	[
		'ARCA',
		'gestion',
		{
			w: 'ARCA',
			abbr: 'ARCA',
			bg: '#1F3A6E',
			fg: '#fff'
		},
		'Factura electrónica A, B y C'
	],
	[
		'Google Analytics',
		'gestion',
		{
			icon: I.logosGoogleAnalytics
		},
		'Métricas de la tienda'
	],
	[
		'Google Ads',
		'gestion',
		{
			icon: I.logosGoogleAds
		},
		'Conversiones'
	],
	[
		'Meta Ads',
		'gestion',
		{
			icon: I.logosMetaIcon
		},
		'Píxel y API de conversiones'
	],
	[
		'Mailchimp',
		'gestion',
		{
			icon: I.logosMailchimpFreddie
		},
		'Email marketing'
	],
	[
		'HubSpot',
		'gestion',
		{
			icon: I.logosHubspot
		},
		'CRM'
	],
	[
		'Zapier',
		'gestion',
		{
			icon: I.logosZapier
		},
		'+6.000 apps'
	],
	[
		'Google Drive',
		'gestion',
		{
			icon: I.logosGoogleDrive
		},
		'Exportes automáticos'
	],
	[
		'Slack',
		'gestion',
		{
			icon: I.logosSlackIcon
		},
		'Avisos al equipo'
	],
	[
		'Caja',
		'ia',
		{
			img: '/caja/01-saludo.png'
		},
		'Finanzas y monotributo'
	],
	[
		'Claude',
		'ia',
		{
			icon: I.siClaude,
			col: '#D97757'
		},
		'Conector MCP'
	],
	[
		'ChatGPT',
		'ia',
		{
			icon: I.siOpenai,
			col: '#10A37F'
		},
		'Conector y apps'
	],
	[
		'Gemini',
		'ia',
		{
			icon: I.siGooglegemini,
			col: '#4285F4'
		},
		'Extensión con tu cuenta'
	],
	[
		'Cualquier IA (MCP)',
		'ia',
		{
			icon: I.phPlugsConnectedBold,
			col: '#8A5A3E'
		},
		'Servidor MCP abierto'
	],
	[
		'Tiendanube',
		'canales',
		{
			w: 'Tiendanube',
			abbr: 'TN',
			bg: '#2C3357',
			fg: '#fff'
		},
		'Migración en un clic'
	],
	[
		'Shopify',
		'canales',
		{
			icon: I.logosShopify
		},
		'Importá tu tienda'
	],
	[
		'WooCommerce',
		'canales',
		{
			icon: I.logosWoocommerceIcon
		},
		'Importá tu tienda'
	],
	[
		'YouTube',
		'canales',
		{
			icon: I.logosYoutubeIcon
		},
		'Productos en tus videos'
	],
	[
		'Pinterest',
		'canales',
		{
			icon: I.logosPinterest
		},
		'Pines de producto'
	],
	[
		'Getnet',
		'cobros',
		{
			w: 'Getnet',
			abbr: 'getnet',
			bg: '#EC0000',
			fg: '#fff'
		},
		'Posnet y link de pago'
	],
	[
		'Banco Galicia',
		'cobros',
		{
			w: 'Galicia',
			abbr: 'GAL',
			bg: '#F26722',
			fg: '#fff'
		},
		'Conciliación de transferencias'
	],
	[
		'Santander',
		'cobros',
		{
			w: 'Santander',
			abbr: 'SAN',
			bg: '#EC0000',
			fg: '#fff'
		},
		'Conciliación de transferencias'
	],
	[
		'BBVA',
		'cobros',
		{
			w: 'BBVA',
			abbr: 'BBVA',
			bg: '#004481',
			fg: '#fff'
		},
		'Conciliación de transferencias'
	],
	[
		'Banco Nación',
		'cobros',
		{
			w: 'Nación',
			abbr: 'BNA',
			bg: '#1E4D8C',
			fg: '#fff'
		},
		'Conciliación de transferencias'
	],
	[
		'Lemon',
		'cobros',
		{
			w: 'Lemon',
			abbr: 'lemon',
			bg: '#00F068',
			fg: '#0B0B0B'
		},
		'Cobros en cripto'
	],
	[
		'Binance Pay',
		'cobros',
		{
			icon: I.siBinance,
			col: '#F0B90B'
		},
		'USDT y cripto'
	],
	[
		'Stripe',
		'cobros',
		{
			icon: I.logosStripe
		},
		'Ventas al exterior'
	],
	[
		'Enviopack',
		'envios',
		{
			w: 'Enviopack',
			abbr: 'EP',
			bg: '#00B6E6',
			fg: '#fff'
		},
		'Cotizador multi transporte'
	],
	[
		'Cruz del Sur',
		'envios',
		{
			w: 'Cruz del Sur',
			abbr: 'CDS',
			bg: '#E30613',
			fg: '#fff'
		},
		'Encomiendas al interior'
	],
	[
		'Via Cargo',
		'envios',
		{
			w: 'Via Cargo',
			abbr: 'VC',
			bg: '#E4002B',
			fg: '#fff'
		},
		'Encomiendas'
	],
	[
		'Moova',
		'envios',
		{
			w: 'Moova',
			abbr: 'moova',
			bg: '#6B2FFA',
			fg: '#fff'
		},
		'Envíos en el día'
	],
	[
		'PedidosYa Envíos',
		'envios',
		{
			w: 'PedidosYa',
			abbr: 'PYa',
			bg: '#FA0050',
			fg: '#fff'
		},
		'Cadetería en minutos'
	],
	[
		'Colppy',
		'gestion',
		{
			w: 'Colppy',
			abbr: 'colppy',
			bg: '#2BB673',
			fg: '#fff'
		},
		'Contabilidad'
	],
	[
		'Xubio',
		'gestion',
		{
			w: 'Xubio',
			abbr: 'xubio',
			bg: '#00A3E0',
			fg: '#fff'
		},
		'Contabilidad'
	],
	[
		'Tango',
		'gestion',
		{
			w: 'Tango',
			abbr: 'TANGO',
			bg: '#E30613',
			fg: '#fff'
		},
		'Gestión y ERP'
	],
	[
		'Google Sheets',
		'gestion',
		{
			icon: I.siGooglesheets,
			col: '#34A853'
		},
		'Stock y precios en planilla'
	],
	[
		'Excel',
		'gestion',
		{
			icon: I.siMicrosoftexcel,
			col: '#217346'
		},
		'Importar y exportar'
	],
	[
		'Gmail',
		'gestion',
		{
			icon: I.logosGoogleGmail
		},
		'Mails a clientes'
	],
	[
		'Google Calendar',
		'gestion',
		{
			icon: I.logosGoogleCalendar
		},
		'Turnos y retiros'
	],
	[
		'Notion',
		'gestion',
		{
			icon: I.logosNotionIcon
		},
		'Reportes del equipo'
	],
	[
		'Klaviyo',
		'gestion',
		{
			col: '#232426',
			w: 'Klaviyo',
			abbr: 'K',
			bg: '#232426',
			fg: '#fff'
		},
		'Email y SMS'
	],
	[
		'Make',
		'gestion',
		{
			icon: I.siMake,
			col: '#6D00CC'
		},
		'Automatizaciones'
	],
	[
		'n8n',
		'gestion',
		{
			icon: I.siN8n,
			col: '#EA4B71'
		},
		'Automatizaciones'
	]
] as const;

export const CATS = [
	['todas', 'Todas'],
	['ia', 'IA y finanzas'],
	['canales', 'Canales de venta'],
	['cobros', 'Cobros'],
	['envios', 'Envíos'],
	['gestion', 'Gestión y marketing']
] as const;

export const LPLANS = [
	[
		'Gratis',
		'Para probar y empezar',
		0,
		[
			'Productos ilimitados',
			'Mercado Pago, transferencia y retiro',
			'Hornero: 100 respuestas con IA',
			'Respuestas automáticas sin límite',
			'1 persona · subdominio de Mostrador'
		],
		'Crear tienda gratis'
	],
	[
		'Emprende',
		'Para vender todos los días',
		19900,
		[
			'Dominio propio y sin marca Mostrador',
			'Todos los medios de pago y envío',
			'Factura de ARCA automática',
			'Hornero: 1.000 respuestas con IA',
			'Carritos abandonados · 2 personas'
		],
		'Probar 14 días'
	],
	[
		'Pro',
		'El más elegido',
		49900,
		[
			'Mercado Libre e Instagram con un solo stock',
			'Punto de venta · 1 caja',
			'Hornero: 3.000 respuestas con IA',
			'Acciones masivas y estadísticas',
			'5 personas con permisos'
		],
		'Probar 14 días',
		true
	],
	[
		'Negocio',
		'Varias sucursales',
		149900,
		[
			'Sucursales, depósitos y cajas ilimitadas',
			'Listas mayoristas, API y webhooks',
			'Hornero: 15.000 respuestas con IA',
			'Personas ilimitadas',
			'Gerente de cuenta'
		],
		'Probar 14 días'
	]
] as const;

export const CMP = [
	['Plan para empezar', 'Gratis', 'Gratis (Inicial)', 1, 1],
	['Plan para vender todos los días', 'Emprende · $ 19.900', 'Esencial · $ 27.999', 1, 1],
	['Plan para crecer', 'Pro · $ 49.900', 'Impulso · $ 79.999', 1, 1],
	['Plan para varias sucursales', 'Negocio · $ 149.900', 'Escala · $ 244.999', 1, 1],
	['Atención por WhatsApp', 'Desde Emprende', 'Desde Impulso', 1, 2],
	['Acciones masivas', 'Desde Pro', 'Desde Impulso', 1, 2],
	['Usuarios con permisos', 'Desde Pro', 'Desde Escala', 1, 2],
	['Migración asistida', 'Gratis en todos', 'En Evolución (a consultar)', 1, 2],
	[
		'Asistente de IA en WhatsApp e Instagram',
		'En todos (según plan)',
		'No figura en el plan',
		1,
		0
	],
	['Factura de ARCA automática', 'Desde Emprende', 'No figura en el plan', 1, 0],
	['Punto de venta en el local', 'Desde Pro', 'No figura en el plan', 1, 0],
	['Verificación en dos pasos y copias restaurables', 'En todos', 'No figura en el plan', 1, 0]
] as const;

export const CHAT = [
	{
		k: 'c',
		t: 'Hola! ¿Tienen el termo de 1 L en negro?'
	},
	{
		k: 'h',
		t: '¡Hola Lucía! Sí, nos quedan 3 en negro. Te lo dejo acá:'
	},
	{
		k: 'prod'
	},
	{
		k: 'c',
		t: '¿Hacen envíos a Córdoba?'
	},
	{
		k: 'h',
		t: 'Sí, por Andreani te llega el jueves. El envío sale $ 4.200, o es gratis si sumás la bombilla de acero ($ 1.700).'
	},
	{
		k: 'c',
		t: 'Dale, sumame la bombilla'
	},
	{
		k: 'h',
		t: '¡Listo! Son $ 44.600 con envío gratis. Podés pagar acá:'
	},
	{
		k: 'pay'
	},
	{
		k: 'sys',
		t: 'Pago aprobado · Pedido #1049 creado · Factura B emitida'
	}
] as const;

export const SERIES = {
	hoy: {
		label: 'hoy',
		total: 412800,
		delta: '+12,4%',
		s: [8, 12, 10, 14, 22, 30, 26, 34, 41, 38, 46, 52, 49, 58],
		p: [6, 9, 11, 12, 18, 20, 24, 25, 30, 31, 35, 36, 40, 42],
		axis: ['8 h', '11 h', '14 h', '17 h', '20 h'],
		kpi: [
			['Pedidos', '18', '+3', 'vs. ayer'],
			['Ticket promedio', '$ 22.930', '+4,1%', 'vs. ayer'],
			['Conversión', '3,1%', '−0,2', 'puntos'],
			['Clientes nuevos', '7', '+2', 'vs. ayer']
		]
	},
	'7d': {
		label: 'últimos 7 días',
		total: 4812300,
		delta: '+18,2%',
		s: [42, 48, 45, 58, 62, 55, 70, 74, 68, 82, 88, 84, 96, 104],
		p: [40, 41, 44, 46, 50, 49, 55, 58, 57, 62, 66, 64, 70, 74],
		axis: ['mar', 'jue', 'sáb', 'dom', 'lun'],
		kpi: [
			['Pedidos', '142', '+21', 'vs. semana anterior'],
			['Ticket promedio', '$ 33.890', '+6,8%', 'vs. semana anterior'],
			['Conversión', '3,4%', '+0,4', 'puntos'],
			['Clientes nuevos', '38', '+9', 'vs. semana anterior']
		]
	},
	'30d': {
		label: 'últimos 30 días',
		total: 19640500,
		delta: '+9,7%',
		s: [60, 64, 58, 66, 72, 70, 78, 74, 82, 86, 80, 90, 94, 98],
		p: [56, 60, 57, 62, 64, 66, 68, 70, 72, 74, 75, 78, 80, 84],
		axis: ['20 sep', '27 sep', '4 oct', '11 oct', '19 oct'],
		kpi: [
			['Pedidos', '611', '+48', 'vs. mes anterior'],
			['Ticket promedio', '$ 32.140', '+2,3%', 'vs. mes anterior'],
			['Conversión', '3,2%', '+0,1', 'puntos'],
			['Clientes nuevos', '154', '+31', 'vs. mes anterior']
		]
	}
} as const;

export const NOTIFS = [
	[
		'wa',
		'Nuevo mensaje de Lucía',
		'“¿Tenés el termo en negro?” · Hornero ya respondió',
		'hace 4 min',
		true
	],
	[
		'mp',
		'Liquidación disponible',
		'$ 1.284.500 de Mercado Pago ya están en tu cuenta',
		'hace 32 min',
		true
	],
	[
		'ml',
		'Pregunta en Mercado Libre',
		'“¿Es apto lavavajillas?” en Mate de calabaza',
		'hace 1 h',
		true
	],
	['warn', 'Stock bajo', 'Termo 1 L · Negro: quedan 3 unidades', 'hace 2 h', false],
	['arca', 'Certificado por vencer', 'Tu certificado de ARCA vence en 5 días', 'ayer', false],
	['star', 'Nueva reseña de 5 estrellas', '“Llegó rapidísimo y súper bien embalado”', 'ayer', false]
] as const;

export const ORDERS = [
	[
		'#1048',
		'Lucía Fernández',
		'WhatsApp',
		'wa',
		'Para preparar',
		'$ 42.900',
		'4 min',
		'1531058240690-006c446962d8'
	],
	['#1047', 'Distribuidora Sur', 'Mayorista', 'pos', 'Pagado', '$ 412.000', '26 min', null],
	[
		'#1046',
		'Martín Gómez',
		'Mercado Libre',
		'ml',
		'Enviado',
		'$ 21.990',
		'1 h',
		'1687422811062-a966b55cb217'
	],
	[
		'#1045',
		'Norma Ruiz',
		'Instagram',
		'ig',
		'Entregado',
		'$ 58.400',
		'2 h',
		'1764173039506-c5cc7b8ec14b'
	],
	[
		'#1044',
		'Carla Ruiz',
		'Tienda online',
		'web',
		'Pago pendiente',
		'$ 54.300',
		'3 h',
		'1520333789090-1afc82db536a'
	]
] as const;

export const OR = [
	{
		n: '#1048',
		c: 'Lucía Fernández',
		ph: '1531058240690-006c446962d8',
		ch: 'wa',
		chn: 'WhatsApp',
		pay: 'mp',
		payn: 'Mercado Pago',
		pst: 'Aprobado',
		env: 'andreani',
		envn: 'Andreani · Córdoba',
		s: 'Para preparar',
		fac: true,
		h: '4 min',
		items: [
			['Termo 1 L · Negro', 1, 42900],
			['Bombilla de acero', 1, 1700]
		],
		t: 44600
	},
	{
		n: '#1047',
		c: 'Distribuidora Sur',
		ph: null,
		ch: 'pos',
		chn: 'Mayorista',
		pay: 'transf',
		payn: 'Transferencia',
		pst: 'Acreditada',
		env: 'moto',
		envn: 'Moto propia · Lanús',
		s: 'Pagado',
		fac: true,
		h: '26 min',
		items: [
			['Yerba orgánica 1 kg', 40, 8200],
			['Mate de calabaza', 10, 9000]
		],
		t: 418000
	},
	{
		n: '#1046',
		c: 'Martín Gómez',
		ph: '1687422811062-a966b55cb217',
		ch: 'ml',
		chn: 'Mercado Libre',
		pay: 'mp',
		payn: 'Mercado Pago',
		pst: 'Aprobado',
		env: 'andreani',
		envn: 'Andreani · Rosario',
		s: 'Enviado',
		fac: true,
		h: '1 h',
		items: [
			['Mate de calabaza', 2, 9000],
			['Yerba orgánica 1 kg', 1, 3990]
		],
		t: 21990
	},
	{
		n: '#1045',
		c: 'Norma Ruiz',
		ph: '1764173039506-c5cc7b8ec14b',
		ch: 'ig',
		chn: 'Instagram',
		pay: 'modo',
		payn: 'MODO',
		pst: 'Aprobado',
		env: 'oca',
		envn: 'OCA · Mendoza',
		s: 'Entregado',
		fac: true,
		h: '2 h',
		items: [['Kit matero completo', 1, 58400]],
		t: 58400
	},
	{
		n: '#1044',
		c: 'Carla Ruiz',
		ph: '1520333789090-1afc82db536a',
		ch: 'web',
		chn: 'Tienda online',
		pay: 'transf',
		payn: 'Transferencia',
		pst: 'Pendiente',
		env: 'retiro',
		envn: 'Retira en el local',
		s: 'Pago pendiente',
		fac: false,
		h: '3 h',
		items: [
			['Termo 1 L · Acero', 1, 46300],
			['Yerba orgánica 1 kg', 1, 8000]
		],
		t: 54300
	},
	{
		n: '#1043',
		c: 'Julián Pereyra',
		ph: '1687422810663-c316494f725a',
		ch: 'tg',
		chn: 'Telegram',
		pay: 'payway',
		payn: 'Payway',
		pst: 'Aprobado',
		env: 'andreani',
		envn: 'Andreani · La Plata',
		s: 'Para preparar',
		fac: true,
		h: '4 h',
		items: [
			['Bombilla de acero', 6, 1700],
			['Mate de calabaza', 2, 10900]
		],
		t: 32000
	},
	{
		n: '#1042',
		c: 'Consumidor final',
		ph: null,
		ch: 'pos',
		chn: 'Local',
		pay: 'efectivo',
		payn: 'Efectivo',
		pst: 'Cobrado',
		env: 'retiro',
		envn: 'En el local',
		s: 'Entregado',
		fac: true,
		h: '5 h',
		items: [
			['Yerba orgánica 1 kg', 1, 6800],
			['Bombilla de acero', 1, 1700]
		],
		t: 8500
	},
	{
		n: '#1041',
		c: 'Valeria Sosa',
		ph: '1556745753-b2904692b3cd',
		ch: 'wa',
		chn: 'WhatsApp',
		pay: 'uala',
		payn: 'Ualá',
		pst: 'Pendiente',
		env: 'moto',
		envn: 'Moto propia · Quilmes',
		s: 'Pago pendiente',
		fac: false,
		h: '6 h',
		items: [['Termo 1 L · Negro', 1, 42900]],
		t: 42900
	},
	{
		n: '#1040',
		c: 'Pablo Medina',
		ph: '1545262722-9e0d80a0bc01',
		ch: 'ml',
		chn: 'Mercado Libre',
		pay: 'mp',
		payn: 'Mercado Pago',
		pst: 'Aprobado',
		env: 'andreani',
		envn: 'Andreani · Salta',
		s: 'Enviado',
		fac: true,
		h: '7 h',
		items: [['Mate de calabaza', 1, 18500]],
		t: 18500
	}
] as const;

export const AREAS = [
	{
		id: 'inicio',
		label: 'Inicio',
		icon: I.phHouseSimple,
		items: [
			['Resumen', I.phSquaresFour],
			['Para hacer hoy', I.phCheckSquareOffset, 4],
			['Actividad', I.phPulse]
		]
	},
	{
		id: 'ventas',
		label: 'Ventas',
		icon: I.phShoppingBagOpen,
		dot: true,
		items: [
			['Pedidos', I.phPackage, 7],
			['Devoluciones', I.phArrowUUpLeft],
			['Ventas en local', I.phStorefront],
			['Carritos abandonados', I.phShoppingCartSimple, 12],
			['WhatsApp', I.phWhatsappLogo, 3],
			['Clientes', I.phUsersThree],
			['Reseñas y preguntas', I.phStar]
		]
	},
	{
		id: 'catalogo',
		label: 'Catálogo',
		icon: I.phTag,
		items: [
			['Productos', I.phTag],
			['Inventario', I.phStack],
			['Listas de precio', I.phListDashes]
		]
	},
	{
		id: 'crecer',
		label: 'Crecer',
		icon: I.phChartLineUp,
		items: [
			['Marketing', I.phMegaphone],
			['Cupones', I.phTicket],
			['Canales', I.phLinkSimple],
			['Informes', I.phChartBar]
		]
	},
	{
		id: 'operacion',
		label: 'Operación',
		icon: I.phWallet,
		items: [
			['Finanzas', I.phWallet],
			['Facturación', I.phReceipt],
			['Envíos', I.phTruck]
		]
	},
	{
		id: 'tienda',
		label: 'Tienda',
		icon: I.phStorefront,
		items: [
			['Mi tienda', I.phPaintBrushBroad],
			['Configuración', I.phGearSix],
			['Integraciones', I.phPlugsConnected]
		]
	},
	{
		id: 'cuenta',
		label: 'Cuenta',
		icon: I.phUserCircle,
		items: [
			['Plan y facturación', I.phCreditCard],
			['Usuarios y permisos', I.phUsers],
			['Dominio', I.phGlobe],
			['Notificaciones', I.phBell],
			['Mis datos', I.phDatabase]
		]
	}
] as const;

export const SIMPLE = [
	['Inicio', I.phHouseSimple],
	['Pedidos', I.phPackage, 7],
	['Productos', I.phTag],
	['WhatsApp', I.phWhatsappLogo, 3],
	['Clientes', I.phUsersThree],
	['Cobros', I.phWallet],
	['Envíos', I.phTruck],
	['Mi tienda', I.phStorefront],
	['Configuración', I.phGearSix],
	['Ayuda', I.phQuestion]
] as const;

export const KICONS = [I.phPackage, I.phReceipt, I.phFunnelSimple, I.phUserPlus] as const;
