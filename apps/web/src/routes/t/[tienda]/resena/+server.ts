import { dejarResena } from '#lib/server/tienda/opiniones.ts';
import type { RequestHandler } from './$types';

/** La reseña que deja quien compró (con el código del pedido queda con compra). */
export const POST: RequestHandler = (event) => dejarResena(event);
