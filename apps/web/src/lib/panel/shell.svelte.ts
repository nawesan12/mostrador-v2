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

export class Shell {
	barra = $state(false);
	notificaciones = $state(false);
	cmdk = $state(false);
	tema = $state<Tema>('light');
	modo = $state<Modo>('completo');

	constructor(tema: Tema, modo: Modo) {
		this.tema = tema;
		this.modo = modo;
	}

	cambiarTema() {
		this.tema = this.tema === 'light' ? 'dark' : 'light';
		guardarCookie('panel_tema', this.tema);
	}

	cambiarModo(m: Modo) {
		this.modo = m;
		guardarCookie('panel_modo', m);
	}

	/** Hornero (el panel de IA) llega en su propio PR; mientras, se avisa. */
	hornero: () => void = () => {};

	cerrarTodo() {
		this.barra = false;
		this.notificaciones = false;
		this.cmdk = false;
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
