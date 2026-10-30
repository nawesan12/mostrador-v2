/// <reference no-default-lib="true"/>
/// <reference lib="esnext" />
/// <reference lib="webworker" />
/**
 * El correo instalable: se guarda sólo la app (código, fuentes, íconos), nunca correos ni
 * adjuntos, que pasan siempre por la red con la sesión. Sin conexión, la bandeja muestra el aviso
 * de /sin-conexion.html en vez de una pantalla rota.
 */
import { version } from '$app/env';
import { assets, immutable } from '$app/manifest';
import { self as sw } from '$app/service-worker';

const CACHE = `correo-${version}`;
// Las rutas del manifiesto vienen sin la barra del principio. De las fuentes, sólo woff2 latinas.
const APP = [
	...immutable
		.map((a) => a.path)
		.filter((p) => !/\.woff$|-(cyrillic|greek|vietnamese|symbols\d?)-/.test(p)),
	...assets.map((a) => a.path).filter((p) => !p.endsWith('.webmanifest') && !p.endsWith('og.jpg'))
].map((p) => (p.startsWith('/') ? p : `/${p}`));

sw.addEventListener('install', (e) => {
	e.waitUntil(
		caches
			.open(CACHE)
			.then((c) => c.addAll(APP))
			.then(() => sw.skipWaiting())
	);
});

sw.addEventListener('activate', (e) => {
	e.waitUntil(
		caches
			.keys()
			.then((ks) => Promise.all(ks.filter((k) => k !== CACHE).map((k) => caches.delete(k))))
			.then(() => sw.clients.claim())
	);
});

sw.addEventListener('fetch', (e) => {
	const u = new URL(e.request.url);
	if (e.request.method !== 'GET' || u.origin !== sw.location.origin) return;
	if (APP.includes(u.pathname)) {
		e.respondWith(caches.match(u.pathname).then((r) => r ?? fetch(e.request)));
		return;
	}
	// Las páginas van siempre a la red; si no hay, el aviso de sin conexión.
	if (e.request.mode === 'navigate')
		e.respondWith(
			fetch(e.request).catch(
				async () => (await caches.match('/sin-conexion.html')) ?? Response.error()
			)
		);
});

// ---- Avisos (Web Push que manda Stalwart) ----

const ESTADO = 'correo-avisos';
/** Hasta cuándo ya se avisó: así cada push muestra sólo lo que entró después. */
async function ultimo(nuevo?: string): Promise<string | null> {
	const c = await caches.open(ESTADO);
	if (nuevo) {
		await c.put('/__ultimo', new Response(nuevo));
		return nuevo;
	}
	const r = await c.match('/__ultimo');
	return r ? r.text() : null;
}

type Aviso = { id: string; de: string; asunto: string; vista: string; casilla: string };

sw.addEventListener('push', (e) => {
	const datos = (() => {
		try {
			return e.data?.json() as Record<string, string> | undefined;
		} catch {
			return undefined;
		}
	})();
	e.waitUntil(
		(async () => {
			// El primer push de Stalwart trae el código para confirmar la suscripción.
			if (datos?.['@type'] === 'PushVerification') {
				await fetch('/avisos/verificar', {
					method: 'POST',
					headers: { 'content-type': 'application/json' },
					body: JSON.stringify({ id: datos.pushSubscriptionId, codigo: datos.verificationCode })
				});
				await ultimo(new Date().toISOString());
				return sw.registration.showNotification('Avisos activados', {
					body: 'Te vamos a avisar acá cuando llegue un correo.',
					icon: '/icono-192.png',
					badge: '/icono-96.png',
					tag: 'avisos'
				});
			}
			const desde = (await ultimo()) ?? new Date(Date.now() - 10 * 60_000).toISOString();
			let avisos: Aviso[] = [];
			try {
				const r = await fetch(`/avisos/nuevos?desde=${encodeURIComponent(desde)}`);
				if (r.ok) avisos = ((await r.json()) as { avisos: Aviso[] }).avisos;
			} catch {
				// Sin red o sin sesión: igual se avisa (todo push tiene que mostrar algo).
			}
			await ultimo(new Date().toISOString());
			if (!avisos.length)
				return sw.registration.showNotification('Llegó un correo nuevo', {
					icon: '/icono-192.png',
					badge: '/icono-96.png',
					tag: 'nuevo',
					data: { url: '/' }
				});
			await Promise.all(
				avisos.slice(0, 3).map((a) =>
					sw.registration.showNotification(a.de, {
						body: `${a.asunto}\n${a.vista}`,
						icon: '/icono-192.png',
						badge: '/icono-96.png',
						tag: a.id,
						data: { url: `/?m=${encodeURIComponent(a.id)}` }
					})
				)
			);
		})()
	);
});

sw.addEventListener('notificationclick', (e) => {
	e.notification.close();
	const url = (e.notification.data as { url?: string } | null)?.url ?? '/';
	e.waitUntil(
		(async () => {
			const ventanas = await sw.clients.matchAll({ type: 'window', includeUncontrolled: true });
			const abierta = ventanas.find((v) => new URL(v.url).origin === sw.location.origin);
			if (abierta) {
				await abierta.focus();
				return (abierta as WindowClient).navigate(url);
			}
			return sw.clients.openWindow(url);
		})()
	);
});
