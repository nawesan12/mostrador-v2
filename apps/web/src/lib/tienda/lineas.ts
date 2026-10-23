/**
 * Las líneas del carrito de la tienda, sin estado: una por producto y variante. El carrito
 * (carrito.svelte.ts) las guarda en el navegador; el servidor las vuelve a validar al comprar.
 */
import type { ProductoPublico } from './tipos.ts';
import { precioDe, textoVariante } from './variantes.ts';

export type Linea = { id: string; varianteId: string | null; cantidad: number };

export const MAX_LINEAS = 50;

/** Identifica la línea: el producto y, si tiene, la variante. */
export const claveLinea = (l: { id: string; varianteId?: string | null }) =>
	l.varianteId ? `${l.id}~${l.varianteId}` : l.id;

/**
 * Lo guardado en el navegador, revisado. Los carritos de antes de las variantes no tenían
 * `varianteId`: sus líneas siguen valiendo como producto sin variante.
 */
export function migrarLineas(g: unknown): Linea[] {
	if (!Array.isArray(g)) return [];
	const vistas = new Set<string>();
	const lineas: Linea[] = [];
	for (const x of g as Partial<Linea>[]) {
		if (!x || typeof x.id !== 'string' || !x.id || x.id.length > 80) continue;
		const cantidad = Number(x.cantidad);
		if (!Number.isInteger(cantidad) || cantidad < 1) continue;
		const l: Linea = {
			id: x.id,
			varianteId: typeof x.varianteId === 'string' && x.varianteId ? x.varianteId : null,
			cantidad: Math.min(99, cantidad)
		};
		if (vistas.has(claveLinea(l))) continue;
		vistas.add(claveLinea(l));
		lineas.push(l);
		if (lineas.length >= MAX_LINEAS) break;
	}
	return lineas;
}

/** Una línea con su producto tal como se compra: precio, foto y stock de la variante. */
export type LineaDetalle = Linea & {
	clave: string;
	/** "Talle M · Negro" (null sin variante). */
	variante: string | null;
	producto: ProductoPublico;
	subtotal: number;
};

/**
 * Las líneas con su producto. Se descartan las de productos que ya no están, las que piden una
 * variante que no existe (o ya no está activa) y las de productos con opciones sin variante.
 */
export function detalleLineas(lineas: Linea[], productos: ProductoPublico[]): LineaDetalle[] {
	const porId = new Map(productos.map((p) => [p.id, p]));
	return lineas.flatMap((l): LineaDetalle[] => {
		const p = porId.get(l.id);
		if (!p) return [];
		const opciones = p.opciones ?? [];
		if (!opciones.length) {
			if (l.varianteId) return [];
			return [
				{ ...l, clave: claveLinea(l), variante: null, producto: p, subtotal: p.precio * l.cantidad }
			];
		}
		const v = (p.variantes ?? []).find((x) => x.id === l.varianteId);
		if (!v) return [];
		const precio = precioDe(v, p.precio);
		return [
			{
				...l,
				clave: claveLinea(l),
				variante: textoVariante(opciones, v.valores),
				producto: { ...p, precio, anterior: null, foto: v.foto ?? p.foto, stock: v.stock },
				subtotal: precio * l.cantidad
			}
		];
	});
}
