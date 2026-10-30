import { error } from '@sveltejs/kit';
import { verificar } from '#lib/server/avisos.ts';
import { Jmap } from '#lib/server/jmap.ts';
import type { RequestHandler } from './$types';

/** El código que Stalwart mandó en el primer push: confirma que el navegador recibe. */
export const POST: RequestHandler = async ({ locals, request }) => {
	const { id, codigo } = (await request.json()) as { id?: string; codigo?: string };
	if (!id || !codigo) error(400, 'Falta el código.');
	await verificar(new Jmap(locals.cred!), id, codigo);
	return new Response(null, { status: 204 });
};
