import { habilitado, procesar, secretoValido } from '#lib/server/telegram.ts';
import type { RequestHandler } from './$types';

/**
 * Webhook del bot de Telegram. Primero el secreto (X-Telegram-Bot-Api-Secret-Token); después se
 * responde 200 enseguida y el mensaje se procesa aparte (Telegram reintenta si tarda).
 */
export const POST: RequestHandler = async ({ request, platform }) => {
	if (!habilitado()) return new Response(null, { status: 404 });
	if (!secretoValido(request.headers.get('x-telegram-bot-api-secret-token')))
		return new Response(null, { status: 401 });
	const u = await request.json().catch(() => null);
	if (!u || typeof u !== 'object') return new Response(null, { status: 400 });
	const trabajo = procesar(u as Parameters<typeof procesar>[0]).catch((e) =>
		console.error('telegram', e instanceof Error ? e.message : e)
	);
	if (platform?.ctx) platform.ctx.waitUntil(trabajo);
	else await trabajo;
	return new Response(null, { status: 200 });
};
