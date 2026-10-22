import { error } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** "Recordar" un carrito (id), "todos" o guardar los automáticos ("automaticos"). */
export const POST: RequestHandler = async ({ locals, params, request }) => {
	const datos = datosDe(locals);
	if (params.id === 'todos') return Response.json({ enviados: await datos.recordarCarritos() });
	if (params.id === 'automaticos') {
		const a = (await request.json().catch(() => null)) as {
			hora?: unknown;
			dia?: unknown;
			email?: unknown;
		} | null;
		await datos.guardarAutomaticos({
			hora: a?.hora === true,
			dia: a?.dia === true,
			email: a?.email === true
		});
		return new Response(null, { status: 204 });
	}
	const r = await datos.recordarCarrito(params.id);
	if (r === 'no-existe') error(404, 'Ese carrito no existe.');
	if (r === 'no-se-puede') error(409, 'Ya le mandaste el recordatorio.');
	return new Response(null, { status: 204 });
};
