import { env } from 'cloudflare:workers';
import { error, json } from '@sveltejs/kit';
import { ErrorApi } from '#lib/server/bff.ts';
import { usaSandbox } from '#lib/server/datos/index.ts';
import { dbSandbox } from '#lib/server/sandbox/sandbox.ts';
import {
	desvincular,
	desvincularApi,
	habilitado,
	linkVinculo,
	linkVinculoApi
} from '#lib/server/telegram.ts';
import type { RequestHandler } from './$types';

/** Un error de Go con su mensaje (sin permiso, sin sesión); lo demás, que reintente. */
function falla(e: unknown): never {
	if (e instanceof ErrorApi && e.estado < 500) error(e.estado, e.message);
	console.error('telegram', e instanceof Error ? e.message : e);
	error(503, 'No pudimos hablar con Telegram. Probá de nuevo en un rato.');
}

/** Vincular: un link al bot con un código de un solo uso (10 minutos). */
export const POST: RequestHandler = async ({ locals }) => {
	if (!habilitado()) error(409, 'Tincho en Telegram todavía no está disponible.');
	if (!usaSandbox()) {
		if (!locals.token) error(401, 'Tu sesión venció. Volvé a ingresar.');
		return json({ url: await linkVinculoApi(locals.token).catch(falla) });
	}
	if (!locals.sandbox)
		error(409, 'Tincho en Telegram llega a tu cuenta cuando la tienda esté en marcha.');
	return json({ url: await linkVinculo(dbSandbox(env.SANDBOX), locals.sandbox.id) });
};

/** Desvincular el chat. */
export const DELETE: RequestHandler = async ({ locals }) => {
	if (!usaSandbox()) {
		if (!locals.token) error(401, 'Tu sesión venció. Volvé a ingresar.');
		await desvincularApi(locals.token).catch(falla);
	} else if (locals.sandbox) await desvincular(dbSandbox(env.SANDBOX), locals.sandbox.id);
	return new Response(null, { status: 204 });
};
