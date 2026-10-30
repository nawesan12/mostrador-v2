import { error } from '@sveltejs/kit';
import { datosDe, usaSandbox } from '#lib/server/datos/index.ts';
import type { Idea } from '#lib/server/datos/tipos.ts';
import type { RequestHandler } from './$types';

const IDEAS: Idea[] = ['carritos', 'stock', 'promo'];

/** Aprobar ({hecha: true}) o deshacer ({hecha: false}) una idea de Tincho. */
export const POST: RequestHandler = async ({ locals, params, request }) => {
	const idea = params.idea as Idea;
	if (!IDEAS.includes(idea)) error(404, 'Esa idea no existe.');
	const cuerpo = (await request.json().catch(() => null)) as { hecha?: unknown } | null;
	if (typeof cuerpo?.hecha !== 'boolean') error(400, 'Falta si la idea queda hecha.');
	const datos = datosDe(locals);
	if (!usaSandbox()) {
		// Con la API, la única idea que Tincho hace solo es recordar los carritos (un mail, que no se
		// puede deshacer). Las otras se hacen desde su sección.
		if (idea !== 'carritos' || !cuerpo.hecha) error(409, 'Esa idea se hace desde su sección.');
		await datos.recordarCarritos();
		return new Response(null, { status: 204 });
	}
	await datos.marcarIdea(idea, cuerpo.hecha);
	return new Response(null, { status: 204 });
};
