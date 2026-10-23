import { error } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** Un dominio con al menos un punto, sin protocolo ni barra. */
const DOMINIO = /^(?=.{4,253}$)([a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,}$/;

/** Conectar un dominio propio. */
export const POST: RequestHandler = async ({ locals, request }) => {
	const c = (await request.json().catch(() => null)) as { dominio?: unknown } | null;
	const d = String(c?.dominio ?? '')
		.trim()
		.toLowerCase()
		.replace(/^https?:\/\//, '')
		.replace(/\/.*$/, '');
	if (!DOMINIO.test(d)) error(400, 'Escribí el dominio, por ejemplo tutienda.com.ar');
	if ((await datosDe(locals).conectarDominio(d)) === 'no-se-puede')
		error(409, 'Ese dominio ya está agregado.');
	return new Response(null, { status: 201 });
};
