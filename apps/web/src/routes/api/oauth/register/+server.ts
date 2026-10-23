import { env } from 'cloudflare:workers';
import { usaSandbox } from '#lib/server/datos/index.ts';
import { golpe } from '#lib/server/limite.ts';
import { json, noDisponible, preflight } from '#lib/server/mcp/http.ts';
import { registrarCliente } from '#lib/server/mcp/oauth.ts';
import { dbSandbox } from '#lib/server/sandbox/sandbox.ts';
import type { RequestHandler } from './$types';

/** Registro dinámico de clientes (RFC 7591), con tope de 20 por hora por IP. */
export const POST: RequestHandler = async ({ request, getClientAddress }) => {
	if (!usaSandbox()) return noDisponible();
	const db = dbSandbox(env.SANDBOX);
	const ip = request.headers.get('cf-connecting-ip') ?? getClientAddress();
	if (!(await golpe(db, `oauth-registro:${ip}`, 20, 3600)))
		return json(
			{ error: 'slow_down', error_description: 'Demasiados registros. Probá en un rato.' },
			429
		);
	const body = (await request.json().catch(() => ({}))) as Record<string, unknown>;
	const c = await registrarCliente(db, body ?? {});
	if (!c)
		return json(
			{
				error: 'invalid_redirect_uri',
				error_description: 'Hace falta al menos una dirección de vuelta https (o localhost).'
			},
			400
		);
	return json(c, 201);
};
export const OPTIONS = preflight;
