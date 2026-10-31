/**
 * El tiempo real del panel (docs/decisiones.md §44): una conexión por pestaña, compartida por todo
 * el panel por contexto. Cada aviso invalida sólo lo que toca (las dependencias `tienda:<cosa>` de
 * los load), juntando los de una ráfaga; al reconectar, al pedirlo el servidor o al volver la
 * pestaña después de un rato se recarga todo (invalidateAll), así no se pierde nada.
 */
import { getContext, setContext } from 'svelte';
import { goto, invalidate, invalidateAll } from '$app/navigation';
import { resolve } from '$app/paths';
import type { UiState } from '#lib/ui.svelte.ts';
import { ConexionTR } from './conexion.svelte.ts';
import { claves, urlTiempoReal, type EventoTR } from './reglas.ts';
import { sonarPedido, sonidoPrendido } from './sonido.ts';

/** Lo que se junta antes de invalidar (una venta toca pedido, stock, cliente… a la vez). */
const RAFAGA = 300;
/** Cuánto se espera antes de avisar un pedido nuevo, por si es propio. */
const ESPERA_PROPIO = 1200;
/** Cuánto vale un «esto lo hice yo» para no avisarme mi propio pedido. */
const PROPIO_MS = 30_000;

type Oyente = (e: EventoTR) => void;

export class TiempoRealPanel {
	conexion: ConexionTR;
	#ui: UiState;
	// eslint-disable-next-line svelte/prefer-svelte-reactivity -- estado interno, no se dibuja
	#pendientes = new Set<string>();
	#rafaga: ReturnType<typeof setTimeout> | undefined;
	// eslint-disable-next-line svelte/prefer-svelte-reactivity -- estado interno, no se dibuja
	#oyentes = new Set<Oyente>();
	// eslint-disable-next-line svelte/prefer-svelte-reactivity -- estado interno, no se dibuja
	#propios = new Map<string, number>();
	#todoPendiente = false;

	constructor(url: string, ui: UiState) {
		this.#ui = ui;
		this.conexion = new ConexionTR(url, {
			alEvento: (e) => this.#alEvento(e),
			alResincronizar: () => this.#todo(),
			// Sin sesión: recargar lleva a ingresar (lo resuelve el panel como hoy).
			alSinSesion: () => void invalidateAll()
		});
	}

	iniciar() {
		this.conexion.iniciar();
	}

	detener() {
		clearTimeout(this.#rafaga);
		this.conexion.detener();
	}

	/** Escuchar los avisos desde una pantalla (la hoja de un pedido, la actividad…). */
	escuchar(fn: Oyente): () => void {
		this.#oyentes.add(fn);
		return () => this.#oyentes.delete(fn);
	}

	/** «Este pedido lo cargué yo» (punto de venta, Crear pedido): no se avisa ni suena. */
	propio(tipo: string, id: string | number) {
		this.#propios.set(`${tipo}:${id}`, Date.now());
	}

	#esPropio(e: EventoTR) {
		const k = `${e.tipo}:${e.id}`;
		const t = this.#propios.get(k);
		for (const [c, v] of this.#propios) if (Date.now() - v > PROPIO_MS) this.#propios.delete(c);
		return t !== undefined && Date.now() - t < PROPIO_MS;
	}

	#alEvento(e: EventoTR) {
		for (const fn of this.#oyentes) {
			try {
				fn(e);
			} catch {
				// Una pantalla con un error no corta los avisos de las demás.
			}
		}
		// El aviso espera un poco: la respuesta de la venta propia (que marca el número) puede llegar
		// después que el aviso del tiempo real.
		if (e.tipo === 'pedido.nuevo')
			setTimeout(() => {
				if (!this.#esPropio(e)) this.#avisarPedido(e);
			}, ESPERA_PROPIO);
		for (const c of claves(e)) this.#pendientes.add(c);
		clearTimeout(this.#rafaga);
		this.#rafaga = setTimeout(() => this.#invalidar(), RAFAGA);
	}

	#invalidar() {
		if (this.#todoPendiente) return;
		// Con la pestaña oculta no se recarga nada: al volver, la conexión pide ponerse al día.
		if (typeof document !== 'undefined' && document.hidden) {
			this.#pendientes.clear();
			this.#todoPendiente = true;
			const volver = () => {
				if (document.hidden) return;
				document.removeEventListener('visibilitychange', volver);
				this.#todo();
			};
			document.addEventListener('visibilitychange', volver);
			return;
		}
		// eslint-disable-next-line svelte/prefer-svelte-reactivity -- copia local
		const ks = new Set(this.#pendientes);
		this.#pendientes.clear();
		if (ks.size) void invalidate((u) => ks.has(u.href));
	}

	#todo() {
		this.#todoPendiente = false;
		this.#pendientes.clear();
		clearTimeout(this.#rafaga);
		for (const fn of this.#oyentes) {
			try {
				fn({ tipo: 'resincronizar' });
			} catch {
				// idem
			}
		}
		void invalidateAll();
	}

	#avisarPedido(e: EventoTR) {
		const n = e.lote ? null : e.id;
		this.#ui.toast({
			k: 'ok',
			title: n ? `Pedido nuevo #${n}` : 'Entraron pedidos nuevos',
			desc: n ? 'Recién entró. Ya está en Pedidos.' : 'Ya están en Pedidos.',
			action: 'Ver',
			onAction: () => {
				const ruta = resolve('/panel/pedidos');
				// eslint-disable-next-line svelte/no-navigation-without-resolve -- resolve() de arriba, con la query
				return goto(n ? `${ruta}?pedido=${encodeURIComponent(n)}` : ruta);
			}
		});
		if (sonidoPrendido()) sonarPedido();
	}
}

const CLAVE = Symbol('tiempo-real');

/** Dónde vive la conexión del panel (se crea al montar el armazón; en la demo queda vacía). */
export class LugarTiempoReal {
	actual = $state<TiempoRealPanel | null>(null);
}

/** Arma el tiempo real del panel (null en el servidor). */
export function crearTiempoReal(origen: string | null | undefined, ui: UiState) {
	if (typeof location === 'undefined') return null;
	return new TiempoRealPanel(urlTiempoReal(location, '/v1/tiempo-real', origen), ui);
}

export const setTiempoReal = (l: LugarTiempoReal) => setContext(CLAVE, l);
export const getTiempoReal = () => getContext<LugarTiempoReal | undefined>(CLAVE);

/**
 * Desde una pantalla del panel: `fn` recibe cada aviso mientras la pantalla está montada (y un
 * `{ tipo: 'resincronizar' }` cuando hay que ponerse al día con todo).
 */
export function escucharTiempoReal(fn: Oyente) {
	const lugar = getTiempoReal();
	$effect(() => {
		const tr = lugar?.actual;
		if (!tr) return;
		return tr.escuchar(fn);
	});
}
