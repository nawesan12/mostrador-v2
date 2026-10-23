import type * as v from 'valibot';
import { ClienteBff } from '../bff.ts';
import type { NuevaPregunta, NuevaResena } from './esquemas.ts';

/** Las dos operaciones públicas en la API Go (con la IP del visitante: Go pone el tope). */
export class VidrieraApi extends ClienteBff {
	resena(slug: string, r: v.InferOutput<typeof NuevaResena>) {
		return this.pedir<void>('POST', `/v1/publico/tiendas/${slug}/resenas`, r);
	}
	pregunta(slug: string, q: v.InferOutput<typeof NuevaPregunta>) {
		return this.pedir<void>('POST', `/v1/publico/tiendas/${slug}/preguntas`, q);
	}
}
