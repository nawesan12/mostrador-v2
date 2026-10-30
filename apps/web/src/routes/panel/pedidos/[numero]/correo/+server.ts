import { error, json, type RequestEvent } from '@sveltejs/kit';
import { ErrorApi } from '#lib/server/bff.ts';
import { correoDe } from '#lib/server/correo-argentino.ts';
import { usaSandbox } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** El envío por Correo Argentino de un pedido. En la demo no hay cuenta de MiCorreo. */
function contexto(event: RequestEvent<{ numero: string }>) {
	const n = Number(event.params.numero);
	if (!Number.isSafeInteger(n) || n < 1) error(404, 'Ese pedido no existe.');
	if (usaSandbox()) return { n, token: null };
	if (!event.locals.token) error(401, 'Tu sesión venció. Volvé a ingresar.');
	return { n, token: event.locals.token };
}

function falla(e: unknown): never {
	if (e instanceof ErrorApi) error(e.estado, e.message);
	console.error('correo', e instanceof Error ? e.message : e);
	error(503, 'No pudimos hablar con Correo Argentino. Probá de nuevo en un rato.');
}

const DEMO = 'Es una tienda de ejemplo: los envíos se cargan en una cuenta de MiCorreo real.';

export const GET: RequestHandler = async (event) => {
	const { n, token } = contexto(event);
	if (!token) return json({ conectada: false, generado: false });
	return json(await correoDe(event).envio(token, n).catch(falla), {
		headers: { 'cache-control': 'no-store' }
	});
};

/** Generar el envío en MiCorreo. */
export const POST: RequestHandler = async (event) => {
	const { n, token } = contexto(event);
	if (!token) error(409, DEMO);
	return json(await correoDe(event).generar(token, n).catch(falla));
};

/** {seguimiento}: el número que dio MiCorreo al pagar el rótulo. */
export const PUT: RequestHandler = async (event) => {
	const { n, token } = contexto(event);
	if (!token) error(409, DEMO);
	const b = (await event.request.json().catch(() => null)) as { seguimiento?: unknown } | null;
	if (typeof b?.seguimiento !== 'string' || b.seguimiento.length > 60)
		error(400, 'Poné el número de seguimiento.');
	return json(await correoDe(event).seguimiento(token, n, b.seguimiento).catch(falla));
};
