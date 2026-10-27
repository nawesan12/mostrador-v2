/**
 * Subir fotos de productos: se achican y pasan a WebP en el navegador (lado mayor 1600 px) antes de
 * mandarlas, así pesan poco y entran en los topes de R2 (lib/server/archivos.ts: 1 MB por foto).
 */
const LADO = 1600;
const TOPE = 1024 * 1024;

/** La foto lista para subir: WebP de hasta 1600 px y menos de 1 MB (baja la calidad si hace falta). */
export async function comprimir(archivo: Blob): Promise<Blob> {
	const bmp = await createImageBitmap(archivo);
	let lado = LADO;
	for (const calidad of [0.82, 0.72, 0.6]) {
		const escala = Math.min(1, lado / Math.max(bmp.width, bmp.height));
		const w = Math.round(bmp.width * escala);
		const h = Math.round(bmp.height * escala);
		const lienzo = new OffscreenCanvas(w, h);
		lienzo.getContext('2d')!.drawImage(bmp, 0, 0, w, h);
		const b = await lienzo.convertToBlob({ type: 'image/webp', quality: calidad });
		if (b.size <= TOPE) return b;
		lado = Math.round(lado * 0.8);
	}
	throw new Error('La foto es demasiado pesada. Probá con otra.');
}

/** Sube una foto y devuelve su URL. Los errores vienen con el mensaje para mostrar. */
export async function subir(archivo: Blob): Promise<string> {
	const r = await fetch('/panel/fotos', {
		method: 'POST',
		headers: { 'content-type': 'application/octet-stream' },
		body: await comprimir(archivo)
	});
	if (!r.ok)
		throw new Error(
			((await r.json().catch(() => null)) as { message?: string } | null)?.message ??
				'No pudimos subir la foto.'
		);
	return new URL(((await r.json()) as { ruta: string }).ruta, location.origin).href;
}
