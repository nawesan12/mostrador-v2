import { error } from '@sveltejs/kit';
import { Jmap } from '#lib/server/jmap.ts';
import type { RequestHandler } from './$types';

/** Un adjunto (o imagen cid:) del correo, pasado por el servidor con la sesión de la persona. */
export const GET: RequestHandler = async ({ locals, params, url }) => {
	const nombre = (url.searchParams.get('n') ?? 'adjunto').slice(0, 200);
	const tipo = (url.searchParams.get('t') ?? 'application/octet-stream').slice(0, 120);
	const r = await new Jmap(locals.cred!).descargar(params.cuenta, params.blob, nombre, tipo);
	if (!r.ok || !r.body) error(r.status === 404 ? 404 : 502, 'No encontramos ese adjunto.');
	// Lo que el navegador podría ejecutar (HTML, SVG) sólo se descarga, nunca se abre acá.
	const seguro = /^(image\/(png|jpeg|gif|webp|avif)|application\/pdf|text\/plain)$/.test(tipo);
	const ver = url.searchParams.get('ver') === '1' && seguro;
	return new Response(r.body, {
		headers: {
			'content-type': seguro ? tipo : 'application/octet-stream',
			'content-disposition': `${ver || tipo.startsWith('image/') ? 'inline' : 'attachment'}; filename*=UTF-8''${encodeURIComponent(nombre)}`,
			'cache-control': 'private, max-age=3600',
			// El PDF se lee en el visor (un iframe del mismo sitio): Chrome no abre PDF con sandbox.
			'content-security-policy':
				ver && tipo === 'application/pdf'
					? "default-src 'none'; img-src 'self' data:; style-src 'unsafe-inline'; frame-ancestors 'self'"
					: "default-src 'none'; img-src 'self' data:; style-src 'unsafe-inline'; sandbox"
		}
	});
};
