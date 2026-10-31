/**
 * El panel instalado como app en el celular (docs/decisiones.md §42): el service worker, la
 * invitación a instalar y el push web. Sólo navegador.
 */

/** Lo que Chrome en Android guarda para instalar con un botón propio. */
interface EventoInstalar extends Event {
	prompt(): Promise<void>;
	userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>;
}

declare global {
	interface Window {
		/** app.html lo guarda apenas llega, antes de que el panel arranque. */
		__tlwInstalar?: EventoInstalar | null;
	}
}

export const SW = '/sw-panel.js';
export const ALCANCE = '/panel';
const CERRADA = 'tlw-instalar-cerrada';

/** El estado de la instalación, compartido por la tarjeta del Inicio y Notificaciones. */
class Instalacion {
	/** El pedido de Chrome para instalar (null: no hay, o ya se usó). */
	evento = $state<EventoInstalar | null>(null);
	instalada = $state(false);
	cerrada = $state(false);
	listo = $state(false);
}
export const instalacion = new Instalacion();

/** Abierto desde el ícono de la pantalla de inicio (no desde el navegador). */
export function esInstalada() {
	if (typeof window === 'undefined') return false;
	return (
		window.matchMedia?.('(display-mode: standalone)').matches ||
		window.matchMedia?.('(display-mode: fullscreen)').matches ||
		(navigator as Navigator & { standalone?: boolean }).standalone === true
	);
}

/** iPhone o iPad (el iPad moderno dice ser una Mac con pantalla táctil). */
export function esIOS() {
	if (typeof navigator === 'undefined') return false;
	return (
		/iPhone|iPad|iPod/.test(navigator.userAgent) ||
		(navigator.userAgent.includes('Macintosh') && navigator.maxTouchPoints > 1)
	);
}

export function esAndroid() {
	return typeof navigator !== 'undefined' && /Android/.test(navigator.userAgent);
}

/** La versión de iOS (16.4 → 16.4), o null si no es un iPhone o no se sabe. */
export function versionIOS(): number | null {
	if (!esIOS()) return null;
	const m = navigator.userAgent.match(/OS (\d+)_(\d+)/);
	return m ? Number(m[1]) + Number(m[2]) / 10 : null;
}

function leerCerrada() {
	try {
		return localStorage.getItem(CERRADA) === '1';
	} catch {
		return false;
	}
}

/** «No, gracias»: la tarjeta no vuelve en este navegador. */
export function cerrarInvitacion() {
	instalacion.cerrada = true;
	try {
		localStorage.setItem(CERRADA, '1');
	} catch {
		// Sin almacenamiento (modo privado): se cierra por esta visita.
	}
}

/** Android: abre el diálogo del sistema para instalar. */
export async function instalar() {
	const e = instalacion.evento;
	if (!e) return false;
	instalacion.evento = null;
	window.__tlwInstalar = null;
	await e.prompt();
	const { outcome } = await e.userChoice;
	if (outcome === 'accepted') instalacion.instalada = true;
	return outcome === 'accepted';
}

/** Lo llama el armazón del panel al montar: registra el service worker y escucha la instalación. */
export function arrancarAppCelular() {
	instalacion.instalada = esInstalada();
	instalacion.cerrada = leerCerrada();
	instalacion.evento = window.__tlwInstalar ?? null;
	instalacion.listo = true;
	const alPedir = (e: Event) => {
		e.preventDefault();
		instalacion.evento = e as EventoInstalar;
	};
	const alInstalar = () => {
		instalacion.instalada = true;
		instalacion.evento = null;
	};
	addEventListener('beforeinstallprompt', alPedir);
	addEventListener('appinstalled', alInstalar);
	navigator.serviceWorker?.register(SW, { scope: ALCANCE }).catch(() => {
		// Sin service worker (navegador viejo, modo privado de Firefox): el panel anda igual.
	});
	return () => {
		removeEventListener('beforeinstallprompt', alPedir);
		removeEventListener('appinstalled', alInstalar);
	};
}

// ---------- Push web ----------

export type EstadoPush =
	| 'cargando'
	/** Sin claves VAPID en el servidor, la demo o el soporte mirando. */
	| 'no-disponible'
	/** iPhone sin instalar: en iOS el push web sólo anda desde la pantalla de inicio (16.4+). */
	| 'instalar-primero'
	/** iPhone con iOS anterior a 16.4. */
	| 'ios-viejo'
	/** El navegador no tiene Web Push. */
	| 'sin-soporte'
	| 'bloqueado'
	| 'apagado'
	| 'prendido';

export function soportaPush() {
	return (
		typeof window !== 'undefined' &&
		'serviceWorker' in navigator &&
		'PushManager' in window &&
		'Notification' in window
	);
}

/** La clave VAPID de base64url a bytes (applicationServerKey). */
export function claveABytes(b64: string) {
	const s = (b64 + '='.repeat((4 - (b64.length % 4)) % 4)).replace(/-/g, '+').replace(/_/g, '/');
	const bin = atob(s);
	const out = new Uint8Array(bin.length);
	for (let i = 0; i < bin.length; i++) out[i] = bin.charCodeAt(i);
	return out;
}

/** Cómo se ve este aparato en la lista («iPhone · Safari», «Android · Chrome»). */
export function aparato(ua: string) {
	const so = /iPhone/.test(ua)
		? 'iPhone'
		: /iPad/.test(ua) || (/Macintosh/.test(ua) && esIOS())
			? 'iPad'
			: /Android/.test(ua)
				? 'Android'
				: /Windows/.test(ua)
					? 'Windows'
					: /Macintosh/.test(ua)
						? 'Mac'
						: 'Otro';
	const nav = /EdgA?\//.test(ua)
		? 'Edge'
		: /SamsungBrowser/.test(ua)
			? 'Samsung Internet'
			: /Firefox|FxiOS/.test(ua)
				? 'Firefox'
				: /Chrome|CriOS/.test(ua)
					? 'Chrome'
					: /Safari/.test(ua)
						? 'Safari'
						: '';
	return nav ? `${so} · ${nav}` : so;
}

async function registro() {
	return (
		(await navigator.serviceWorker.getRegistration(ALCANCE)) ??
		(await navigator.serviceWorker.register(SW, { scope: ALCANCE }))
	);
}

/** El estado del push en este celular, con lo que dijo el servidor. */
export async function estadoPush(srv: { clave: string; suscripta: boolean }): Promise<EstadoPush> {
	if (!srv.clave) return 'no-disponible';
	if (esIOS() && !esInstalada()) {
		const v = versionIOS();
		return v !== null && v < 16.4 ? 'ios-viejo' : 'instalar-primero';
	}
	if (!soportaPush()) return 'sin-soporte';
	if (Notification.permission === 'denied') return 'bloqueado';
	if (Notification.permission !== 'granted') return 'apagado';
	const reg = await navigator.serviceWorker.getRegistration(ALCANCE);
	const sub = await reg?.pushManager.getSubscription();
	return sub && srv.suscripta ? 'prendido' : 'apagado';
}

/**
 * Activa el push en este celular. Pide el permiso primero (en el iPhone tiene que ser apenas se toca
 * el botón), se suscribe con la clave VAPID y manda la suscripción al servidor.
 */
export async function activarPush(clave: string): Promise<'prendido' | 'bloqueado' | 'error'> {
	const permiso = await Notification.requestPermission();
	if (permiso === 'denied') return 'bloqueado';
	if (permiso !== 'granted') return 'error';
	try {
		const reg = await registro();
		await navigator.serviceWorker.ready;
		const opciones = { userVisibleOnly: true, applicationServerKey: claveABytes(clave) };
		let sub: PushSubscription;
		try {
			sub = await reg.pushManager.subscribe(opciones);
		} catch {
			// Una suscripción vieja con otra clave (se cambiaron las VAPID): se da de baja y otra vez.
			await (await reg.pushManager.getSubscription())?.unsubscribe();
			sub = await reg.pushManager.subscribe(opciones);
		}
		const j = sub.toJSON();
		const res = await fetch('/panel/notificaciones/push', {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({
				endpoint: j.endpoint,
				p256dh: j.keys?.p256dh ?? '',
				auth: j.keys?.auth ?? '',
				aparato: aparato(navigator.userAgent)
			})
		});
		return res.ok ? 'prendido' : 'error';
	} catch {
		return 'error';
	}
}

/** Desactiva el push en este celular (y lo borra del servidor). */
export async function desactivarPush() {
	const reg = await navigator.serviceWorker.getRegistration(ALCANCE);
	await (await reg?.pushManager.getSubscription())?.unsubscribe();
	const res = await fetch('/panel/notificaciones/push', { method: 'DELETE' });
	return res.ok;
}

export async function probarPush() {
	const res = await fetch('/panel/notificaciones/push/prueba', { method: 'POST' });
	return res.ok;
}
