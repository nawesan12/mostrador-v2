/**
 * Contar una vista de la tienda online (decisiones §36): lo llaman `hooks.server.ts` (carga
 * completa de una página) y /t/[tienda]/vista (navegación dentro de la tienda). La medición en sí
 * está en `../visitas.ts`; acá se arma con la API real y se aplican los filtros del pedido.
 */
import type { RequestEvent } from '@sveltejs/kit';
import { env } from 'cloudflare:workers';
import { esBot, esPrefetch, Medidor, paginaDe } from '../visitas.ts';

/** Sólo con la API real: en la demo (sandbox) y en las vitrinas no se mide nada. */
export const medidor = new Medidor({
	api: () =>
		String(env.DATA_SOURCE) === 'api' && env.API_URL && env.CLAVE_BFF
			? { url: String(env.API_URL), clave: String(env.CLAVE_BFF) }
			: null
});

type Evento = Pick<RequestEvent, 'request' | 'getClientAddress' | 'locals' | 'params'>;

/** Cuenta la vista de `ruta` si es de una persona (no un bot ni un prefetch). Nunca tira. */
export function contarVista(event: Evento, ruta: string | null | undefined, producto?: string) {
	try {
		const h = event.request.headers;
		const slug = event.params.tienda;
		const p = paginaDe(ruta, producto);
		// Alguien del equipo de Tu Local Web mirando una tienda no es una visita.
		if (!p || !slug || event.locals.mirada || esPrefetch(h)) return false;
		const agente = h.get('user-agent') ?? '';
		if (esBot(agente)) return false;
		let ip = '';
		try {
			ip = h.get('cf-connecting-ip') ?? event.getClientAddress();
		} catch {
			return false;
		}
		return medidor.anotar({
			tienda: slug,
			pagina: p.pagina,
			producto: p.producto,
			ip,
			agente,
			sesion: event.locals.token
		});
	} catch {
		return false;
	}
}
