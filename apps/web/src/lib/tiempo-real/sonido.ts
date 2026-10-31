/**
 * El sonido de «pedido nuevo» (Notificaciones › En este dispositivo). Apagado por defecto: lo prende
 * cada persona en cada navegador (es del aparato, no de la tienda), así que vive en localStorage.
 * Un timbre corto de dos notas armado con WebAudio: nada que descargar.
 */
const CLAVE = 'tlw-sonido-pedidos';

export function sonidoPrendido(): boolean {
	try {
		return localStorage.getItem(CLAVE) === '1';
	} catch {
		return false;
	}
}

export function prenderSonido(si: boolean) {
	try {
		if (si) localStorage.setItem(CLAVE, '1');
		else localStorage.removeItem(CLAVE);
	} catch {
		// Sin almacenamiento (modo privado): queda apagado.
	}
}

let contexto: AudioContext | null = null;

/** Suena el timbre. Si el navegador todavía no dejó reproducir (sin un toque antes), no pasa nada. */
export function sonarPedido() {
	try {
		contexto ??= new AudioContext();
		const ctx = contexto;
		if (ctx.state === 'suspended') void ctx.resume();
		const t = ctx.currentTime;
		for (const [i, f] of [880, 1318.5].entries()) {
			const o = ctx.createOscillator();
			const g = ctx.createGain();
			o.type = 'sine';
			o.frequency.value = f;
			const t0 = t + i * 0.16;
			g.gain.setValueAtTime(0, t0);
			g.gain.linearRampToValueAtTime(0.18, t0 + 0.02);
			g.gain.exponentialRampToValueAtTime(0.0001, t0 + 0.45);
			o.connect(g).connect(ctx.destination);
			o.start(t0);
			o.stop(t0 + 0.5);
		}
	} catch {
		// Sin WebAudio: el aviso se ve igual.
	}
}
