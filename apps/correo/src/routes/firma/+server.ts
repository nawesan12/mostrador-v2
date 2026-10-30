import { json } from '@sveltejs/kit';
import { Jmap } from '#lib/server/jmap.ts';
import type { RequestHandler } from './$types';

/** La firma de una casilla (su identidad), para agregarla al redactar. */
export const GET: RequestHandler = async ({ locals, url }) => {
	const cuenta = url.searchParams.get('cuenta') ?? '';
	const j = new Jmap(locals.cred!);
	const c = (await j.cuentas()).find((x) => x.id === cuenta);
	if (!c) return json({ firma: '' });
	const r = await j.llamar(
		[['Identity/get', { accountId: cuenta }, 'i']],
		['urn:ietf:params:jmap:submission']
	);
	const ids = (r.get('i')?.list ?? []) as { email: string; textSignature: string }[];
	return json({ firma: (ids.find((x) => x.email === c.email) ?? ids[0])?.textSignature ?? '' });
};
