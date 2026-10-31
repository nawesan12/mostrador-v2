import { error } from '@sveltejs/kit';
import { planVigente } from '#lib/panel/planes.ts';
import { datosDe } from '#lib/server/datos/index.ts';
import type { PageServerLoad } from './$types';

/** El equipo es sólo de la dueña (Go responde 403 a las demás si igual lo intentan). */
export const load: PageServerLoad = async ({ locals, parent, depends }) => {
	// Tiempo real (§44): se recarga con los avisos de lo que muestra.
	depends('tienda:equipo');
	const [{ acceso }, c] = await Promise.all([parent(), datosDe(locals).configuracion()]);
	if (!acceso.permisos.includes('cuenta'))
		error(403, 'Sólo la dueña de la tienda maneja el equipo.');
	return {
		equipo: c.equipo,
		plan: planVigente(c.plan.actual, c.plan.enPrueba)
	};
};
