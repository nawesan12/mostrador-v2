import { error, json } from '@sveltejs/kit';
import * as v from 'valibot';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/**
 * Push web en este celular (docs/decisiones.md §43): la suscripción del navegador va a Go atada a
 * la sesión. El navegador nunca habla con Go: pasa por acá.
 */
const B64 = v.pipe(v.string(), v.regex(/^[A-Za-z0-9_+/=-]+$/), v.maxLength(200));
const Suscripcion = v.object({
	endpoint: v.pipe(v.string(), v.url(), v.startsWith('https://'), v.maxLength(1000)),
	p256dh: B64,
	auth: B64,
	aparato: v.optional(v.pipe(v.string(), v.maxLength(60)), '')
});

/** La clave pública VAPID (vacía: apagado) y si esta sesión ya está suscripta. */
export const GET: RequestHandler = async ({ locals }) => {
	// El soporte mirando la tienda (o una API que no responde) ve «no disponible», no un error.
	const estado = await datosDe(locals)
		.pushWeb()
		.catch(() => ({ clave: '', suscripta: false }));
	return json(estado, { headers: { 'cache-control': 'no-store' } });
};

export const POST: RequestHandler = async ({ locals, request }) => {
	const r = v.safeParse(Suscripcion, await request.json().catch(() => null));
	if (!r.success) error(400, 'El navegador no mandó una suscripción válida.');
	if ((await datosDe(locals).suscribirPushWeb(r.output)) !== 'ok')
		error(409, 'No pudimos activar los avisos en este celular.');
	return new Response(null, { status: 204 });
};

export const DELETE: RequestHandler = async ({ locals }) => {
	await datosDe(locals).bajaPushWeb();
	return new Response(null, { status: 204 });
};
