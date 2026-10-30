import { error, json } from '@sveltejs/kit';
import { claveVapid, desuscribir, suscribir, type Suscripcion } from '#lib/server/avisos.ts';
import { Jmap } from '#lib/server/jmap.ts';
import type { RequestHandler } from './$types';

/** La clave pública con la que el navegador se suscribe (la de Stalwart). */
export const GET: RequestHandler = async ({ locals }) => {
	const clave = await claveVapid(new Jmap(locals.cred!));
	if (!clave) error(501, 'Este servidor de correo no manda avisos.');
	return json({ clave });
};

export const POST: RequestHandler = async ({ locals, request }) => {
	const b = (await request.json()) as { suscripcion?: Suscripcion; dispositivo?: string };
	const s = b.suscripcion;
	if (!s?.endpoint?.startsWith('https://') || !s.keys?.p256dh || !s.keys?.auth)
		error(400, 'Suscripción inválida.');
	const id = await suscribir(new Jmap(locals.cred!), s, b.dispositivo || 'navegador');
	return json({ id });
};

export const DELETE: RequestHandler = async ({ locals, request }) => {
	const { endpoint } = (await request.json()) as { endpoint?: string };
	if (endpoint) await desuscribir(new Jmap(locals.cred!), endpoint);
	return new Response(null, { status: 204 });
};
