import { json } from '@sveltejs/kit';
import { nuevos } from '#lib/server/avisos.ts';
import { Jmap } from '#lib/server/jmap.ts';
import type { RequestHandler } from './$types';

/** Lo que entró después de `desde` (el service worker lo pide al recibir un push). */
export const GET: RequestHandler = async ({ locals, url }) => {
	const d = new Date(url.searchParams.get('desde') ?? '');
	const desde =
		Number.isNaN(+d) || Date.now() - +d > 86_400_000 ? new Date(Date.now() - 10 * 60_000) : d;
	return json({
		avisos: await nuevos(new Jmap(locals.cred!), desde.toISOString().replace(/\.\d+Z$/, 'Z'))
	});
};
