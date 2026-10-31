/**
 * «Volver a vender»: los productos de un pedido viejo como renglones de Crear pedido, con el precio y
 * el stock de hoy. Lo que cambió se avisa para revisarlo antes de guardar; nada se guarda solo.
 */
import { normalizarWhatsapp } from '#lib/whatsapp.ts';
import { pesos } from '../formato.ts';
import { vendibles, type VarianteFila } from '../vendibles.ts';

export type ItemVendido = {
	productoId: string | null;
	varianteId: string | null;
	nombre: string;
	/** "Talle M · Color Negro", como se vendió. */
	variante?: string | null;
	cantidad: number;
	precio: number;
};

type Producto = {
	id: string;
	nombre: string;
	sku: string;
	precio: number;
	stock: number;
	variantes?: VarianteFila[];
};

const igual = (a: string, b: string) => a.trim().toLowerCase() === b.trim().toLowerCase();

/**
 * Cada ítem se busca por su variante, después por el producto y, en los pedidos de antes de guardar
 * el producto, por el nombre. Sin stock no se suma; con menos stock, se suma lo que hay.
 */
export function volverAVender(items: ItemVendido[], productos: Producto[]) {
	const todos = vendibles(productos);
	const cantidades = new Map<string, number>();
	const avisos: string[] = [];
	for (const it of items) {
		const v = it.varianteId
			? todos.find((x) => x.varianteId === it.varianteId)
			: it.productoId
				? todos.find((x) => x.clave === it.productoId)
				: todos.find((x) => !x.varianteId && igual(x.nombre, it.nombre));
		if (!v) {
			const nombre = it.variante ? `${it.nombre} · ${it.variante}` : it.nombre;
			avisos.push(`«${nombre}» ya no está a la venta.`);
			continue;
		}
		const ya = cantidades.get(v.clave) ?? 0;
		const queda = v.stock - ya;
		if (queda <= 0) {
			avisos.push(`«${v.nombre}» no tiene stock.`);
			continue;
		}
		const q = Math.min(it.cantidad, queda);
		if (q < it.cantidad)
			avisos.push(`De «${v.nombre}» ${queda === 1 ? 'queda 1' : `quedan ${queda}`}: sumamos ${q}.`);
		if (v.precio !== it.precio)
			avisos.push(`«${v.nombre}» cambió de precio: ${pesos(it.precio)} → ${pesos(v.precio)}.`);
		cantidades.set(v.clave, ya + q);
	}
	return { renglones: [...cantidades].map(([clave, cantidad]) => ({ clave, cantidad })), avisos };
}

type Cliente = { id: string; nombre: string; telefono: string | null };

/** El cliente del pedido en la lista de clientes: por WhatsApp y, si no, por el nombre. */
export function clienteDe<C extends Cliente>(
	comprador: { nombre: string; telefono: string | null },
	clientes: C[]
): C | null {
	const tel = comprador.telefono ? normalizarWhatsapp(comprador.telefono) : '';
	return (
		(tel && clientes.find((c) => c.telefono && normalizarWhatsapp(c.telefono) === tel)) ||
		clientes.find((c) => igual(c.nombre, comprador.nombre)) ||
		null
	);
}
