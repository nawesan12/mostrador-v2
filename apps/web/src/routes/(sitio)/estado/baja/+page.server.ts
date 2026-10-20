import { env } from 'cloudflare:workers';
import type { Actions, PageServerLoad } from './$types';

// Como en novedades: el link del aviso sólo muestra el botón; la baja es un POST.
export const load: PageServerLoad = ({ url }) => ({ token: url.searchParams.get('token') ?? '' });

export const actions: Actions = {
	default: async ({ request }) => {
		const token = String((await request.formData()).get('token') ?? '');
		if (token)
			await env.ESTADO.prepare('DELETE FROM suscripciones WHERE token_baja = ?1').bind(token).run();
		// Siempre "listo": no se le dice a nadie si un token existía.
		return { ok: true };
	}
};
