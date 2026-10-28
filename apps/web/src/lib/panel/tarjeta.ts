/**
 * La tarjeta A6 para los paquetes (pieza 7a de Imprimibles): la tienda elige el modelo del frente,
 * los textos con variables que se completan por pedido y el QR del dorso. Lógica pura: la usan la
 * página de la tarjeta, la impresión y el servidor (para validar lo que se guarda).
 */
export type Modelo = 'arco' | 'sello' | 'carta';
export type DestinoQr = 'resena' | 'cupon' | 'instagram' | 'whatsapp';

export type Tarjeta = {
	modelo: Modelo;
	titulo: string;
	mensaje: string;
	dorsoTitulo: string;
	dorsoTexto: string;
	qr: DestinoQr;
	/** El código del cupón cuando el QR lleva a "volver a comprar con un cupón". */
	cupon: string;
	/** La cuenta de Instagram, sin @, cuando el QR lleva ahí. */
	instagram: string;
};

export const MODELOS: [Modelo, string][] = [
	['arco', 'Arco con foto'],
	['sello', 'Color de la tienda'],
	['carta', 'Carta']
];
export const DESTINOS: [DestinoQr, string][] = [
	['resena', 'Dejar una reseña'],
	['cupon', 'Volver con un cupón'],
	['instagram', 'Seguir en Instagram'],
	['whatsapp', 'WhatsApp de la tienda']
];

export const TOPES = { titulo: 40, mensaje: 140, dorsoTitulo: 40, dorsoTexto: 140 } as const;

export const VARIABLES = [
	['{nombre}', 'Nombre de quien compra'],
	['{fecha}', 'Fecha del pedido'],
	['{productos}', 'Lo que compró'],
	['{pedido}', 'Número de pedido']
] as const;

export const TARJETA_INICIAL: Tarjeta = {
	modelo: 'arco',
	titulo: '¡Gracias, {nombre}!',
	mensaje: 'Lo preparamos a mano el {fecha}. Que lo disfrutes.',
	dorsoTitulo: '¿Te gustó?',
	dorsoTexto: 'Dejanos una reseña con foto: nos ayuda un montón a seguir creciendo.',
	qr: 'resena',
	cupon: '',
	instagram: ''
};

/** Lo que va en las variables de un pedido. */
export type Variables = { nombre: string; fecha: Date; productos: string[]; pedido: number };

const MESES = [
	'enero',
	'febrero',
	'marzo',
	'abril',
	'mayo',
	'junio',
	'julio',
	'agosto',
	'septiembre',
	'octubre',
	'noviembre',
	'diciembre'
];

/** "Lucía" de "Lucía Pérez"; "20 de octubre"; "Mate y bombilla" (y "y 2 más"). */
export function completar(texto: string, v: Variables) {
	// En hora de Argentina, aunque se imprima desde otra zona horaria.
	const [mes, dia] = new Intl.DateTimeFormat('en-CA', {
		timeZone: 'America/Argentina/Buenos_Aires',
		month: 'numeric',
		day: 'numeric'
	})
		.formatToParts(v.fecha)
		.filter((x) => x.type === 'month' || x.type === 'day')
		.map((x) => Number(x.value));
	const fecha = `${dia} de ${MESES[mes - 1]}`;
	const productos =
		v.productos.length <= 2
			? v.productos.join(' y ')
			: `${v.productos.slice(0, 2).join(', ')} y ${v.productos.length - 2} más`;
	return texto
		.replaceAll('{nombre}', v.nombre.trim().split(/\s+/)[0] || 'vos')
		.replaceAll('{fecha}', fecha)
		.replaceAll('{productos}', productos)
		.replaceAll('{pedido}', `#${v.pedido}`);
}

/** El ejemplo de la vista previa (el pedido de Lucía de Imprimibles). */
export const EJEMPLO: Variables = {
	nombre: 'Lucía Pérez',
	fecha: new Date('2026-10-20T12:00:00-03:00'),
	productos: ['Mate de calabaza', 'Bombilla de alpaca'],
	pedido: 1048
};

/** Adónde lleva el QR del dorso (null si falta el dato: cupón o Instagram sin cargar, sin WhatsApp). */
export function linkQr(t: Tarjeta, tienda: { url: string; whatsapp: string }) {
	if (t.qr === 'cupon') return t.cupon ? `${tienda.url}/c/${t.cupon}` : null;
	if (t.qr === 'instagram') return t.instagram ? `https://instagram.com/${t.instagram}` : null;
	if (t.qr === 'whatsapp') {
		const n = tienda.whatsapp.replace(/\D/g, '');
		return n ? `https://wa.me/${n.startsWith('54') ? n : `54${n}`}` : null;
	}
	return `${tienda.url}#resenas`;
}

/** Revisa una tarjeta como la guarda el servidor. Devuelve el error para mostrar, o null. */
export function errorTarjeta(t: Tarjeta): string | null {
	if (!MODELOS.some(([m]) => m === t.modelo)) return 'Elegí un modelo.';
	if (!DESTINOS.some(([d]) => d === t.qr)) return 'Elegí adónde lleva el QR.';
	for (const k of Object.keys(TOPES) as (keyof typeof TOPES)[])
		if (t[k].length > TOPES[k]) return `El texto es más largo que ${TOPES[k]} caracteres.`;
	if (!t.titulo.trim()) return 'Poné un título.';
	if (!/^[A-Z0-9-]{0,20}$/.test(t.cupon)) return 'Ese cupón no existe.';
	if (!/^[a-zA-Z0-9._]{0,30}$/.test(t.instagram)) return 'Revisá la cuenta de Instagram.';
	return null;
}
