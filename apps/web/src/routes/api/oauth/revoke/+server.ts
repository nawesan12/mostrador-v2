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
	else await revocarApi(p.get('token') ?? '');
	return new Response(null, { status: 200, headers: { ...CORS, 'Cache-Control': 'no-store' } });
};
export const OPTIONS = preflight;
