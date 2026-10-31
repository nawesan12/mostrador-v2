/**
 * Servir la imagen al compartir: fuentes del build, fotos de la tienda (las subidas, del disco; las
 * demás, de internet) y la caché. La URL lleva la versión de lo dibujado (lib/tienda/compartir.ts):
 * se cachea un año y, al editar la tienda, la página apunta a otra.
 */
import { env } from 'cloudflare:workers';
import { read } from '$app/server';
import { CLAVE } from '#lib/server/archivos.ts';
import { versionDe, type Tarjeta } from '#lib/tienda/compartir.ts';
import { enCache } from './cache.ts';
import { bajar, renderizar } from './imagen.ts';

const ARCHIVOS = import.meta.glob('./fuentes/*.woff', {
	query: '?url',
	import: 'default',
	eager: true
}) as Record<string, string>;

const cargadas = new Map<string, Promise<ArrayBuffer>>();
function fuente(archivo: string) {
	let p = cargadas.get(archivo);
	if (!p) {
		const url = ARCHIVOS[`./fuentes/${archivo}.woff`];
		if (!url) throw new Error(`falta la fuente ${archivo}`);
		p = read(url).arrayBuffer();
		p.catch(() => cargadas.delete(archivo));
		cargadas.set(archivo, p);
	}
	return p;
}

/** Las fotos subidas (/f/fotos/…) se leen de los archivos; las demás se bajan. */
async function bytes(src: string): Promise<Uint8Array | null> {
	const clave = src.match(/^(?:https?:\/\/[^/]+)?\/f\/(fotos\/[^?#]+)$/)?.[1];
	if (clave) {
		if (!CLAVE.test(clave)) return null;
		const o = await env.FILES.get(clave).catch(() => null);
		return o && 'arrayBuffer' in o ? new Uint8Array(await o.arrayBuffer()) : null;
	}
	return bajar(src);
}

/** workerd se presenta así; Node (el VPS) dice «Node.js/…». */
export const enWorkers = () =>
	typeof navigator !== 'undefined' && navigator.userAgent === 'Cloudflare-Workers';

const sinImagen = () =>
	new Response('No pudimos armar la imagen. Probá en un rato.', {
		status: 503,
		headers: { 'cache-control': 'no-store', 'retry-after': '60' }
	});

/** La imagen como respuesta HTTP, con el cache-control que corresponde a su URL. */
export async function responderImagen(tarjeta: Tarjeta, prefijo: string, cacheControl: string) {
	// En el motor de prueba (wrangler dev, los e2e) no se puede dibujar: no hay vips ni procesos, y
	// el yoga de satori busca `self.location`, que workerd no tiene. Sólo se dibuja en el VPS (Node).
	if (enWorkers()) return sinImagen();
	const clave = `${prefijo}-${versionDe(tarjeta)}`.toLowerCase().replace(/[^a-z0-9-]/g, '-');
	let jpeg: Uint8Array;
	try {
		jpeg = await enCache(clave.slice(0, 120), () =>
			renderizar(tarjeta, { fuentes: fuente, bytes })
		);
	} catch (e) {
		// Sin vips (o con un error al dibujar): que WhatsApp vuelva a probar más tarde.
		console.error('compartir: no se pudo dibujar', e);
		return sinImagen();
	}
	return new Response(jpeg as BodyInit, {
		headers: {
			'content-type': 'image/jpeg',
			'content-length': String(jpeg.byteLength),
			'cache-control': cacheControl,
			'x-content-type-options': 'nosniff'
		}
	});
}
