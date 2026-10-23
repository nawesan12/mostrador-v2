/**
 * Datos de las piezas de marca (/marca). Nada escrito a mano si ya existe en el código: los datos
 * legales salen de `legales.ts`, los planes de `panel/planes.ts` y la comparación con Tiendanube de
 * la landing, así una pieza nunca dice algo distinto del sitio.
 */
import { CMP } from '#lib/landing/datos.ts';
import { fmt } from '#lib/landing/formato.ts';
import { LEGALES } from '#lib/legales.ts';
import { PLANES, precioAnual } from '#lib/panel/planes.ts';

export { fmt };

export const DOMINIO = 'mostradorapp.com.ar';
export const SITIO = `https://${DOMINIO}`;
export const MAIL = {
	hola: `hola@${DOMINIO}`,
	ayuda: `ayuda@${DOMINIO}`,
	legales: `legales@${DOMINIO}`
} as const;

/** "Nahuel Santillán · CUIT 20-44859234-1 · Mar del Plata, provincia de Buenos Aires · …" */
function datosDelProveedor() {
	for (const legal of LEGALES)
		for (const b of legal.bloques) {
			if (b.t !== 'seccion') continue;
			for (const c of b.cuerpo) {
				if (!('p' in c)) continue;
				const m = c.p.match(/^(.+?) · CUIT ([\d-]+) · (.+?) · /);
				if (m) return { titular: m[1], cuit: m[2], domicilio: m[3] };
			}
		}
	throw new Error('legales.ts ya no trae los datos del proveedor');
}

const p = datosDelProveedor();
const [ciudad, ...resto] = p.domicilio.split(', ');
const region = resto.join(', ');
export const EMPRESA = {
	...p,
	/** "Mar del Plata" */
	ciudad,
	/** "Provincia de Buenos Aires" (para la segunda línea de una dirección). */
	region: region.charAt(0).toUpperCase() + region.slice(1),
	/** Para el encabezado de las cartas: "Mar del Plata, 20 de octubre de 2026". */
	ciudadCarta: ciudad,
	cuitNumero: Number(p.cuit.replaceAll('-', ''))
};

/** La prueba gratis de todos los planes (landing, ayuda y términos). */
export const PRUEBA_DIAS = 14;
/** Pagando por año (`precioAnual`). */
export const DESCUENTO_ANUAL = Math.round(100 - precioAnual(100));

const plan = (id: (typeof PLANES)[number]['id']) => PLANES.find((x) => x.id === id)!;
export const PLAN = { emprende: plan('emprende'), pro: plan('pro'), negocio: plan('negocio') };
/** El plan pago más barato: "Desde $ 19.900 por mes". */
export const DESDE = Math.min(...PLANES.filter((x) => x.precio > 0).map((x) => x.precio));

/**
 * Cada plan pago contra su par de Tiendanube, como la tabla de la landing (`CMP`):
 * "Emprende · $ 19.900" | "Esencial · $ 27.999".
 */
export const VERSUS = CMP.slice(1, 4).map(([, nuestro, ellos]) => {
	const [nombre] = nuestro.split(' · ');
	const mio = PLANES.find((x) => x.nombre === nombre);
	const precio = Number(ellos.replace(/\D/g, ''));
	if (!mio || Number(nuestro.replace(/\D/g, '')) !== mio.precio)
		throw new Error(`La comparación de la landing no coincide con los planes: ${nuestro}`);
	return {
		plan: mio,
		competencia: {
			nombre: `Tiendanube ${ellos.split(' · ')[0]}`,
			plan: ellos.split(' · ')[0],
			precio
		},
		descuento: Math.round((1 - mio.precio / precio) * 100),
		// El mismo cálculo que "Ahorrás por año" en la calculadora de la landing.
		ahorroAnual: (precio - mio.precio) * 12
	};
});
/** Pro contra Tiendanube Impulso: el plan "más elegido". */
export const VS_PRO = VERSUS.find((v) => v.plan.id === 'pro')!;
export const AHORRO_MAXIMO = Math.max(...VERSUS.map((v) => v.ahorroAnual));

const zona = 'America/Argentina/Buenos_Aires';
/** "23 de octubre de 2026" */
export const fechaLarga = (d = new Date()) =>
	d.toLocaleDateString('es-AR', { day: 'numeric', month: 'long', year: 'numeric', timeZone: zona });
/** "2026-10-23", en hora de Argentina. */
export const fechaIso = (d = new Date()) => d.toLocaleDateString('en-CA', { timeZone: zona });
/**
 * El año del próximo Hot Sale (se hace en mayo): pasado mayo, la pieza ya anuncia el del año que
 * viene.
 */
export function anioHotSale(d = new Date()) {
	const [anio, mes] = fechaIso(d).split('-').map(Number);
	return mes > 5 ? anio + 1 : anio;
}
/** "23/10/2026" */
export const fechaCorta = (d = new Date()) =>
	d.toLocaleDateString('es-AR', {
		day: '2-digit',
		month: '2-digit',
		year: 'numeric',
		timeZone: zona
	});
/** "octubre 2026" */
export const mesAnio = (d = new Date()) =>
	d
		.toLocaleDateString('es-AR', { month: 'long', year: 'numeric', timeZone: zona })
		.replace(' de ', ' ');

/** Links de los QR: a rutas que existen, con utm para saber de qué pieza vino cada visita. */
export const QR = {
	carpeta: `${SITIO}/?utm_source=carpeta`,
	flyer: `${SITIO}/onboarding?utm_source=flyer`,
	flyerJugada: `${SITIO}/onboarding?utm_source=flyer-jugada`,
	onepager: `${SITIO}/onboarding?utm_source=onepager`,
	rollup: `${SITIO}/onboarding?utm_source=rollup`,
	triptico: `${SITIO}/contacto?utm_source=triptico`,
	paquete: `${SITIO}/onboarding?utm_source=tarjeta-paquete`
} as const;

/** vCard 3.0 para el QR de la tarjeta personal. */
export function vcard(p: { nombre: string; cargo: string; telefono: string; mail: string }) {
	const esc = (s: string) => s.replace(/[\\,;]/g, (c) => '\\' + c).replace(/\n/g, ' ');
	return [
		'BEGIN:VCARD',
		'VERSION:3.0',
		`FN:${esc(p.nombre)}`,
		'ORG:Mostrador',
		`TITLE:${esc(p.cargo)}`,
		`TEL;TYPE=CELL:${esc(p.telefono)}`,
		`EMAIL:${esc(p.mail)}`,
		`URL:${SITIO}`,
		'END:VCARD'
	].join('\n');
}

/**
 * QR de ARCA de la factura modelo: el formato real (https://www.afip.gob.ar/fe/qr/) con datos de
 * muestra y código de autorización en cero, para que nadie lo tome por un comprobante válido.
 */
export function qrArcaMuestra(importe: number, fecha: string) {
	const datos = {
		ver: 1,
		fecha,
		cuit: EMPRESA.cuitNumero,
		ptoVta: 1,
		tipoCmp: 6,
		nroCmp: 0,
		importe,
		moneda: 'PES',
		ctz: 1,
		tipoDocRec: 99,
		nroDocRec: 0,
		tipoCodAut: 'E',
		codAut: 0
	};
	const b64 = btoa(JSON.stringify(datos));
	return `https://www.afip.gob.ar/fe/qr/?p=${b64}`;
}

export type Persona = {
	nombre: string;
	cargo: string;
	telefono: string;
	mail: string;
	/** URL local (blob:) de la foto elegida, o vacío para mostrar las iniciales. */
	foto: string;
};
export const PERSONA_INICIAL: Persona = {
	nombre: 'Tu nombre',
	cargo: 'Tu cargo',
	telefono: 'Tu teléfono',
	mail: `tu.nombre@${DOMINIO}`,
	foto: ''
};
export const iniciales = (nombre: string) =>
	nombre
		.split(/\s+/)
		.filter(Boolean)
		.slice(0, 2)
		.map((x) => x[0]!.toUpperCase())
		.join('');

export type Evento = {
	titulo: string;
	ciudad: string;
	fecha: string;
	lugar: string;
	bajada: string;
	condiciones: string;
	link: string;
};
export const EVENTO_INICIAL: Evento = {
	titulo: 'Nombre del evento',
	ciudad: 'Ciudad',
	fecha: 'Fecha',
	lugar: 'Lugar del evento',
	bajada: 'Contá en una línea de qué se trata y para quién es.',
	condiciones: 'Entrada libre · cupos limitados',
	link: SITIO
};

/** Une las dos últimas palabras con un espacio duro: ningún párrafo termina con una palabra sola. */
export const sinViuda = (t: string) => t.replace(/ (\S+)$/, ' $1');
