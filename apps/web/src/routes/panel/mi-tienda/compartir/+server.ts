import { responderImagen } from '#lib/server/compartir/index.ts';
import { tiendaDelBorrador } from '#lib/server/compartir/borrador.ts';
import { tarjetaDe } from '#lib/tienda/compartir.ts';
import type { RequestHandler } from './$types';

/** La vista previa de la imagen al compartir, con el borrador de Mi tienda (sin cachear). */
export const GET: RequestHandler = async ({ locals, cookies }) => {
	const { tienda } = await tiendaDelBorrador(locals, cookies);
	return responderImagen(tarjetaDe(tienda), `b-${tienda.slug}`, 'private, no-store');
};
