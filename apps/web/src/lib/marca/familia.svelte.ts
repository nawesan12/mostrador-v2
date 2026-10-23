import { getContext, setContext } from 'svelte';
import { SvelteMap } from 'svelte/reactivity';

/** Una pieza que se puede exportar: la registra cada `Lamina` al montarse. */
export type Registro = { nombre: string; png: () => Promise<Blob> };

/**
 * Estado de una familia de piezas (papelería, redes…): el ancho disponible para mostrar las piezas
 * escaladas y la lista de piezas para "Descargar todo".
 */
export class Familia {
	rect = $state<DOMRectReadOnly | undefined>();
	/** Ancho útil del lienzo (sin el padding); 0 antes de medir. */
	ancho = $derived(this.rect?.width ?? 0);
	total = $state(0);
	#piezas = new SvelteMap<string, Registro>();

	constructor(readonly clave: string) {}

	registrar(r: Registro) {
		this.#piezas.set(r.nombre, r);
		this.total = this.#piezas.size;
		return () => {
			this.#piezas.delete(r.nombre);
			this.total = this.#piezas.size;
		};
	}

	/** En el orden del documento (el de montaje). */
	get piezas() {
		return [...this.#piezas.values()];
	}
}

const CLAVE = Symbol('familia-marca');
export const crearFamilia = (clave: string) => setContext(CLAVE, new Familia(clave));
export const usarFamilia = () => getContext<Familia>(CLAVE);
