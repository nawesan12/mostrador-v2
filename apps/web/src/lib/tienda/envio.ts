/**
 * Cotización de envío por código postal. Hasta conectar la API de Correo Argentino es una tabla por
 * región (precio base y días hábiles); el envío gratis sale de lo que fija la tienda. Andreani no se
 * ofrece: no es una integración. Gemelo en Go: vidriera.Cotizar.
 */
import type { OpcionEnvio, TiendaPublica } from './tipos.ts';

const REGIONES: [desde: number, hasta: number, base: number, dias: number][] = [
	[1000, 1999, 6900, 2], // CABA y GBA
	[2000, 3999, 7900, 3], // Santa Fe, Entre Ríos, Litoral
	[4000, 5999, 8900, 4], // NOA y Córdoba
	[6000, 7999, 7500, 3], // Buenos Aires interior
	[8000, 9499, 9900, 5], // Patagonia norte
	[9500, 9999, 11900, 6] // Patagonia sur
];

const DIAS = ['domingo', 'lunes', 'martes', 'miércoles', 'jueves', 'viernes', 'sábado'];
function llega(dias: number, desde = new Date()) {
	const d = new Date(desde);
	let habiles = 0;
	while (habiles < dias) {
		d.setDate(d.getDate() + 1);
		if (d.getDay() !== 0 && d.getDay() !== 6) habiles++;
	}
	return `Llega el ${DIAS[d.getDay()]} ${d.getDate()}`;
}

export const cpValido = (cp: string) => /^\d{4}$/.test(cp.trim());

export function cotizar(
	cp: string,
	subtotal: number,
	t: Pick<TiendaPublica, 'envios' | 'direccion'>,
	hoy = new Date()
): OpcionEnvio[] {
	const n = Number(cp.trim());
	const r = REGIONES.find(([a, b]) => n >= a && n <= b);
	const out: OpcionEnvio[] = [];
	const gratis = t.envios.gratisDesde !== null && subtotal >= t.envios.gratisDesde;
	if (r) {
		if (t.envios.correo)
			out.push({
				id: 'correo',
				nombre: 'Correo Argentino a sucursal',
				detalle: llega(r[3] + 2, hoy),
				precio: gratis ? 0 : Math.round((r[2] * 0.8) / 100) * 100
			});
		// La moto llega en el día dentro de CABA.
		if (t.envios.moto && n >= 1000 && n <= 1499)
			out.push({
				id: 'moto',
				nombre: 'Moto en el día',
				detalle: 'Si comprás antes de las 16 h',
				precio: 4500
			});
	}
	if (t.envios.retiro)
		out.push({
			id: 'retiro',
			nombre: 'Retiro en el local',
			detalle: t.direccion || 'Te avisamos cuando esté listo',
			precio: 0
		});
	return out;
}
