/**
 * Lo del carrito que ya no se puede comprar como está: se agotó, quedan menos o ya no se vende.
 * Con el tiempo real (docs/decisiones.md §44) se le avisa a quien compra antes de pagar, en vez de
 * enterarse por el error del checkout.
 */
import { detalleLineas, claveLinea, type Linea } from './lineas.ts';
import { textoVariante } from './variantes.ts';
import type { ProductoPublico } from './tipos.ts';

export type Faltante = {
	clave: string;
	nombre: string;
	/** Cuántas hay ahora (0: agotado o ya no se vende). */
	quedan: number;
	pedidas: number;
};

export function faltantes(lineas: Linea[], productos: ProductoPublico[]): Faltante[] {
	const vivas = new Map(detalleLineas(lineas, productos).map((d) => [d.clave, d]));
	const porId = new Map(productos.map((p) => [p.id, p]));
	return lineas.flatMap((l): Faltante[] => {
		const clave = claveLinea(l);
		const d = vivas.get(clave);
		if (!d) {
			// El producto (o la variante) ya no está publicado.
			const p = porId.get(l.id);
			const v = p?.variantes.find((x) => x.id === l.varianteId);
			const nombre = p
				? v
					? `${p.nombre} (${textoVariante(p.opciones, v.valores)})`
					: p.nombre
				: 'Un producto';
			return [{ clave, nombre, quedan: 0, pedidas: l.cantidad }];
		}
		if (d.producto.stock >= l.cantidad) return [];
		const nombre = d.variante ? `${d.producto.nombre} (${d.variante})` : d.producto.nombre;
		return [{ clave, nombre, quedan: Math.max(0, d.producto.stock), pedidas: l.cantidad }];
	});
}

/** El aviso en castellano: «Se agotó X» o «De X quedan 2». */
export function textoFaltantes(fs: Faltante[]): { titulo: string; detalle: string } {
	const agotados = fs.filter((f) => f.quedan === 0);
	const titulo =
		fs.length === 1
			? agotados.length
				? `Se agotó ${fs[0].nombre}`
				: `De ${fs[0].nombre} quedan ${fs[0].quedan}`
			: 'Cambió el stock de tu carrito';
	const detalle =
		fs.length === 1
			? agotados.length
				? 'Ya no queda. Sacalo del carrito para seguir con la compra.'
				: `Tenés ${fs[0].pedidas} en el carrito: ajustá la cantidad para seguir.`
			: `${fs.map((f) => (f.quedan ? `${f.nombre}: quedan ${f.quedan}` : `${f.nombre}: agotado`)).join(' · ')}.`;
	return { titulo, detalle };
}
