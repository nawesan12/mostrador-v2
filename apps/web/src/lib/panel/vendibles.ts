/**
 * Lo que se puede vender desde el panel (punto de venta y crear pedido): un producto sin opciones es
 * un renglón; uno con opciones, un renglón por cada variante activa ("Remera · Talle M · Color
 * Negro"), con su precio y su stock. Así la grilla y el ticket no cambian y el pedido lleva su
 * `varianteId`.
 */
import { precioDe } from '#lib/tienda/variantes.ts';

/** La variante como la trae la lista de productos (igual en la API Go y en el sandbox). */
export type VarianteFila = {
	id: string;
	texto: string;
	sku: string;
	precio: number | null;
	stock: number;
	activa: boolean;
};

type ProductoVendible = {
	id: string;
	nombre: string;
	sku: string;
	precio: number;
	stock: number;
	variantes?: VarianteFila[];
};

/** El producto con el nombre, SKU, precio y stock de la variante (y su `clave` de renglón). */
export type Vendible<P> = P & {
	/** Clave del renglón: el id del producto, o `producto:variante`. */
	clave: string;
	varianteId: string | null;
	/** "Talle M · Color Negro" (vacío sin opciones). */
	variante: string;
};

export function vendibles<P extends ProductoVendible>(productos: P[]): Vendible<P>[] {
	return productos.flatMap((p): Vendible<P>[] => {
		if (!p.variantes?.length) return [{ ...p, clave: p.id, varianteId: null, variante: '' }];
		return p.variantes
			.filter((v) => v.activa)
			.map((v) => ({
				...p,
				clave: `${p.id}:${v.id}`,
				varianteId: v.id,
				variante: v.texto,
				nombre: `${p.nombre} · ${v.texto}`,
				sku: v.sku || p.sku,
				precio: precioDe(v, p.precio),
				stock: v.stock
			}));
	});
}
