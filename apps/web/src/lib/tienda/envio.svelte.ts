/**
 * La cotización de envío que usan las plantillas: la de envio.ts más la de Correo Argentino, que se
 * pide a la API (por el BFF, /t/<tienda>/correo) la primera vez que aparece un CP y se guarda por
 * tienda, CP y unidades. Mientras no llega, Correo no aparece: nunca un precio inventado.
 */
import { SvelteMap } from 'svelte/reactivity';
import { carritoDe } from './carrito.svelte.ts';
import {
	comoLlega,
	cotizar as cotizarBase,
	cpValido,
	entregaCorta,
	envia,
	esCorreo,
	resumenMoto,
	textoEnvios,
	type CotizacionCorreo
} from './envio.ts';
import type { OpcionEnvio, TiendaPublica } from './tipos.ts';

export { comoLlega, cpValido, entregaCorta, envia, esCorreo, resumenMoto, textoEnvios };

type Estado = CotizacionCorreo | 'error';
type Tienda = Pick<TiendaPublica, 'slug' | 'envios' | 'direccion' | 'vitrina'>;

const cotizaciones = new SvelteMap<string, Estado>();
// En vuelo: no es estado (se lee dentro de $derived, y ahí no se puede escribir estado).
// eslint-disable-next-line svelte/prefer-svelte-reactivity -- a propósito, no reactivo
const pedidos = new Map<string, Promise<void>>();

/** Las unidades del carrito (el peso del paquete sale de ahí); al menos 1 para la ficha. */
const unidadesDe = (slug: string) => Math.max(1, carritoDe(slug).unidades);

const clave = (slug: string, cp: string, unidades: number) => `${slug}|${cp.trim()}|${unidades}`;

function pedir(t: Tienda, cp: string, unidades: number) {
	const k = clave(t.slug, cp, unidades);
	let p = pedidos.get(k);
	if (!p) {
		const q = new URLSearchParams({ cp: cp.trim(), unidades: String(unidades) });
		p = fetch(`/t/${encodeURIComponent(t.slug)}/correo?${q}`)
			.then(async (r) => {
				if (!r.ok) throw new Error(String(r.status));
				cotizaciones.set(k, (await r.json()) as CotizacionCorreo);
			})
			.catch(() => {
				cotizaciones.set(k, 'error');
				// Se puede reintentar (reintentarCorreo).
				pedidos.delete(k);
			});
		pedidos.set(k, p);
	}
	return p;
}

const conCorreo = (t: Tienda, cp: string) =>
	t.envios.correo && !t.vitrina && typeof window !== 'undefined' && cpValido(cp);

/** Las opciones para un CP. Reactiva: cuando llega la cotización de Correo, se recalcula. */
export function cotizar(cp: string, subtotal: number, t: Tienda): OpcionEnvio[] {
	let correo: CotizacionCorreo | null = null;
	if (conCorreo(t, cp)) {
		const u = unidadesDe(t.slug);
		const e = cotizaciones.get(clave(t.slug, cp, u));
		if (e === undefined) void pedir(t, cp, u);
		else if (e !== 'error') correo = e;
	}
	return cotizarBase(cp, subtotal, t, correo);
}

/** Igual, pero espera a Correo (el "Calcular" de la ficha del producto, que no es reactivo). */
export async function cotizarYa(cp: string, subtotal: number, t: Tienda) {
	if (conCorreo(t, cp)) await pedir(t, cp, unidadesDe(t.slug));
	return cotizar(cp, subtotal, t);
}

/** Cómo va la cotización de Correo para el checkout: null si no corresponde. */
export function estadoCorreo(cp: string, t: Tienda): 'cotizando' | 'error' | 'ok' | null {
	if (!conCorreo(t, cp)) return null;
	const e = cotizaciones.get(clave(t.slug, cp, unidadesDe(t.slug)));
	return e === undefined ? 'cotizando' : e === 'error' ? 'error' : 'ok';
}

export function reintentarCorreo(cp: string, t: Tienda) {
	const k = clave(t.slug, cp, unidadesDe(t.slug));
	if (cotizaciones.get(k) === 'error') {
		cotizaciones.delete(k);
		void pedir(t, cp, unidadesDe(t.slug));
	}
}
