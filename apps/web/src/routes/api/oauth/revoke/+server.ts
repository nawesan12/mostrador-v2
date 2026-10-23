import { env } from 'cloudflare:workers';
import { usaSandbox } from '#lib/server/datos/index.ts';
import { revocarApi } from '#lib/server/mcp/api.ts';
import { CORS, parametros, preflight } from '#lib/server/mcp/http.ts';
import { revocar } from '#lib/server/mcp/oauth.ts';
import { dbSandbox } from '#lib/server/sandbox/sandbox.ts';
import type { RequestHandler } from './$types';

/** Revocación (RFC 7009): siempre 200, exista o no el token (no confirma tokens ajenos). */
export const POST: RequestHandler = async ({ request }) => {
	const p = await parametros(request);
	if (usaSandbox()) await revocar(dbSandbox(env.SANDBOX), p.get('token') ?? '');
	else if (!(await revocarApi(p.get('token') ?? '')))
		// RFC 7009 §2.2.1: con 503 la IA sabe que el token puede seguir vivo y reintenta.
		return new Response(null, {
			status: 503,
			headers: { ...CORS, 'Cache-Control': 'no-store', 'Retry-After': '30' }
		});
	return new Response(null, { status: 200, headers: { ...CORS, 'Cache-Control': 'no-store' } });
};
export const OPTIONS = preflight;
