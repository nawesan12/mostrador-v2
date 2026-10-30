import { error, fail, redirect } from '@sveltejs/kit';
import { usaSandbox } from '#lib/server/datos/index.ts';
import { pagarLink, verLink } from '#lib/server/links.ts';
import type { Actions, PageServerLoad } from './$types';
import { irAfuera } from '#lib/server/afuera.ts';

export const load: PageServerLoad = async ({ params, setHeaders }) => {
	const link = await verLink(params.codigo.toUpperCase());
	if (!link) error(404, 'Ese link de pago no existe o ya venció.');
	setHeaders({ 'cache-control': 'no-store' });
	return { link, demo: usaSandbox() };
};

export const actions: Actions = {
	default: async ({ params }) => {
		const r = await pagarLink(params.codigo.toUpperCase());
		if (r === 'no-existe') error(404, 'Ese link de pago no existe o ya venció.');
		if (r === 'ya-pagado') return fail(409, { error: 'Este link ya se pagó.' });
		if (r === 'sin-cobros')
			return fail(409, {
				error: 'La tienda todavía no puede cobrar con tarjeta. Escribile para pagar de otra forma.'
			});
		if (r === 'error')
			return fail(502, { error: 'Mercado Pago no respondió. Probá de nuevo en un rato.' });
		if (typeof r === 'object') irAfuera(r.url);
		return { pagado: true };
	}
};
