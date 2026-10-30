import { json } from '@sveltejs/kit';
import { Jmap } from '#lib/server/jmap.ts';
import { enviar, type Nuevo } from '#lib/server/enviar.ts';
import type { RequestHandler } from './$types';

export const POST: RequestHandler = async ({ locals, request }) => {
	const n = (await request.json()) as Nuevo;
	try {
		await enviar(new Jmap(locals.cred!), {
			cuenta: String(n.cuenta ?? ''),
			para: (n.para ?? []).map(String),
			cc: (n.cc ?? []).map(String),
			asunto: String(n.asunto ?? '').slice(0, 300),
			texto: String(n.texto ?? '').slice(0, 500_000),
			adjuntos: (n.adjuntos ?? []).slice(0, 20),
			responde: n.responde ? String(n.responde) : null,
			programado: n.programado ? String(n.programado) : null
		});
	} catch (e) {
		return json({ error: e instanceof Error ? e.message : 'No se pudo mandar.' }, { status: 400 });
	}
	return json({ ok: true });
};
