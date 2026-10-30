import { error, redirect } from '@sveltejs/kit';

/** Los sitios a los que la app manda a la gente: pagar o conectar Mercado Pago y Mercado Libre. */
const CONFIABLES = /^(www\.|auth\.|sandbox\.)?mercado(pago|libre)\.com(\.ar)?$/;

/**
 * Redirige a un sitio externo de confianza. SvelteKit 3 no deja redirigir afuera sin decir a qué
 * origen; acá se valida el host antes, para no abrir una redirección a cualquier lado.
 */
export function irAfuera(url: string): never {
	let u: URL;
	try {
		u = new URL(url);
	} catch {
		error(502, 'Respuesta inesperada.');
	}
	if (u.protocol !== 'https:' || !CONFIABLES.test(u.hostname)) error(502, 'Respuesta inesperada.');
	redirect(303, u.href, { external: [u.origin] });
}
