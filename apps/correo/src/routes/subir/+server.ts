import { error, json } from '@sveltejs/kit';
import { Jmap } from '#lib/server/jmap.ts';
import type { RequestHandler } from './$types';

const MAX = 25 * 1024 * 1024;

/** Sube un adjunto para un correo nuevo, a la cuenta de la casilla desde la que se manda. */
export const POST: RequestHandler = async ({ locals, request, url }) => {
	const cuenta = url.searchParams.get('cuenta') ?? '';
	const f = await request.formData();
	const archivo = f.get('archivo');
	if (!(archivo instanceof File) || !cuenta) error(400, 'Falta el archivo.');
	if (archivo.size > MAX) error(413, 'El archivo pasa los 25 MB.');
	const r = await new Jmap(locals.cred!).subir(cuenta, archivo);
	return json({
		blobId: r.blobId,
		nombre: archivo.name,
		tipo: archivo.type || 'application/octet-stream',
		tamano: archivo.size
	});
};
