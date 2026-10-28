import { redirect } from '@sveltejs/kit';
import { usaSandbox } from '#lib/server/datos/index.ts';
import { accesoEquipo, borrarSesionEquipo } from '#lib/server/staff/index.ts';
import type { RequestHandler } from './$types';

/** Cierra la sesión del equipo (en Go y la cookie). */
export const POST: RequestHandler = async (event) => {
	if (!usaSandbox() && event.locals.equipo)
		await accesoEquipo(event)
			.salir(event.locals.equipo)
			.catch(() => {});
	borrarSesionEquipo(event.cookies);
	redirect(303, usaSandbox() ? '/demo' : '/staff/ingresar');
};
