import { json } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import { MENSAJES, motivoDeFormato } from '#lib/tienda/direccion.ts';
import type { RequestHandler } from './$types';

/**
 * La dirección de la tienda (decisiones §38). GET ?slug= dice si está disponible (el campo lo pide
 * mientras se escribe); PUT {slug} la cambia. Lo usan el onboarding y Configuración › Negocio.
 */
export const GET: RequestHandler = async ({ locals, url, setHeaders }) => {
	setHeaders({ 'cache-control': 'no-store' });
	const slug = (url.searchParams.get('slug') ?? '').slice(0, 100);
	// El formato no necesita a nadie: se responde acá.
	const m = motivoDeFormato(slug);
	if (m === 'invalido') return json({ slug, disponible: false, motivo: m, mensaje: MENSAJES[m] });
	return json(await datosDe(locals).disponibilidadDireccion(slug));
};

export const PUT: RequestHandler = async ({ locals, request }) => {
	const cuerpo = (await request.json().catch(() => null)) as { slug?: unknown } | null;
	const slug = typeof cuerpo?.slug === 'string' ? cuerpo.slug : '';
	const r = await datosDe(locals).cambiarDireccion(slug);
	return json(r, { status: r.ok ? 200 : r.motivo === 'invalido' ? 400 : 409 });
};
