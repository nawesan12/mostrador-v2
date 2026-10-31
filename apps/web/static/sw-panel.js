/*
 * Service worker del panel instalado como app (docs/decisiones.md §42). Alcance: /panel.
 *
 * - No guarda nada del panel: los pedidos, las ventas y los clientes son privados y van siempre a la
 *   red. Sólo se guardan la pantalla «Sin conexión» y sus dos imágenes.
 * - Sin red, una navegación del panel muestra «Sin conexión» con «Reintentar».
 * - Push web: muestra el aviso que manda el worker (title, body, url, tienda, tag) y al tocarlo abre
 *   esa página del panel, en la tienda del aviso.
 */
const CACHE = 'tlw-panel-v1';
const SIN_CONEXION = '/panel/sin-conexion.html';
const GUARDADO = [SIN_CONEXION, '/hornero/desenchufado.webp', '/mark.svg'];

self.addEventListener('install', (event) => {
	event.waitUntil(
		caches
			.open(CACHE)
			.then((c) => c.addAll(GUARDADO))
			.then(() => self.skipWaiting())
	);
});

self.addEventListener('activate', (event) => {
	event.waitUntil(
		caches
			.keys()
			.then((ks) => Promise.all(ks.filter((k) => k !== CACHE).map((k) => caches.delete(k))))
			.then(() => self.clients.claim())
	);
});

self.addEventListener('fetch', (event) => {
	const req = event.request;
	if (req.mode !== 'navigate' || req.method !== 'GET') return;
	event.respondWith(
		fetch(req).catch(async () => (await caches.match(SIN_CONEXION)) ?? Response.error())
	);
});

self.addEventListener('push', (event) => {
	let d;
	try {
		d = event.data ? event.data.json() : {};
	} catch {
		d = {};
	}
	const titulo = typeof d.title === 'string' && d.title ? d.title : 'Tu Local Web';
	event.waitUntil(
		self.registration.showNotification(titulo, {
			body: typeof d.body === 'string' ? d.body : '',
			icon: '/panel/icono-192.png',
			badge: '/panel/insignia-96.png',
			tag: typeof d.tag === 'string' && d.tag ? d.tag : undefined,
			lang: 'es-AR',
			data: { url: typeof d.url === 'string' ? d.url : '/panel', tienda: d.tienda ?? '' }
		})
	);
});

/** Sólo páginas del panel, del mismo origen. */
function destino(url) {
	try {
		const u = new URL(url, self.location.origin);
		if (u.origin === self.location.origin && /^\/panel(\/|$|\?)/.test(u.pathname + u.search))
			return u.href;
	} catch {
		// URL rota: al Inicio.
	}
	return new URL('/panel', self.location.origin).href;
}

self.addEventListener('notificationclick', (event) => {
	event.notification.close();
	const { url, tienda } = event.notification.data ?? {};
	const hacia = destino(url);
	event.waitUntil(
		(async () => {
			// El aviso puede ser de otra de sus tiendas: se elige esa antes de abrir la página.
			if (typeof tienda === 'string' && tienda)
				await fetch('/panel/tienda', {
					method: 'PUT',
					headers: { 'content-type': 'application/json' },
					body: JSON.stringify({ tiendaId: tienda }),
					credentials: 'same-origin'
				}).catch(() => null);
			const abiertas = await self.clients.matchAll({ type: 'window', includeUncontrolled: true });
			const panel = abiertas.find((c) => new URL(c.url).pathname.startsWith('/panel'));
			if (panel) {
				await panel.focus();
				return panel.navigate(hacia).catch(() => self.clients.openWindow(hacia));
			}
			return self.clients.openWindow(hacia);
		})()
	);
});
