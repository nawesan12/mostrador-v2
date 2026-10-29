import { env } from 'cloudflare:workers';
import { error, json } from '@sveltejs/kit';
import { LIMITES, subirFoto } from '#lib/server/archivos.ts';
import { datosDe, usaSandbox } from '#lib/server/datos/index.ts';
import { golpe } from '#lib/server/limite.ts';
import { dbSandbox } from '#lib/server/sandbox/sandbox.ts';
import type { RequestHandler } from './$types';

/**
 * Subir una foto de producto: el cuerpo es la imagen (ya comprimida a WebP en el navegador).
 * Devuelve su ruta en /f/ (con el origen, una URL https que sirve también para Mercado Libre).
 */
export const POST: RequestHandler = async (event) => {
	datosDe(event.locals); // sin sesión, a ingresar
	const largo = Number(event.request.headers.get('content-length') ?? 0);
	if (largo > LIMITES.archivo) error(413, 'La foto pesa más de 1 MB.');
	const bytes = new Uint8Array(await event.request.arrayBuffer());
	const db = dbSandbox(env.SANDBOX);
	// La tienda es la dueña (su id va en la URL). El id del sandbox es la cookie del visitante: no se
	// publica, sólo se guarda para borrar las fotos cuando la demo vence.
	const tienda = (await datosDe(event.locals).tienda()).id;
	if (!(await golpe(db, `fotos:${tienda}`, 60, 3600))) error(429, 'Subiste muchas fotos seguidas.');
	const sandbox = usaSandbox() ? (event.locals.sandbox?.id ?? null) : null;
	const clave = await subirFoto(db, tienda, bytes, sandbox);
	// La ruta: el navegador le pone su origen (https://tulocalweb.com.ar o, en las pruebas,
	// localhost; en wrangler dev el origen del pedido es el de producción).
	return json({ ruta: `/f/${clave}` }, { status: 201 });
};
