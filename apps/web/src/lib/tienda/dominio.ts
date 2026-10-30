/**
 * Dominios propios de las tiendas (mitienda.com.ar). Todo host que no es de Tu Local Web es el de
 * una tienda: el hook `reroute` (src/hooks.ts) lleva sus paths a /t/<slug> y el servidor no deja
 * ver nada más en ese dominio (ni el panel ni la landing).
 */
export function esDominioPropio(host: string): boolean {
	const h = host.toLowerCase();
	return !(
		h === 'localhost' ||
		h === '127.0.0.1' ||
		h === 'tulocalweb.com.ar' ||
		h.endsWith('.tulocalweb.com.ar') ||
		h.endsWith('.workers.dev')
	);
}

/** Paths que en un dominio propio se sirven tal cual, sin pasarlos a /t/<slug>. */
export const GLOBALES_EN_DOMINIO = /^\/(api\/dominio|legales)(\/|$)/;

/** El path interno (/t/<slug>/…) de un path pedido en el dominio de la tienda `slug`. */
export function pathInterno(slug: string, path: string): string | undefined {
	if (GLOBALES_EN_DOMINIO.test(path)) return;
	const base = `/t/${slug}`;
	if (path === base || path.startsWith(`${base}/`)) return;
	return path === '/' ? base : base + path;
}
