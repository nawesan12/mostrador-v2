import { env } from 'cloudflare:workers';
import { error } from '@sveltejs/kit';
import * as v from 'valibot';
import { usaSandbox } from '#lib/server/datos/index.ts';
import { desconectarApi } from '#lib/server/mcp/api.ts';
import { desconectar } from '#lib/server/mcp/oauth.ts';
import { dbSandbox } from '#lib/server/sandbox/sandbox.ts';
import type { RequestHandler } from './$types';

const Pedido = v.object({ id: v.pipe(v.string(), v.minLength(1), v.maxLength(40)) });

/** Desconectar una IA: revoca sus tokens (la conexión se borra entera). */
export const DELETE: RequestHandler = async ({ locals, request }) => {
	const b = v.safeParse(Pedido, await request.json().catch(() => null));
	if (!b.success) error(400, 'Pedido inválido.');
	if (!usaSandbox()) {
		if (!(await desconectarApi(locals.token, b.output.id)))
			error(404, 'Esa conexión ya no existe.');
		return new Response(null, { status: 204 });
	}
	if (!locals.sandbox) error(401, 'Tu demo venció. Elegí una tienda de nuevo.');
	if (!(await desconectar(dbSandbox(env.SANDBOX), locals.sandbox.id, b.output.id)))
		error(404, 'Esa conexión ya no existe.');
	return new Response(null, { status: 204 });
};
