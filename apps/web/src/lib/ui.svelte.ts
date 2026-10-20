import { getContext, setContext } from 'svelte';
import type { ClaveMarca } from '#lib/marcas.ts';

export type Toast = {
	id: number;
	/** Tipo: define el ícono y el color de la barra (ok, warn, err, info, ai, wa, mp, andreani…). */
	k: ClaveMarca;
	title: string;
	desc?: string;
	/** Acción útil ("Deshacer", "Ver pedido"…). Al tocarla, se ejecuta y el toast se va. */
	action?: string;
	onAction?: () => void | Promise<void>;
};

/** Vida de un toast (handoff: barra que se vacía en 5,5 s). */
export const VIDA_TOAST = 5500;
/** Máximo de toasts a la vez (handoff: "máximo 3"). */
const MAX_TOASTS = 3;

/**
 * Estado de UI por pestaña (va por contexto para no compartir estado entre requests en SSR),
 * igual que en Caja.
 */
export class UiState {
	toasts = $state<Toast[]>([]);
	#seq = 0;
	// eslint-disable-next-line svelte/prefer-svelte-reactivity -- timers internos, no se dibujan
	#timers = new Map<number, ReturnType<typeof setTimeout>>();

	toast(t: Omit<Toast, 'id'>) {
		const id = ++this.#seq;
		this.toasts = [...this.toasts, { id, ...t }].slice(-MAX_TOASTS);
		this.#timers.set(
			id,
			setTimeout(() => this.dismiss(id), VIDA_TOAST)
		);
		return id;
	}

	/** Toast de "hecho" con Deshacer: el patrón de toda acción que cambia estado. */
	undoable(title: string, desc: string | undefined, undo: () => void | Promise<void>) {
		return this.toast({ k: 'ok', title, desc, action: 'Deshacer', onAction: undo });
	}

	async act(id: number) {
		const t = this.toasts.find((x) => x.id === id);
		this.dismiss(id);
		await t?.onAction?.();
	}

	dismiss(id: number) {
		clearTimeout(this.#timers.get(id));
		this.#timers.delete(id);
		this.toasts = this.toasts.filter((t) => t.id !== id);
	}
}

const KEY = Symbol('ui');
export const setUi = (ui: UiState) => setContext(KEY, ui);
export const getUi = () => getContext<UiState>(KEY);
