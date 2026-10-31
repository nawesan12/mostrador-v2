/**
 * El tiempo real de la tienda online (docs/decisiones.md §44), sin sesión: de esa tienda sólo
 * llega que cambió el stock o la ficha de un producto y, en la página de un pedido, que cambió ESE
 * pedido (el código público lo identifica; ni el número viaja). Nada de datos de nadie.
 */
import { invalidate, invalidateAll } from '$app/navigation';
import { ConexionTR } from './conexion.svelte.ts';
import { urlTiempoReal, type EventoTR } from './reglas.ts';

/** Lo que se recarga en la tienda online. */
export const VIDRIERA = 'tienda:vidriera';
export const PEDIDO_PUBLICO = 'tienda:pedido-publico';

export function conectarVidriera(op: {
	slug: string;
	/** El código público del pedido (sólo en su página). */
	codigo?: string;
	/** El origen de la API si no es el de la página (dominio propio, pruebas). */
	origen?: string | null;
	/** Después de recargar por un aviso (para comparar el carrito, avisar el pago…). */
	alRecargar?: (e: EventoTR) => void;
}): ConexionTR | null {
	if (typeof location === 'undefined') return null;
	// eslint-disable-next-line svelte/prefer-svelte-reactivity -- se arma una vez
	const q = new URLSearchParams({ tienda: op.slug });
	if (op.codigo) q.set('pedido', op.codigo);
	let rafaga: ReturnType<typeof setTimeout> | undefined;
	let eventos: EventoTR[] = [];
	const c = new ConexionTR(urlTiempoReal(location, `/v1/tiempo-real/publico?${q}`, op.origen), {
		alEvento: (e) => {
			eventos.push(e);
			clearTimeout(rafaga);
			// Varias ventas juntas: una sola recarga.
			rafaga = setTimeout(async () => {
				const es = eventos;
				eventos = [];
				// eslint-disable-next-line svelte/prefer-svelte-reactivity -- copia local
				const claves = new Set<string>(
					es.map((x) => (x.tipo.startsWith('producto.') ? VIDRIERA : PEDIDO_PUBLICO))
				);
				await invalidate((u) => claves.has(u.href));
				for (const x of es) op.alRecargar?.(x);
			}, 400);
		},
		alResincronizar: () => void invalidateAll()
	});
	c.iniciar();
	return c;
}
