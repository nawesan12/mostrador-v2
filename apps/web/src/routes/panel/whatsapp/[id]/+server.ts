import { error } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** Mandar un mensaje ({texto}), tomar o devolver ({tomada}) o marcar leída ({leida}). */
export const POST: RequestHandler = async ({ locals, params, request }) => {
	const c = (await request.json().catch(() => null)) as {
		texto?: unknown;
		tomada?: unknown;
		leida?: unknown;
	} | null;
	const texto = typeof c?.texto === 'string' ? c.texto.trim().slice(0, 2000) : undefined;
	const tomada = typeof c?.tomada === 'boolean' ? c.tomada : undefined;
	if (!texto && tomada === undefined && c?.leida !== true) error(400, 'Escribí un mensaje.');
	const r = await datosDe(locals).responder(params.id, { texto, tomada, leida: c?.leida === true });
	if (r === 'no-existe') error(404, 'Esa conversación no existe.');
	if (r === 'no-se-puede') error(409, 'No se pudo mandar.');
	return new Response(null, { status: 204 });
};
