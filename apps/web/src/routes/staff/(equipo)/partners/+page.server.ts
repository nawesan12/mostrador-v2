import { fail } from '@sveltejs/kit';
import { ErrorApi } from '#lib/server/bff.ts';
import { EnLaDemo } from '#lib/server/staff/demo.ts';
import { equipoDe, exigir } from '#lib/server/staff/index.ts';
import type { Actions, PageServerLoad } from './$types';

/** Postulaciones de partners y los activos (con sus tiendas y comisiones). */
export const load: PageServerLoad = async (event) => {
	const { sesion } = await event.parent();
	exigir(sesion.permisos, 'partners.ver');
	return { partners: await equipoDe(event).partners() };
};

export const actions: Actions = {
	decidir: async (event) => {
		const f = await event.request.formData();
		const decision = f.get('decision') === 'aprobar' ? 'aprobar' : 'rechazar';
		const motivo = String(f.get('motivo') ?? '').trim();
		if (decision === 'rechazar' && motivo.length < 5)
			return fail(400, { error: 'Contá el motivo del rechazo: le llega por mail.' });
		try {
			await equipoDe(event).decidir(String(f.get('id')), decision, motivo);
			return { decision };
		} catch (e) {
			if (e instanceof EnLaDemo) return fail(409, { demo: true });
			if (e instanceof ErrorApi) return fail(e.estado, { error: e.message });
			throw e;
		}
	}
};
