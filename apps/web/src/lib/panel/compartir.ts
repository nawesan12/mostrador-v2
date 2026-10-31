import { SITE } from '#lib/site.ts';
import type { UiState } from '#lib/ui.svelte.ts';

/** El link público de la tienda. */
export const linkTienda = (slug: string) => `${SITE}/t/${slug}`;

/** El link público de un producto (por su slug; el id también anda y redirige al slug). */
export const linkProducto = (tienda: string, producto: { id: string; slug?: string | null }) =>
	`${linkTienda(tienda)}/producto/${encodeURIComponent(producto.slug || producto.id)}`;

/** «Mate de calabaza · $ 12.000 — https://…»: lo que acompaña al link al compartir un producto. */
export const textoProducto = (nombre: string, precio: number, url: string) =>
	`${nombre} · $ ${Math.round(precio).toLocaleString('es-AR')} — ${url}`;

/** Celular o tablet: ahí la hoja de compartir del sistema es lo natural; en la compu, copiar. */
export const esTactil = () =>
	typeof matchMedia === 'function' && matchMedia('(pointer: coarse)').matches;

/**
 * La imagen para compartir como archivo (la misma que arma el servidor para la vista previa). Si
 * tarda o falla, se comparte sin ella: el link igual muestra la vista previa. La espera es corta
 * porque Safari sólo abre la hoja poco después del toque.
 */
async function archivoDe(imagen: string, nombre: string): Promise<File | null> {
	const corte = new AbortController();
	const reloj = setTimeout(() => corte.abort(), 1200);
	try {
		const r = await fetch(imagen, { signal: corte.signal });
		if (!r.ok) return null;
		return new File([await r.blob()], `${nombre}.jpg`, { type: 'image/jpeg' });
	} catch {
		return null;
	} finally {
		clearTimeout(reloj);
	}
}

/**
 * La hoja de compartir del sistema, con la imagen como archivo si el navegador lo permite.
 * 'listo' si se compartió o la persona la cerró; 'no' si no hay hoja (o no se pudo abrir).
 */
export async function hojaDelSistema(d: {
	titulo: string;
	texto: string;
	url: string;
	imagen?: string;
	archivo?: string;
}): Promise<'listo' | 'no'> {
	if (typeof navigator === 'undefined' || typeof navigator.share !== 'function') return 'no';
	const base = { title: d.titulo, text: d.texto };
	const archivo = d.imagen ? await archivoDe(d.imagen, d.archivo ?? 'producto') : null;
	const conArchivo = archivo ? { ...base, files: [archivo] } : null;
	const intentos = conArchivo && navigator.canShare?.(conArchivo) ? [conArchivo, base] : [base];
	for (const datos of intentos) {
		try {
			await navigator.share(datos);
			return 'listo';
		} catch (e) {
			// Si la persona cerró la hoja, no hay nada que avisar.
			if ((e as DOMException)?.name === 'AbortError') return 'listo';
			// Algunos navegadores no aceptan el archivo (o la espera de la imagen): sin él.
		}
	}
	return 'no';
}

/** Copia al portapapeles; false si el navegador no dejó. */
export async function copiar(texto: string) {
	try {
		await navigator.clipboard.writeText(texto);
		return true;
	} catch {
		return false;
	}
}

const whatsapp = (texto: string) => {
	window.open(`https://wa.me/?text=${encodeURIComponent(texto)}`, '_blank', 'noopener');
};

/**
 * Compartir la tienda: en el celular, la hoja de compartir del sistema (WhatsApp, Instagram, lo que
 * tenga); en la compu, el link queda copiado y el aviso ofrece mandarlo por WhatsApp.
 */
export async function compartirTienda(ui: UiState, tienda: { nombre: string; slug: string }) {
	const url = linkTienda(tienda.slug);
	const texto = `¡Hola! Mirá ${tienda.nombre}, ya podés comprar online: ${url}`;
	if (esTactil() && (await hojaDelSistema({ titulo: tienda.nombre, texto, url })) === 'listo')
		return;
	const copiado = await copiar(url);
	ui.toast({
		k: 'link',
		title: copiado ? 'Link de tu tienda copiado' : 'El link de tu tienda',
		desc: url.replace(/^https?:\/\//, ''),
		action: 'Enviar por WhatsApp',
		onAction: () => whatsapp(texto)
	});
}

export type ProductoParaCompartir = {
	id: string;
	nombre: string;
	precio: number;
	slug?: string | null;
	publicado: boolean;
};

/**
 * Compartir un producto: en el celular, la hoja del sistema con la imagen del producto como archivo
 * (si el navegador lo permite) y «Nombre · $ precio — link»; en la compu, el link queda copiado y el
 * aviso ofrece mandarlo por WhatsApp. Pausado no se comparte: el link no mostraría nada, así que el
 * aviso ofrece publicarlo.
 */
export async function compartirProducto(
	ui: UiState,
	tienda: string,
	p: ProductoParaCompartir,
	publicar?: () => void | Promise<void>
) {
	if (!p.publicado) {
		ui.toast({
			k: 'warn',
			title: 'Primero publicalo',
			desc: 'Está pausado: con el link nadie lo vería en tu tienda.',
			...(publicar ? { action: 'Publicar', onAction: publicar } : {})
		});
		return;
	}
	const url = linkProducto(tienda, p);
	const texto = textoProducto(p.nombre, p.precio, url);
	if (esTactil()) {
		const listo = await hojaDelSistema({
			titulo: p.nombre,
			texto,
			url,
			// Relativa: la arma el mismo servidor (y en local, el local), por el id.
			imagen: `/t/${tienda}/producto/${encodeURIComponent(p.id)}/compartir/0.jpg`,
			archivo: p.slug || 'producto'
		});
		if (listo === 'listo') return;
	}
	const copiado = await copiar(url);
	ui.toast({
		k: 'link',
		title: copiado ? 'Link del producto copiado' : 'El link del producto',
		desc: url.replace(/^https?:\/\//, ''),
		action: 'Enviar por WhatsApp',
		onAction: () => whatsapp(texto)
	});
}
