import { equipoDe, exigir } from '#lib/server/staff/index.ts';
import type { PageServerLoad } from './$types';

/** Quién del equipo hizo qué y cuándo (incluido entrar a ver una tienda), con filtros en la URL. */
export const load: PageServerLoad = async (event) => {
	const { sesion } = await event.parent();
	exigir(sesion.permisos, 'auditoria.ver');
	const q = event.url.searchParams;
	return {
		auditoria: await equipoDe(event).auditoria({
			miembro: q.get('miembro') || undefined,
			accion: q.get('accion') || undefined,
			tienda: q.get('tienda') || undefined
		}),
		filtros: { miembro: q.get('miembro') ?? '', accion: q.get('accion') ?? '' }
	};
};
