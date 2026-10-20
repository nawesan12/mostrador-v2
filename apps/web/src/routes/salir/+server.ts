import { redirect } from '@sveltejs/kit';
import { cuentasDe, usaSandbox } from '#lib/server/datos/index.ts';
import { borrarSesion } from '#lib/server/auth/sesion.ts';
import type { RequestHandler } from './$types';

/** Cerrar sesión: POST (un link no puede sacar a nadie de su cuenta). */
export const POST: RequestHandler = async (event) => {
	if (!usaSandbox() && event.locals.token) {
		// Si la API no responde, igual se borra la cookie: la sesión vence sola en Go.
		await cuentasDe(event)
			.salir(event.locals.token)
			.catch(() => {});
	}
	borrarSesion(event.cookies);
	redirect(303, '/ingresar');
};
