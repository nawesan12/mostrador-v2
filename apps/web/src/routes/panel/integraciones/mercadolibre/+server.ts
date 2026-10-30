import { error, json } from '@sveltejs/kit';
import * as v from 'valibot';
import { datosDe } from '#lib/server/datos/index.ts';
import { guardarVuelta } from '#lib/server/volver.ts';
import type { RequestHandler } from './$types';

/** Conectar: adónde mandar a la persona (Mercado Libre, o la vuelta directa en la demo). */
export const POST: RequestHandler = async ({ locals, url: pedido, cookies }) => {
	// ?volver=/onboarding?paso=3: al terminar en el proveedor se vuelve ahí, no a Integraciones.
	guardarVuelta(cookies, pedido.searchParams.get('volver'), pedido.protocol === 'https:');
	const url = await datosDe(locals).conectarMercadoLibre();
	if (!url) error(409, 'La conexión con Mercado Libre todavía no está disponible.');
	if (
		!url.startsWith('https://auth.mercadolibre.com.ar/') &&
		!url.startsWith('/integraciones/mercadolibre/')
	)
		error(502, 'Respuesta inesperada.');
	return json({ url });
};

const Ajustes = v.object({
	recargo: v.pipe(v.number(), v.integer(), v.minValue(0), v.maxValue(60)),
	pausarSinStock: v.boolean()
});

/** Recargo por comisión y pausar sin stock. */
export const PUT: RequestHandler = async ({ locals, request }) => {
	const b = v.safeParse(Ajustes, await request.json().catch(() => null));
	if (!b.success) error(400, 'Pedido inválido.');
	const r = await datosDe(locals).ajustarMercadoLibre(b.output.recargo, b.output.pausarSinStock);
	if (r !== 'ok') error(409, 'Conectá Mercado Libre primero.');
	return new Response(null, { status: 204 });
};

export const DELETE: RequestHandler = async ({ locals }) => {
	await datosDe(locals).desconectarMercadoLibre();
	return new Response(null, { status: 204 });
};
