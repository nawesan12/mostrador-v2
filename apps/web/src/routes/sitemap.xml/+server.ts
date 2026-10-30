import { env } from 'cloudflare:workers';
import { ARTICULOS } from '#lib/ayuda/articulos.ts';
import { usaSandbox } from '#lib/server/datos/index.ts';
import { SITE } from '#lib/site.ts';
import type { RequestHandler } from './$types';

type Entrada = {
	tienda: string;
	dominio: string | null;
	producto: string | null;
	actualizado: string;
};

/** Las páginas públicas de la web, que no cambian seguido. */
const FIJAS = [
	'/',
	'/migrar',
	'/nosotros',
	'/seguridad',
	'/contacto',
	'/partners',
	'/estado',
	'/ayuda',
	'/legales/terminos',
	'/legales/privacidad',
	'/legales/cookies'
];

const esc = (s: string) =>
	s.replace(
		/[&<>"']/g,
		(c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&apos;' })[c]!
	);

/**
 * El sitemap: la web, la ayuda y lo público de cada tienda (portada y productos publicados). Con
 * dominio propio verificado, la URL es la de su dominio (la canónica).
 */
export const GET: RequestHandler = async () => {
	let tiendas: Entrada[] = [];
	if (!usaSandbox()) {
		const r = await fetch(`${String(env.API_URL)}/v1/publico/sitemap`, {
			signal: AbortSignal.timeout(8000)
		}).catch(() => null);
		if (r?.ok) tiendas = (await r.json()) as Entrada[];
	}
	const urls: { loc: string; mod?: string }[] = [
		...FIJAS.map((p) => ({ loc: SITE + p })),
		...ARTICULOS.map((a) => ({ loc: `${SITE}/ayuda/${a.categoria}/${a.slug}` })),
		...tiendas.map((t) => {
			const base = t.dominio ? `https://${t.dominio}` : `${SITE}/t/${t.tienda}`;
			return {
				loc: t.producto ? `${base}/producto/${t.producto}` : t.dominio ? `${base}/` : base,
				mod: t.actualizado.slice(0, 10)
			};
		})
	];
	const xml =
		'<?xml version="1.0" encoding="UTF-8"?>\n<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">\n' +
		urls
			.map(
				(u) => `<url><loc>${esc(u.loc)}</loc>${u.mod ? `<lastmod>${u.mod}</lastmod>` : ''}</url>`
			)
			.join('\n') +
		'\n</urlset>\n';
	return new Response(xml, {
		headers: {
			'content-type': 'application/xml; charset=utf-8',
			'cache-control': 'public, max-age=3600'
		}
	});
};
