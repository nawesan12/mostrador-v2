/**
 * Fotos de productos en R2, como `storage.ts` de Caja: topes duros porque la cuenta es de pago por
 * uso y el plan es quedar en el gratis (10 GB de R2 para toda la cuenta, compartidos con Caja). La
 * foto llega ya comprimida a WebP desde el navegador; acá se valida el formato por sus bytes.
 */
import { env } from 'cloudflare:workers';
import { error } from '@sveltejs/kit';
import { sql } from 'drizzle-orm';
import { inArray } from './sandbox/sql.ts';
import type { DbSandbox } from './sandbox/sandbox.ts';
import { archivos } from './sandbox/schema.ts';

export const LIMITES = {
	/** Por foto (llega a ~1600 px en WebP, que pesa bastante menos). */
	archivo: 1024 * 1024,
	/** Por dueño: la tienda (o el visitante de la demo). */
	porDuenio: 15 * 1024 * 1024,
	/** Mostrador entero. */
	total: 1024 * 1024 * 1024
};

/** WebP, JPEG o PNG, mirando la firma (no el tipo que dice el navegador). */
export function tipoDe(b: Uint8Array): string | null {
	const ascii = (i: number, n: number) => String.fromCharCode(...b.subarray(i, i + n));
	if (b.length > 12 && ascii(0, 4) === 'RIFF' && ascii(8, 4) === 'WEBP') return 'image/webp';
	if (b.length > 3 && b[0] === 0xff && b[1] === 0xd8 && b[2] === 0xff) return 'image/jpeg';
	if (b.length > 8 && b[0] === 0x89 && ascii(1, 3) === 'PNG') return 'image/png';
	return null;
}

const EXT: Record<string, string> = {
	'image/webp': 'webp',
	'image/jpeg': 'jpg',
	'image/png': 'png'
};

/** La clave pública de un archivo: `fotos/{dueño}/{uuid}.{ext}`. */
export const CLAVE = /^fotos\/[0-9a-f-]{36}\/[0-9a-f-]{36}\.(webp|jpg|png)$/;

/**
 * Sube una foto respetando los topes. Devuelve su clave. `duenio` es el id de la tienda (va en la
 * URL); `sandbox`, el del visitante en la demo (no se publica: sólo sirve para la purga).
 */
export async function subirFoto(
	db: DbSandbox,
	duenio: string,
	bytes: Uint8Array,
	sandbox: string | null = null
) {
	const tipo = tipoDe(bytes);
	if (!tipo) error(415, 'Subí una foto en JPG, PNG o WebP.');
	if (bytes.byteLength > LIMITES.archivo) error(413, 'La foto pesa más de 1 MB.');
	const [uso] = await db
		.select({
			duenio: sql<number>`coalesce(sum(case when ${archivos.duenio} = ${duenio} then ${archivos.bytes} else 0 end), 0)`,
			total: sql<number>`coalesce(sum(${archivos.bytes}), 0)`
		})
		.from(archivos);
	if (Number(uso?.duenio ?? 0) + bytes.byteLength > LIMITES.porDuenio)
		error(413, 'Llegaste al límite de 15 MB de fotos.');
	if (Number(uso?.total ?? 0) + bytes.byteLength > LIMITES.total)
		error(507, 'Subir fotos está pausado por un rato. Probá más tarde.');
	const clave = `fotos/${duenio}/${crypto.randomUUID()}.${EXT[tipo]}`;
	await env.FILES.put(clave, bytes, {
		httpMetadata: { contentType: tipo, cacheControl: 'public, max-age=31536000, immutable' }
	});
	await db
		.insert(archivos)
		.values({ clave, duenio, sandboxId: sandbox, bytes: bytes.byteLength, creadoEn: Date.now() });
	return clave;
}

/** Borra las fotos de los sandboxes vencidos de la demo. */
export async function borrarDe(db: DbSandbox, sandboxes: string[]) {
	if (!sandboxes.length) return 0;
	const filas = await db
		.select({ clave: archivos.clave })
		.from(archivos)
		.where(inArray(archivos.sandboxId, sandboxes));
	if (!filas.length) return 0;
	// R2 borra hasta 1000 claves por llamada.
	await env.FILES.delete(filas.map((f) => f.clave));
	await db.delete(archivos).where(inArray(archivos.sandboxId, sandboxes));
	return filas.length;
}
