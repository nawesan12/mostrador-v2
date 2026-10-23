/**
 * El carrito de la tienda online: vive en el navegador (por tienda) y el servidor lo valida al
 * cerrar la compra (precios y stock salen de la base, nunca de acá).
 */
import { browser } from '$app/env';
import type { ProductoPublico } from './tipos.ts';

export type Linea = { id: string; cantidad: number };

const MAX_LINEAS = 50;

export class Carrito {
	lineas = $state<Linea[]>([]);
	#clave: string;

	constructor(slug: string) {
		this.#clave = `carrito:${slug}`;
		if (!browser) return;
		try {
			const g = JSON.parse(localStorage.getItem(this.#clave) ?? '[]') as Linea[];
			this.lineas = Array.isArray(g)
				? g.filter((l) => typeof l.id === 'string' && l.cantidad > 0).slice(0, MAX_LINEAS)
				: [];
		} catch {
			this.lineas = [];
		}
	}

	#guardar() {
		try {
			localStorage.setItem(this.#clave, JSON.stringify(this.lineas));
		} catch {
			// Sin almacenamiento (modo privado): el carrito dura lo que la página.
		}
	}

	get unidades() {
		return this.lineas.reduce((s, l) => s + l.cantidad, 0);
	}

	sumar(id: string, cantidad = 1, tope = 99) {
		const l = this.lineas.find((x) => x.id === id);
		if (l) l.cantidad = Math.min(tope, l.cantidad + cantidad);
		else if (this.lineas.length < MAX_LINEAS)
			this.lineas.push({ id, cantidad: Math.min(tope, cantidad) });
		this.#guardar();
	}

	fijar(id: string, cantidad: number) {
		this.lineas =
			cantidad > 0
				? this.lineas.map((l) => (l.id === id ? { ...l, cantidad } : l))
				: this.lineas.filter((l) => l.id !== id);
		this.#guardar();
	}

	vaciar() {
		this.lineas = [];
		this.#guardar();
	}

	/** Las líneas con su producto (las de productos que ya no están se descartan). */
	detalle(productos: ProductoPublico[]) {
		// eslint-disable-next-line svelte/prefer-svelte-reactivity -- local, no se modifica
		const porId = new Map(productos.map((p) => [p.id, p]));
		return this.lineas.flatMap((l) => {
			const p = porId.get(l.id);
			return p ? [{ ...l, producto: p, subtotal: p.precio * l.cantidad }] : [];
		});
	}
}

// eslint-disable-next-line svelte/prefer-svelte-reactivity -- registro de carritos, no es estado
const carritos = new Map<string, Carrito>();
/** Un carrito por tienda (y por pestaña). */
export function carritoDe(slug: string) {
	let c = carritos.get(slug);
	if (!c) carritos.set(slug, (c = new Carrito(slug)));
	return c;
}
