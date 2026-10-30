import { tiendaDelBorrador } from '#lib/server/compartir/borrador.ts';
import { hechosDe } from '#lib/tienda/compartir.ts';
import type { PageServerLoad } from './$types';

/** El editor, más lo que se sabe de la tienda para sugerir cómo se ve al compartirla. */
export const load: PageServerLoad = async ({ locals, cookies }) => {
	const { mi, tienda } = await tiendaDelBorrador(locals, cookies);
	return { ...mi, hechos: hechosDe(tienda) };
};
