/**
 * Lo que necesita la tarjeta A6 de la tienda: la tarjeta guardada, la marca (nombre, colores y
 * tipografía de la plantilla), adónde lleva el QR y una foto para el modelo con arco.
 */
import { plantilla } from '#lib/plantillas.ts';
import type { Datos } from './datos/tipos.ts';

export async function contextoTarjeta(datos: Datos, origen: string) {
	const [tarjeta, tienda, conf, mi] = await Promise.all([
		datos.tarjeta(),
		datos.tienda(),
		datos.configuracion(),
		datos.miTienda()
	]);
	const p = plantilla(tienda.plantilla) ?? plantilla('ceibo')!;
	return {
		tarjeta,
		tienda: {
			nombre: tienda.nombre,
			url: `${origen}/t/${tienda.slug}`,
			whatsapp: conf.negocio.whatsapp,
			dominio: mi.direccion
		},
		paleta: { bg: p.bg, tx: p.tx, ac: p.ac, ff: p.ff, oscura: p.oscura },
		foto: mi.vista.portada.foto ?? mi.vista.productos[0]?.foto ?? null
	};
}
