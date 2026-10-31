/**
 * Una conexión de tiempo real (docs/decisiones.md §44): abre el WebSocket, reconecta con espera
 * exponencial y azar, y avisa cuando hay que ponerse al día (al volver de un corte, cuando el
 * servidor pide resincronizar o al volver la pestaña a primer plano). Sin WebSocket todo sigue
 * andando como antes: esto sólo adelanta lo que se vería al recargar.
 */
import { espera, leerEvento, trasCierre, type EventoTR } from './reglas.ts';

export type EstadoConexion = 'conectando' | 'conectada' | 'reconectando' | 'parada';

/** Sin ningún mensaje (ni el latido de cada 25 s) en este tiempo, la conexión se da por muerta. */
const SILENCIO_MAXIMO = 65_000;
/** Una pestaña que estuvo oculta más que esto se pone al día al volver. */
const OCULTA_MUCHO = 15_000;

export type OpcionesConexion = {
	/** Cada aviso del servidor (sin los de control). */
	alEvento: (e: EventoTR) => void;
	/** Hay que recargar todo: se pudo haber perdido algo. */
	alResincronizar: () => void;
	/** El servidor dijo que no hay sesión: no se reintenta. */
	alSinSesion?: () => void;
};

export class ConexionTR {
	estado = $state<EstadoConexion>('conectando');
	/** Desde cuándo no hay conexión (null si está andando): el aviso aparece si dura. */
	caidaDesde = $state<number | null>(null);

	#url: string;
	#op: OpcionesConexion;
	#ws: WebSocket | null = null;
	#intento = 0;
	#reintento: ReturnType<typeof setTimeout> | undefined;
	#silencio: ReturnType<typeof setTimeout> | undefined;
	#vivo = false;
	#conecto = false;
	#perdio = false;
	#ocultaDesde: number | null = null;
	#limpiar: (() => void)[] = [];

	constructor(url: string, op: OpcionesConexion) {
		this.#url = url;
		this.#op = op;
	}

	iniciar() {
		if (this.#vivo || typeof WebSocket === 'undefined') return;
		this.#vivo = true;
		const enLinea = () => this.#yaMismo();
		const visibilidad = () => {
			if (document.hidden) {
				this.#ocultaDesde = Date.now();
				return;
			}
			const estuvo = this.#ocultaDesde ? Date.now() - this.#ocultaDesde : 0;
			this.#ocultaDesde = null;
			if (this.estado !== 'conectada') this.#yaMismo();
			else if (estuvo > OCULTA_MUCHO) this.#op.alResincronizar();
		};
		addEventListener('online', enLinea);
		document.addEventListener('visibilitychange', visibilidad);
		this.#limpiar.push(
			() => removeEventListener('online', enLinea),
			() => document.removeEventListener('visibilitychange', visibilidad)
		);
		this.#abrir();
	}

	detener() {
		this.#vivo = false;
		clearTimeout(this.#reintento);
		clearTimeout(this.#silencio);
		for (const f of this.#limpiar.splice(0)) f();
		const ws = this.#ws;
		this.#ws = null;
		ws?.close(1000);
		this.estado = 'parada';
		this.caidaDesde = null;
	}

	/** Reconectar ya (volvió la red o la pestaña), sin esperar el turno del reintento. */
	#yaMismo() {
		if (!this.#vivo || this.#ws || this.estado === 'parada') return;
		clearTimeout(this.#reintento);
		this.#abrir();
	}

	#vigilar(ws: WebSocket) {
		clearTimeout(this.#silencio);
		this.#silencio = setTimeout(() => {
			if (this.#ws === ws) ws.close(4999, 'silencio');
		}, SILENCIO_MAXIMO);
	}

	#abrir() {
		if (!this.#vivo) return;
		let ws: WebSocket;
		try {
			ws = new WebSocket(this.#url);
		} catch {
			return this.#programar('reintentar');
		}
		this.#ws = ws;
		this.#vigilar(ws);
		ws.onmessage = (m) => {
			if (this.#ws !== ws) return;
			this.#vigilar(ws);
			const e = leerEvento(m.data);
			if (!e || e.tipo === 'latido') return;
			if (e.tipo === 'hola') {
				const volvio = this.#conecto && this.#perdio;
				this.#conecto = true;
				this.#perdio = false;
				this.#intento = 0;
				this.estado = 'conectada';
				this.caidaDesde = null;
				// Lo que pasó mientras no había conexión no llegó: se recarga todo una vez.
				if (volvio) this.#op.alResincronizar();
				return;
			}
			if (e.tipo === 'resincronizar') return this.#op.alResincronizar();
			this.#op.alEvento(e);
		};
		ws.onclose = (c) => {
			if (this.#ws !== ws) return;
			this.#ws = null;
			clearTimeout(this.#silencio);
			if (this.#conecto) this.#perdio = true;
			if (!this.#vivo) return;
			const que = trasCierre(c.code);
			if (que === 'parar') {
				this.estado = 'parada';
				this.caidaDesde = null;
				if (c.code === 4401) this.#op.alSinSesion?.();
				return;
			}
			this.#programar(que);
		};
	}

	#programar(que: 'ya' | 'reintentar' | 'mas-tarde') {
		this.estado = 'reconectando';
		this.caidaDesde ??= Date.now();
		// Con la pestaña oculta no apura: al volver, reconecta en el acto (visibilitychange).
		const d =
			que === 'ya'
				? 0
				: que === 'mas-tarde'
					? 60_000 + espera(0)
					: espera(this.#intento++) * (document.hidden ? 4 : 1);
		clearTimeout(this.#reintento);
		this.#reintento = setTimeout(() => this.#abrir(), d);
	}
}
