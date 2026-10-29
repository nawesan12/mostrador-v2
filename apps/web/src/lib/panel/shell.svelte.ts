/**
 * Estado del armazón del panel: qué está abierto (barra lateral, notificaciones, ⌘K) y las
 * preferencias de la persona (tema y modo), que se guardan en cookies para el primer render.
 */
import { getContext, setContext } from 'svelte';

export type Tema = 'light' | 'dark';
export type Modo = 'simple' | 'completo';

const guardarCookie = (k: string, v: string) => {
	document.cookie = `${k}=${v}; path=/panel; max-age=${60 * 60 * 24 * 365}; samesite=lax`;
};

export type LinkPrevio = {
	monto?: number;
	concepto?: string;
	/** Nombre de quien paga (para el saludo). */
	para?: string;
	telefono?: string | null;
	pedido?: number | null;
};

export class Shell {
	barra = $state(false);
	notificaciones = $state(false);
	cmdk = $state(false);
	tema = $state<Tema>('light');
	modo = $state<Modo>('completo');
	/** La tienda toma turnos (o es de servicios): se ve la sección Turnos. */
	turnos = $state(false);

	// El modo Simple ya no se ofrece (confundía): quien lo tenía guardado vuelve al completo.
	constructor(tema: Tema) {
		this.tema = tema;
	}

	cambiarTema() {
		this.tema = this.tema === 'light' ? 'dark' : 'light';
		guardarCookie('panel_tema', this.tema);
	}

	cambiarModo(m: Modo) {
		this.modo = m;
		guardarCookie('panel_modo', m);
	}

	/** Panel de Tincho abierto. */
	ia = $state(false);
	/** Pregunta que llega de afuera (⌘K) para mandarle a Tincho al abrir. */
	preguntaPendiente = $state<string | null>(null);
	/** La charla con Tincho (queda mientras se navega por el panel). */
	charla = $state<{ rol: 'persona' | 'hornero'; texto: string }[]>([]);

	/** Abre Tincho (y, si viene, le manda la pregunta). */
	hornero(pregunta?: string) {
		this.barra = false;
		this.notificaciones = false;
		this.cmdk = false;
		this.preguntaPendiente = pregunta?.trim() || null;
		this.ia = true;
	}

	/** Modal "Crear link de pago" y lo que llega precargado (desde un pedido o una charla). */
	link = $state(false);
	linkPrevio = $state<LinkPrevio>({});
	abrirLink(previo: LinkPrevio = {}) {
		this.cmdk = false;
		this.linkPrevio = previo;
		this.link = true;
	}

	cerrarTodo() {
		this.barra = false;
		this.notificaciones = false;
		this.cmdk = false;
		this.ia = false;
		this.link = false;
	}
}

const KEY = Symbol('shell');
export const setShell = (s: Shell) => setContext(KEY, s);
export const getShell = () => getContext<Shell>(KEY);

/** "hace 4 min", "hace 2 h", "ayer"… (handoff). */
export function hace(iso: string, ahora = Date.now()) {
	const min = Math.max(0, Math.round((ahora - Date.parse(iso)) / 60_000));
	if (min < 1) return 'recién';
	if (min < 60) return `hace ${min} min`;
	const h = Math.round(min / 60);
	if (h < 24) return `hace ${h} h`;
	if (h < 48) return 'ayer';
	return `hace ${Math.round(h / 24)} días`;
}
