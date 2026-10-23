import { env } from 'cloudflare:workers';
import { usaSandbox } from '#lib/server/datos/index.ts';
import { tokenApi } from '#lib/server/mcp/api.ts';
import { json, parametros, preflight } from '#lib/server/mcp/http.ts';
import { canjearCodigo, renovar } from '#lib/server/mcp/oauth.ts';
import { dbSandbox } from '#lib/server/sandbox/sandbox.ts';
import type { RequestHandler } from './$types';

/** Endpoint de token: authorization_code (con PKCE) y refresh_token (rota). Clientes públicos. */
export const POST: RequestHandler = async ({ request }) => {
	const p = await parametros(request);
	if (!usaSandbox()) {
		const out = await tokenApi(p);
		if (out === undefined) return json({ error: 'unsupported_grant_type' }, 400);
		return out ? json(out) : json({ error: 'invalid_grant' }, 400);
	}
	const db = dbSandbox(env.SANDBOX);
	const grant = p.get('grant_type');
	const clienteId = p.get('client_id') ?? '';
	const out =
		grant === 'authorization_code'
			? await canjearCodigo(db, {
					codigo: p.get('code') ?? '',
					clienteId,
					redireccion: p.get('redirect_uri') ?? '',
					verificador: p.get('code_verifier') ?? ''
				})
			: grant === 'refresh_token'
				? await renovar(db, { refresh: p.get('refresh_token') ?? '', clienteId })
				: undefined;
	if (out === undefined) return json({ error: 'unsupported_grant_type' }, 400);
	if (!out) return json({ error: 'invalid_grant' }, 400);
	return json(out);
};
export const OPTIONS = preflight;
