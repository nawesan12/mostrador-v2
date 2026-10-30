/**
 * Subir fotos de productos: se achican en el navegador (lado mayor 1600 px) para que suban rápido
 * con datos del celular, y el servidor las deja todas en WebP de 1600 px (lib/server/archivos.ts).
 */
const LADO = 1600;
const TOPE = 8 * 1024 * 1024;
/** Lo que acepta el servidor sin achicar (una foto grande de iPhone). */
const MAXIMO = 15 * 1024 * 1024;

/**
 * La foto lista para subir: WebP de hasta 1600 px y menos de 1 MB (baja la calidad si hace falta).
 * Safari no sabe codificar WebP: devuelve un PNG sin comprimir que casi nunca entra en 1 MB y la foto
 * se rechazaba en el iPhone. Ahí va el PNG si entra (un logo conserva su transparencia) y si no, JPEG.
 */
export async function comprimir(archivo: Blob): Promise<Blob> {
	const bmp = await createImageBitmap(archivo);
	let lado = LADO;
	let tipo = 'image/webp';
	for (const calidad of [0.82, 0.72, 0.6]) {
		const escala = Math.min(1, lado / Math.max(bmp.width, bmp.height));
		const w = Math.round(bmp.width * escala);
		const h = Math.round(bmp.height * escala);
		const lienzo = new OffscreenCanvas(w, h);
		lienzo.getContext('2d')!.drawImage(bmp, 0, 0, w, h);
		let b = await lienzo.convertToBlob({ type: tipo, quality: calidad });
		if (b.type !== tipo) {
			if (b.type === 'image/png' && b.size <= TOPE) return b;
			tipo = 'image/jpeg';
			b = await lienzo.convertToBlob({ type: tipo, quality: calidad });
		}
		if (b.size <= TOPE) return b;
		lado = Math.round(lado * 0.8);
	}
	throw new Error('La foto es demasiado pesada. Probá con otra.');
}

/** Achica si se puede; si el navegador no puede leerla, va como está y la optimiza el servidor. */
async function paraSubir(archivo: Blob): Promise<Blob> {
	try {
		return await comprimir(archivo);
	} catch {
		if (archivo.size <= MAXIMO) return archivo;
		throw new Error('La foto pesa más de 15 MB. Probá con otra.');
	}
}

/** Sube una foto y devuelve su URL. Los errores vienen con el mensaje para mostrar. */
export async function subir(archivo: Blob): Promise<string> {
	const r = await fetch('/panel/fotos', {
		method: 'POST',
		headers: { 'content-type': 'application/octet-stream' },
		body: await paraSubir(archivo)
	});
	if (!r.ok)
		throw new Error(
			((await r.json().catch(() => null)) as { message?: string } | null)?.message ??
				'No pudimos subir la foto.'
		);
	return new URL(((await r.json()) as { ruta: string }).ruta, location.origin).href;
}
