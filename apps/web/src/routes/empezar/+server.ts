import { redirect } from '@sveltejs/kit';
import { guardarOrigen, leerOrigen, origenDe } from '#lib/server/origen.ts';
import { contarClic } from '#lib/server/partners/index.ts';
import type { RequestHandler } from './$types';

/**
 * La entrada de los QR y links de campaña (/empezar?utm_source=flyer, ?ref=…, ?cupon=…): anota de
 * dónde vino y lleva a crear la cuenta. El primer origen manda (no se pisa con una visita
 * posterior), como la comisión de un partner. La primera visita con el link de un partner suma un
 * clic en su panel (sin esperarlo: va después de responder).
 */
export const GET: RequestHandler = (event) => {
	const { url, cookies, platform } = event;
	const o = origenDe(url);
	if (o && !leerOrigen(cookies)) {
		guardarOrigen(cookies, o);
		if (o.ref) {
			const clic = contarClic(event, o.ref);
			if (platform?.ctx) platform.ctx.waitUntil(clic);
		}
	}
	redirect(303, '/ingresar?m=registro');
};
