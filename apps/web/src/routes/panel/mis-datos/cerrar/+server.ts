import { error } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import { COOKIE } from '#lib/server/sandbox/sandbox.ts';
import type { RequestHandler } from './$types';

/** Cerrar la cuenta (con el nombre de la tienda como confirmación). */
export const POST: RequestHandler = async ({ locals, request, cookies }) => {
	const c = (await request.json().catch(() => null)) as { confirmacion?: unknown } | null;
	const r = await datosDe(locals).cerrarCuenta(String(c?.confirmacion ?? ''));
	if (r === 'no-se-puede') error(400, 'Escribí el nombre de la tienda para confirmar.');
	// En la demo se borró todo: la cookie ya no apunta a nada.
	cookies.delete(COOKIE, { path: '/' });
	return new Response(null, { status: 204 });
};
