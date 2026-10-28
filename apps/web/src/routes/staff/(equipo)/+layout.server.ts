import { redirect } from '@sveltejs/kit';
import { ErrorApi } from '#lib/server/bff.ts';
import { borrarSesionEquipo, equipoDe } from '#lib/server/staff/index.ts';
import type { LayoutServerLoad } from './$types';

/** Quién está del equipo y qué puede hacer (la navegación sale de sus permisos). */
export const load: LayoutServerLoad = async (event) => {
	const datos = equipoDe(event);
	try {
		return { sesion: await datos.sesion(), demo: datos.demo };
	} catch (err) {
		if (err instanceof ErrorApi && err.estado === 401) {
			borrarSesionEquipo(event.cookies);
			redirect(303, '/staff/ingresar');
		}
		throw err;
	}
};
