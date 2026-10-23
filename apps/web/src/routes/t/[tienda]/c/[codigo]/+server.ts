import { redirect } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

/**
 * El link de un cupón (el que comparte la tienda en historias o por WhatsApp): deja el código anotado
 * para esa tienda y lleva al catálogo; el checkout lo trae cargado y lo aplica (CampoCupon). Si el
 * cupón no vale, el checkout dice por qué: acá no se valida para no adelantar nada.
 */
export const GET: RequestHandler = ({ params, cookies }) => {
	const slug = params.tienda;
	const codigo = params.codigo.trim().toUpperCase();
	if (/^[a-z0-9-]{2,60}$/.test(slug) && /^[A-Z0-9-]{3,20}$/.test(codigo))
		cookies.set(`cupon-${slug}`, codigo, {
			path: `/t/${slug}`,
			httpOnly: false,
			sameSite: 'lax',
			maxAge: 7 * 86_400
		});
	redirect(303, `/t/${slug}/catalogo`);
};
