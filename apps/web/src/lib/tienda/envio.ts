/**
 * Las formas de entrega para un código postal. Correo Argentino sale sólo de la cotización real de
 * la API MiCorreo (`correo`, la pide envio.svelte.ts): sin cotización, no se ofrece, nunca con un
 * precio inventado. La moto y el retiro son de la tienda (Configuración › Envíos y retiro): la moto
 * con su costo fijo y la zona que puso, y el retiro con su dirección y horario. El envío gratis sale
 * de la promoción automática y vale también para la moto. Andreani no se ofrece: no es una
 * integración. Gemelo en Go: vidriera.Cotizar y vidriera.envioCorreo.
 */
import type { EnviosTienda, OpcionEnvio, TiendaPublica } from './tipos.ts';

/** Lo que cotizó Correo Argentino para un CP (components/schemas/CotizacionCorreo). */
export type OpcionCorreo = { precio: number; producto: string; diasMin?: string; diasMax?: string };
export type CotizacionCorreo = { domicilio: OpcionCorreo | null; sucursal: OpcionCorreo | null };

export const cpValido = (cp: string) => /^\d{4}$/.test(cp.trim());

/** Junta las partes que vinieron con « · » (nada de separadores colgando). */
export const unir = (...partes: (string | null | undefined)[]) =>
	partes
		.map((p) => p?.trim())
		.filter(Boolean)
		.join(' · ');

/** «Llega a Palermo y Belgrano · En el día»: lo que puso la tienda para la moto. */
export const detalleMoto = (e: Pick<EnviosTienda, 'motoZona' | 'motoDemora'>) =>
	unir(e.motoZona.trim() && `Llega a ${e.motoZona.trim()}`, e.motoDemora);

/** «$ 3.500 · Palermo y Belgrano»: el costo y la zona de la moto, para la portada de la tienda. */
export const resumenMoto = (e: Pick<EnviosTienda, 'motoCosto' | 'motoZona'>) =>
	unir(
		e.motoCosto > 0 ? '$ ' + Math.round(e.motoCosto).toLocaleString('es-AR') : 'Sin costo',
		e.motoZona
	);

/** «Gorriti 4500 · Lun a vie de 10 a 19»: dónde y cuándo se retira. */
export const detalleRetiro = (e: Pick<EnviosTienda, 'retiroDireccion' | 'retiroHorario'>) =>
	unir(e.retiroDireccion, e.retiroHorario);

export const esCorreo = (id: string | null | undefined) =>
	id === 'correo-domicilio' || id === 'correo-sucursal';

/** "Llega en 2 a 5 días hábiles" si Correo informa el plazo; si no, el producto (Paq.ar Clásico). */
function detalleCorreo(o: OpcionCorreo, donde: string) {
	const [a, b] = [o.diasMin?.trim(), o.diasMax?.trim()];
	if (a && b && a !== b) return `${donde} · llega en ${a} a ${b} días hábiles`;
	if (a || b)
		return `${donde} · llega en ${a || b} ${String(a || b) === '1' ? 'día hábil' : 'días hábiles'}`;
	return `${donde} · ${o.producto || 'Paq.ar'}`;
}

export function cotizar(
	cp: string,
	subtotal: number,
	t: Pick<TiendaPublica, 'envios'>,
	correo: CotizacionCorreo | null = null
): OpcionEnvio[] {
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
	// La moto va a una dirección: pide el CP. Llega a la zona que puso la tienda.
	if (cpValido(cp) && t.envios.moto)
		out.push({
			id: 'moto',
			nombre: 'Envío en moto',
			detalle: detalleMoto(t.envios),
			precio: gratis ? 0 : t.envios.motoCosto
		});
	if (t.envios.retiro)
		out.push({
			id: 'retiro',
			nombre: 'Retiro en el local',
			detalle: detalleRetiro(t.envios) || 'Te avisamos cuando esté listo',
			precio: 0
		});
	return out;
}
