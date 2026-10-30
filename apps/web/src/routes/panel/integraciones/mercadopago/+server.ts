import { error, json } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import { guardarVuelta } from '#lib/server/volver.ts';
import type { RequestHandler } from './$types';

/** Conectar: adónde mandar a la persona (Mercado Pago, o la vuelta directa en la demo). */
export const POST: RequestHandler = async ({ locals, url: pedido, cookies }) => {
	// ?volver=/onboarding?paso=3: al terminar en el proveedor se vuelve ahí, no a Integraciones.
	guardarVuelta(cookies, pedido.searchParams.get('volver'), pedido.protocol === 'https:');
	const url = await datosDe(locals).conectarMercadoPago();
	if (!url) error(409, 'La conexión con Mercado Pago todavía no está disponible.');
	if (
		!url.startsWith('https://auth.mercadopago.com/') &&
		!url.startsWith('/integraciones/mercadopago/')
	)
		error(502, 'Respuesta inesperada.');
	return json({ url });
};

/** Desconectar: se borran los tokens de la tienda. */
export const DELETE: RequestHandler = async ({ locals }) => {
	await datosDe(locals).desconectarMercadoPago();
	return new Response(null, { status: 204 });
};
