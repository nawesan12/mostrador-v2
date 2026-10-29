import { env } from 'cloudflare:workers';
import { leerEstado } from '#lib/server/estado/lectura.ts';
import { SITE } from '#lib/site.ts';
import type { RequestHandler } from './$types';

/** Incidentes de /estado como RSS ("También por RSS", handoff). */
const esc = (s: string) =>
	s.replace(
		/[&<>"']/g,
		(c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&apos;' })[c]!
	);

export const GET: RequestHandler = async () => {
	const { incidentes } = await leerEstado(env.ESTADO, 90);
	const items = incidentes
		.map(
			(i) =>
				`<item><title>${esc(i.titulo)}${i.fin ? ' (resuelto)' : ''}</title><link>${SITE}/estado</link><guid isPermaLink="false">incidente-${i.id}</guid><pubDate>${new Date(i.inicio).toUTCString()}</pubDate><description>${esc(i.detalle)}</description></item>`
		)
		.join('');
	const xml = `<?xml version="1.0" encoding="UTF-8"?><rss version="2.0"><channel><title>Estado de Tu Local Web</title><link>${SITE}/estado</link><description>Incidentes y mantenimientos de Tu Local Web.</description><language>es-AR</language>${items}</channel></rss>`;
	return new Response(xml, {
		headers: {
			'content-type': 'application/rss+xml; charset=utf-8',
			'cache-control': 'public, max-age=300'
		}
	});
};
