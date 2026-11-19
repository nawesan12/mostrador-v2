import { env } from 'cloudflare:workers';
import { error, json } from '@sveltejs/kit';
import { LIMITES, subirFoto } from '#lib/server/archivos.ts';
import { datosDe, usaSandbox } from '#lib/server/datos/index.ts';
import { golpe } from '#lib/server/limite.ts';
import { planDeTienda } from '#lib/panel/planes.ts';
import { dbSandbox } from '#lib/server/sandbox/sandbox.ts';
import type { RequestHandler } from './$types';

/**
 * Subir una foto de producto: el cuerpo es la imagen (achicada en el navegador si se pudo). El
 * servidor la deja en WebP de 1600 px (lib/server/archivos.ts).
 * Devuelve su ruta en /f/ (con el origen, una URL https que sirve también para Mercado Libre).
 */
export const POST: RequestHandler = async (event) => {
	datosDe(event.locals); // sin sesión, a ingresar
	const largo = Number(event.request.headers.get('content-length') ?? 0);
	if (largo > LIMITES.subida) error(413, 'La foto pesa más de 15 MB.');
	const bytes = new Uint8Array(await event.request.arrayBuffer());
	const db = dbSandbox(env.SANDBOX);
	// La tienda es la dueña (su id va en la URL). El id del sandbox es la cookie del visitante: no se
	// publica, sólo se guarda para borrar las fotos cuando la demo vence.
	const datos = datosDe(event.locals);
	const [t, acceso] = await Promise.all([datos.tienda(), datos.acceso()]);
	const tienda = t.id;
	// Las fotos son de los productos y de Mi tienda: sólo quien puede editarlos, y no quien quedó en
	// sólo lectura por el plan (la misma regla que aplica la API a lo demás).
	const permisos = acceso.permisos as readonly string[];
	if (!['productos', 'tienda'].some((p) => permisos.includes(p)))
		error(403, 'Tu rol no puede subir fotos.');
	if (acceso.soloLectura)
		error(
			402,
			'Tu lugar en el equipo quedó en sólo lectura por el plan. La dueña puede pasar de plan o liberar un lugar.'
		);
	if (!(await golpe(db, `fotos:${tienda}`, 60, 3600))) error(429, 'Subiste muchas fotos seguidas.');
	const sandbox = usaSandbox() ? (event.locals.sandbox?.id ?? null) : null;
	// El espacio para fotos del plan que rige (en la prueba, el de Crece, con la fecha de la API). Un
	// plan pago tiene reservado lo último del espacio total (archivos.ts › LIMITES.reservaPagos).
	const plan = planDeTienda(t);
	const clave = await subirFoto(
		db,
		tienda,
		bytes,
		sandbox,
		plan.fotosMB * 1024 * 1024,
		t.plan !== 'gratis'
	);
	// La ruta: el navegador le pone su origen (https://tulocalweb.com.ar o, en las pruebas,
	// localhost; en wrangler dev el origen del pedido es el de producción).
	return json({ ruta: `/f/${clave}` }, { status: 201 });
};
