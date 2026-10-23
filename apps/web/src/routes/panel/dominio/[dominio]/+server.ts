import { error } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

const resp = (r: string) => {
	if (r === 'no-existe') error(404, 'Ese dominio no está agregado.');
	if (r === 'no-se-puede') error(409, 'Todavía se está verificando.');
	return new Response(null, { status: 204 });
};

/** Hacerlo principal. */
export const PUT: RequestHandler = async ({ locals, params }) =>
	resp(await datosDe(locals).dominioPrincipal(params.dominio));
/** Quitarlo. */
export const DELETE: RequestHandler = async ({ locals, params }) =>
	resp(await datosDe(locals).quitarDominio(params.dominio));
