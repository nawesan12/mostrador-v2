import { error } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** Mandar otra vez la invitación (link nuevo, otros 7 días; el anterior deja de valer). */
export const POST: RequestHandler = async ({ locals, params }) => {
	const r = await datosDe(locals).reenviarInvitacion(params.id);
	if (r === 'no-existe') error(404, 'Esa invitación ya no está.');
	return new Response(null, { status: 204 });
};
