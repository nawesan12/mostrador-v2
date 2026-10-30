/**
 * Instalar la app y activar los avisos, del lado del navegador. Chrome y Android avisan con
 * `beforeinstallprompt` que se puede instalar; iPhone no tiene ese botón (se agrega desde
 * Compartir) y sólo recibe avisos con la app ya instalada.
 */
type PedidoInstalar = Event & {
	prompt: () => Promise<void>;
	userChoice: Promise<{ outcome: string }>;
};

export const app = $state({
	instalable: null as PedidoInstalar | null,
	instalada: false,
	ios: false,
	permiso: 'default' as NotificationPermission | 'no-soportado',
	activando: false,
	error: ''
});

export function iniciar() {
	app.instalada =
		matchMedia('(display-mode: standalone)').matches ||
		(navigator as Navigator & { standalone?: boolean }).standalone === true;
	app.ios = /iPhone|iPad|iPod/.test(navigator.userAgent);
	app.permiso =
		'Notification' in window && 'PushManager' in window && 'serviceWorker' in navigator
			? Notification.permission
			: 'no-soportado';
	addEventListener('beforeinstallprompt', (e) => {
		e.preventDefault();
		app.instalable = e as PedidoInstalar;
	});
	addEventListener('appinstalled', () => {
		app.instalada = true;
		app.instalable = null;
	});
}

export async function instalar() {
	const p = app.instalable;
	if (!p) return;
	await p.prompt();
	if ((await p.userChoice).outcome === 'accepted') app.instalada = true;
	app.instalable = null;
}

function bytes(base64url: string): Uint8Array<ArrayBuffer> {
	const b = atob(
		base64url.replace(/-/g, '+').replace(/_/g, '/') + '='.repeat((4 - (base64url.length % 4)) % 4)
	);
	const out = new Uint8Array(new ArrayBuffer(b.length));
	for (let i = 0; i < b.length; i++) out[i] = b.charCodeAt(i);
	return out;
}

/** Pide el permiso y suscribe este navegador a los avisos de Stalwart. */
export async function activarAvisos(): Promise<boolean> {
	app.error = '';
	app.activando = true;
	try {
		const permiso = await Notification.requestPermission();
		app.permiso = permiso;
		if (permiso !== 'granted') return false;
		const r = await fetch('/avisos');
		if (!r.ok) throw new Error('El servidor de correo no manda avisos.');
		const { clave } = (await r.json()) as { clave: string };
		const reg = await navigator.serviceWorker.ready;
		const sus =
			(await reg.pushManager.getSubscription()) ??
			(await reg.pushManager.subscribe({
				userVisibleOnly: true,
				applicationServerKey: bytes(clave)
			}));
		const j = sus.toJSON() as { endpoint: string; keys: { p256dh: string; auth: string } };
		const g = await fetch('/avisos', {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({
				suscripcion: { endpoint: j.endpoint, keys: j.keys },
				dispositivo: `${app.ios ? 'iPhone' : /Android/.test(navigator.userAgent) ? 'Android' : 'Navegador'} ${new Date().toISOString().slice(0, 10)}`
			})
		});
		if (!g.ok) throw new Error('No se pudo registrar este dispositivo.');
		return true;
	} catch (e) {
		app.error = e instanceof Error ? e.message : 'No se pudieron activar los avisos.';
		return false;
	} finally {
		app.activando = false;
	}
}

const LISTO = 'correo-bienvenida';
export const yaVista = () => {
	try {
		return localStorage.getItem(LISTO) === '1';
	} catch {
		return false;
	}
};
export const marcarVista = () => {
	try {
		localStorage.setItem(LISTO, '1');
	} catch {
		// Sin almacenamiento (navegación privada): se vuelve a mostrar, no pasa nada.
	}
};
