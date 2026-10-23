import { env } from 'cloudflare:workers';
import { error, json } from '@sveltejs/kit';
import { dbSandbox } from '#lib/server/sandbox/sandbox.ts';
import { desvincular, habilitado, linkVinculo } from '#lib/server/telegram.ts';
import type { RequestHandler } from './$types';

/** Vincular: un link al bot con un código de un solo uso (10 minutos). */
export const POST: RequestHandler = async ({ locals }) => {
	if (!habilitado() || !locals.sandbox)
		error(409, 'Hornero en Telegram llega a tu cuenta cuando la tienda esté en marcha.');
	return json({ url: await linkVinculo(dbSandbox(env.SANDBOX), locals.sandbox.id) });
};

/** Desvincular el chat. */
export const DELETE: RequestHandler = async ({ locals }) => {
	if (locals.sandbox) await desvincular(dbSandbox(env.SANDBOX), locals.sandbox.id);
	return new Response(null, { status: 204 });
};
