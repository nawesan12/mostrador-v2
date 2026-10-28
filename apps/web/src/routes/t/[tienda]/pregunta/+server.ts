import { preguntar } from '#lib/server/tienda/opiniones.ts';
import type { RequestHandler } from './$types';

/** Una pregunta sobre un producto: la tienda la responde desde el panel. */
export const POST: RequestHandler = (event) => preguntar(event);
