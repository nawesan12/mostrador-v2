/**
 * Las formas de entrega para un código postal. Correo Argentino sale sólo de la cotización real de
 * la API MiCorreo (`correo`, la pide envio.svelte.ts): sin cotización, no se ofrece, nunca con un
 * precio inventado. La moto (CABA) y el retiro son de la tienda; el envío gratis sale de lo que fija
 * la tienda. Andreani no se ofrece: no es una integración. Gemelo en Go: vidriera.Cotizar y
 * vidriera.envioCorreo.
 */
import type { OpcionEnvio, TiendaPublica } from './tipos.ts';

/** Lo que cotizó Correo Argentino para un CP (components/schemas/CotizacionCorreo). */
export type OpcionCorreo = { precio: number; producto: string; diasMin?: string; diasMax?: string };
export type CotizacionCorreo = { domicilio: OpcionCorreo | null; sucursal: OpcionCorreo | null };

export const cpValido = (cp: string) => /^\d{4}$/.test(cp.trim());

export const esCorreo = (id: string | null | undefined) =>
	id === 'correo-domicilio' || id === 'correo-sucursal';

/** "Llega en 2 a 5 días hábiles" si Correo informa el plazo; si no, el producto (Paq.ar Clásico). */
function detalleCorreo(o: OpcionCorreo, donde: string) {
	const [a, b] = [o.diasMin?.trim(), o.diasMax?.trim()];
	if (a && b && a !== b) return `${donde} · llega en ${a} a ${b} días hábiles`;
	if (a || b) return `${donde} · llega en ${a || b} días hábiles`;
	return `${donde} · ${o.producto || 'Paq.ar'}`;
}

export function cotizar(
	cp: string,
	subtotal: number,
	t: Pick<TiendaPublica, 'envios' | 'direccion'>,
	correo: CotizacionCorreo | null = null
): OpcionEnvio[] {
	const n = Number(cp.trim());
	const out: OpcionEnvio[] = [];
	const gratis = t.envios.gratisDesde !== null && subtotal >= t.envios.gratisDesde;
	if (t.envios.correo && correo && cpValido(cp)) {
		if (correo.domicilio)
			out.push({
				id: 'correo-domicilio',
				nombre: 'Correo Argentino a domicilio',
				detalle: detalleCorreo(correo.domicilio, 'Te lo lleva el cartero'),
				precio: gratis ? 0 : correo.domicilio.precio
			});
		if (correo.sucursal)
			out.push({
				id: 'correo-sucursal',
				nombre: 'Correo Argentino a sucursal',
				detalle: detalleCorreo(correo.sucursal, 'Retirás en la sucursal que elijas'),
				precio: gratis ? 0 : correo.sucursal.precio
			});
	}
	// La moto llega en el día dentro de CABA.
	if (cpValido(cp) && t.envios.moto && n >= 1000 && n <= 1499)
		out.push({
			id: 'moto',
			nombre: 'Moto en el día',
			detalle: 'Si comprás antes de las 16 h',
			precio: 4500
		});
	if (t.envios.retiro)
		out.push({
			id: 'retiro',
			nombre: 'Retiro en el local',
			detalle: t.direccion || 'Te avisamos cuando esté listo',
			precio: 0
		});
	return out;
}
