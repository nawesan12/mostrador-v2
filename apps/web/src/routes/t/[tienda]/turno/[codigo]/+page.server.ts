import { error, fail } from '@sveltejs/kit';
import { usaSandbox } from '#lib/server/datos/index.ts';
import { cancelar, turnoDe } from '#lib/server/tienda/turnos.ts';
import type { Actions, PageServerLoad } from './$types';

/** La confirmación del turno: el código es público y aleatorio (no se adivina el de otro). */
export const load: PageServerLoad = async ({ params, url }) => {
	const turno = await turnoDe(params.tienda, params.codigo, usaSandbox());
	if (!turno) error(404, 'No encontramos ese turno.');
	// Vuelta de Mercado Pago (pago=ok|pendiente|error).
	const pago = url.searchParams.get('pago');
	return { turno, pago: pago === 'ok' || pago === 'pendiente' || pago === 'error' ? pago : null };
};

export const actions: Actions = {
	cancelar: async ({ params }) => {
		const r = await cancelar(params.tienda, params.codigo, usaSandbox());
		if (r === 'no-existe') error(404, 'No encontramos ese turno.');
		if (r === 'tarde')
			return fail(409, { error: 'Ya no se puede cancelar online: escribile a la tienda.' });
		return { cancelado: true };
	}
};
