import { SITE } from '#lib/site.ts';
import type { UiState } from '#lib/ui.svelte.ts';

/** El link público de la tienda. */
export const linkTienda = (slug: string) => `${SITE}/t/${slug}`;

/**
 * Compartir la tienda: en el celular, la hoja de compartir del sistema (WhatsApp, Instagram, lo que
 * tenga); en la compu, el link queda copiado y el aviso ofrece mandarlo por WhatsApp.
 */
export async function compartirTienda(ui: UiState, tienda: { nombre: string; slug: string }) {
	const url = linkTienda(tienda.slug);
	const texto = `¡Hola! Mirá ${tienda.nombre}, ya podés comprar online: ${url}`;
	const tactil = typeof matchMedia === 'function' && matchMedia('(pointer: coarse)').matches;
	if (tactil && typeof navigator.share === 'function') {
		try {
			await navigator.share({ title: tienda.nombre, text: texto });
			return;
		} catch (e) {
			// Si la persona cerró la hoja, no hay nada que avisar.
			if ((e as DOMException)?.name === 'AbortError') return;
		}
	}
	let copiado = true;
	try {
		await navigator.clipboard.writeText(url);
	} catch {
		copiado = false;
	}
	ui.toast({
		k: 'link',
		title: copiado ? 'Link de tu tienda copiado' : 'El link de tu tienda',
		desc: url.replace(/^https?:\/\//, ''),
		action: 'Enviar por WhatsApp',
		onAction: () => {
			window.open(`https://wa.me/?text=${encodeURIComponent(texto)}`, '_blank', 'noopener');
		}
	});
}
