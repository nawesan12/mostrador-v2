import { getContext, setContext } from 'svelte';
import { FEED } from './datos.ts';

/**
 * Estado vivo de la landing, como el componente del handoff (design/Mostrador v4.dc.html):
 * el feed de ventas avanza cada 3,2 s (motion "media"), el chat de Hornero cada 1,7 s y el widget
 * aparece a los 2,6 s. Se crea en el navegador (onMount); en el servidor queda en su estado inicial.
 */
export class Landing {
	tick = $state(0);
	heroTotal = $state(412800);
	mig = $state(64);
	chat = $state(0);
	widget = $state(false);
	widgetOpen = $state(false);
	menu = $state(false);
	/** Plantilla abierta en el modal (índice en PLANTILLAS) o null. */
	tplI = $state<number | null>(null);

	#timers: (ReturnType<typeof setInterval> | ReturnType<typeof setTimeout>)[] = [];

	iniciar(chatLargo: number) {
		this.#timers.push(
			setInterval(() => {
				this.heroTotal += FEED[(this.tick + 3) % FEED.length][3];
				this.tick += 1;
				this.mig = this.mig >= 100 ? 64 : Math.min(100, this.mig + 6);
			}, 3200),
			setInterval(() => (this.chat = this.chat >= chatLargo + 2 ? 0 : this.chat + 1), 1700),
			// Hornero aparece cerrado: sólo el avatar, sin globo ni aviso hasta que lo tocan.
			setTimeout(() => (this.widget = true), 2600)
		);
	}

	detener() {
		for (const t of this.#timers) clearInterval(t);
		this.#timers = [];
	}
}

const KEY = Symbol('landing');
export const setLanding = (l: Landing) => setContext(KEY, l) as Landing;
export const getLanding = () => getContext<Landing>(KEY);
