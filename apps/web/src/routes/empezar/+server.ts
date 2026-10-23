import { redirect } from '@sveltejs/kit';
import { guardarOrigen, leerOrigen, origenDe } from '#lib/server/origen.ts';
import type { RequestHandler } from './$types';

/**
 * La entrada de los QR y links de campaña (/empezar?utm_source=flyer, ?ref=…, ?cupon=…): anota de
 * dónde vino y lleva a crear la cuenta. El primer origen manda (no se pisa con una visita
 * posterior), como la comisión de un partner.
 */
export const GET: RequestHandler = ({ url, cookies }) => {
	const o = origenDe(url);
	if (o && !leerOrigen(cookies)) guardarOrigen(cookies, o);
	redirect(303, '/ingresar?m=registro');
};
