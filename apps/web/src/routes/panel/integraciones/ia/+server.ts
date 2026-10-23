import { env } from 'cloudflare:workers';
import { error } from '@sveltejs/kit';
import * as v from 'valibot';
import { usaSandbox } from '#lib/server/datos/index.ts';
import { desconectar, NO_DISPONIBLE } from '#lib/server/mcp/oauth.ts';
import { dbSandbox } from '#lib/server/sandbox/sandbox.ts';
import type { RequestHandler } from './$types';

const Pedido = v.object({ id: v.pipe(v.string(), v.minLength(1), v.maxLength(40)) });

/** Desconectar una IA: revoca sus tokens (la conexión se borra entera). */
export const DELETE: RequestHandler = async ({ locals, request }) => {
	if (!usaSandbox()) error(501, NO_DISPONIBLE);
	if (!locals.sandbox) error(401, 'Tu demo venció. Elegí una tienda de nuevo.');
	const b = v.safeParse(Pedido, await request.json().catch(() => null));
	if (!b.success) error(400, 'Pedido inválido.');
	if (!(await desconectar(dbSandbox(env.SANDBOX), locals.sandbox.id, b.output.id)))
		error(404, 'Esa conexión ya no existe.');
	return new Response(null, { status: 204 });
};
