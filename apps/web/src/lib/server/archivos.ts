/**
 * Fotos de productos: en producción, una carpeta en el disco del VPS (plataforma/vps.ts reemplaza al
 * binding FILES de Cloudflare); en wrangler dev, R2 local. Topes duros: el disco es uno solo. Toda foto que llega (JPG, PNG o WebP, aunque sea una grande
 * de iPhone) se pasa a WebP de 1600 px como máximo, calidad 80, girada según la cámara y sin
 * metadatos, con vips en el VPS: todas quedan iguales y livianas. Sin vips (pruebas locales), se
 * guarda como llegó.
 */
import { env } from 'cloudflare:workers';
import { error } from '@sveltejs/kit';
import { sql } from 'drizzle-orm';
import { inArray } from './sandbox/sql.ts';
import type { DbSandbox } from './sandbox/sandbox.ts';
import { archivos } from './sandbox/schema.ts';

export const LIMITES = {
	/** Lo que se acepta subir (una foto grande de iPhone); se guarda ya optimizada. */
	subida: 15 * 1024 * 1024,
	/** Por foto, ya optimizada (1600 px en WebP pesa bastante menos). */
	archivo: 1024 * 1024,
	/**
	 * Por dueño, si no se pasa otro: el visitante de la demo. Cada tienda tiene el de su plan
	 * (lib/panel/planes.ts › fotosMB: 15, 150, 600 o 2.048 MB).
	 */
	porDuenio: 15 * 1024 * 1024,
	/**
	 * Tu Local Web entero (el disco del VPS; un freno, no un tope de plan). Antes de subirlo, confirmar
	 * el espacio libre del disco y que /data/archivos entra en la copia diaria.
	 */
	total: 20 * 1024 * 1024 * 1024,
	/**
	 * Lo último del total queda para las tiendas que pagan: pasado este punto, Gratis y la prueba no
	 * suben más (así unas cuantas cuentas de prueba llenas no frenan a las que pagan).
	 */
	reservaPagos: 0.7,
	/** Desde acá se avisa en el log del servidor, una vez por proceso, para agrandar a tiempo. */
	aviso: 0.8
};

let avisado = false;

/** Archivos para descargar (el premio «descarga» de una captación): PDF, EPUB o ZIP, sin tocar. */
export const LIMITES_DOCUMENTO = {
	/** Por archivo. */
	archivo: 20 * 1024 * 1024,
	/** Por dueño: alcanza para varias guías y catálogos. */
	porDuenio: 60 * 1024 * 1024,
	/** Tu Local Web entero (aparte de las fotos). */
	total: 3 * 1024 * 1024 * 1024
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

/**
 * La foto, en WebP de hasta 1600 × 1600 (sin agrandar), calidad 80, girada según la cámara y sin
 * metadatos (ubicación incluida). Si vips no está o falla, vuelve la foto como llegó.
 */
export async function optimizar(bytes: Uint8Array): Promise<Uint8Array> {
	try {
		const [{ execFile }, { promisify }, fs, os, path] = await Promise.all([
			import('node:child_process'),
			import('node:util'),
			import('node:fs/promises'),
			import('node:os'),
			import('node:path')
		]);
		const dir = await fs.mkdtemp(path.join(os.tmpdir(), 'foto-'));
		try {
			const entra = path.join(dir, 'entra');
			const sale = path.join(dir, 'sale.webp');
			await fs.writeFile(entra, bytes);
			await promisify(execFile)(
				'vips',
				['thumbnail', entra, `${sale}[Q=80,keep=none]`, '1600', '--size', 'down'],
				{ timeout: 20_000 }
			);
			const out = new Uint8Array(await fs.readFile(sale));
			return tipoDe(out) === 'image/webp' ? out : bytes;
		} finally {
			await fs.rm(dir, { recursive: true, force: true });
		}
	} catch {
		return bytes;
	}
}

/**
 * PDF, EPUB o ZIP, mirando la firma. El EPUB es un ZIP cuyo primer archivo es `mimetype` (sin
 * comprimir) con `application/epub+zip`; cualquier otro ZIP es un ZIP.
 */
export function tipoDocumentoDe(b: Uint8Array): string | null {
	const ascii = (i: number, n: number) => String.fromCharCode(...b.subarray(i, i + n));
	if (b.length > 5 && ascii(0, 5) === '%PDF-') return 'application/pdf';
	if (b.length > 4 && b[0] === 0x50 && b[1] === 0x4b && b[2] === 0x03 && b[3] === 0x04) {
		if (ascii(30, 8) === 'mimetype' && ascii(38, 20) === 'application/epub+zip')
			return 'application/epub+zip';
		return 'application/zip';
	}
	return null;
}

const EXT_DOCUMENTO: Record<string, string> = {
	'application/pdf': 'pdf',
	'application/epub+zip': 'epub',
	'application/zip': 'zip'
};

/** La clave pública de un archivo para descargar: `archivos/{dueño}/{uuid}.{ext}`. */
export const CLAVE_DOCUMENTO = /^archivos\/[0-9a-f-]{36}\/[0-9a-f-]{36}\.(pdf|epub|zip)$/;

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
	sandbox: string | null = null,
	/** El espacio para fotos del plan de la tienda, en bytes. */
	porDuenio: number = LIMITES.porDuenio,
	/** La tienda paga un plan (usa la reserva del total). */
	pago = false
) {
	if (!tipoDe(bytes)) error(415, 'Subí una foto en JPG, PNG o WebP.');
	if (bytes.byteLength > LIMITES.subida) error(413, 'La foto pesa más de 15 MB.');
	bytes = await optimizar(bytes);
	const tipo = tipoDe(bytes)!;
	if (bytes.byteLength > LIMITES.archivo) error(413, 'La foto pesa más de 1 MB.');
	const [uso] = await db
		.select({
			duenio: sql<number>`coalesce(sum(case when ${archivos.duenio} = ${duenio} then ${archivos.bytes} else 0 end), 0)`,
			total: sql<number>`coalesce(sum(${archivos.bytes}), 0)`
		})
		.from(archivos)
		.where(sql`${archivos.clave} like 'fotos/%'`);
	if (Number(uso?.duenio ?? 0) + bytes.byteLength > porDuenio)
		error(
			413,
			`Llegaste a los ${Math.round(porDuenio / 1024 / 1024).toLocaleString('es-AR')} MB de fotos de tu plan. Borrá fotos que ya no uses o pasá a un plan más grande.`
		);
	const total = Number(uso?.total ?? 0) + bytes.byteLength;
	if (total > LIMITES.total * LIMITES.aviso && !avisado) {
		avisado = true;
		console.error(
			`archivos: las fotos ocupan ${Math.round((total / LIMITES.total) * 100)} % del total (${Math.round(total / 1024 / 1024)} MB): agrandar LIMITES.total o el disco`
		);
	}
	if (total > LIMITES.total || (!pago && total > LIMITES.total * LIMITES.reservaPagos))
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

/** Lo que ocupan las fotos de un dueño (la tienda), en bytes: lo que cuenta para el espacio del plan. */
export async function usoDeFotos(db: DbSandbox, duenio: string): Promise<number> {
	const [uso] = await db
		.select({ bytes: sql<number>`coalesce(sum(${archivos.bytes}), 0)` })
		.from(archivos)
		.where(sql`${archivos.clave} like 'fotos/%' and ${archivos.duenio} = ${duenio}`);
	return Number(uso?.bytes ?? 0);
}

/**
 * Sube un archivo para descargar (PDF, EPUB o ZIP) tal como llegó, respetando los topes. Devuelve
 * su clave. Los mismos dueños que las fotos; la cuenta de espacio va aparte.
 */
export async function subirDocumento(
	db: DbSandbox,
	duenio: string,
	bytes: Uint8Array,
	sandbox: string | null = null,
	/** Dónde se guarda (las pruebas pasan uno en memoria). */
	destino: Pick<R2Bucket, 'put'> = env.FILES
) {
	const tipo = tipoDocumentoDe(bytes);
	if (!tipo) error(415, 'Subí un archivo PDF, EPUB o ZIP.');
	if (bytes.byteLength > LIMITES_DOCUMENTO.archivo) error(413, 'El archivo pesa más de 20 MB.');
	const [uso] = await db
		.select({
			duenio: sql<number>`coalesce(sum(case when ${archivos.duenio} = ${duenio} then ${archivos.bytes} else 0 end), 0)`,
			total: sql<number>`coalesce(sum(${archivos.bytes}), 0)`
		})
		.from(archivos)
		.where(sql`${archivos.clave} like 'archivos/%'`);
	if (Number(uso?.duenio ?? 0) + bytes.byteLength > LIMITES_DOCUMENTO.porDuenio)
		error(413, 'Llegaste al límite de 60 MB de archivos. Borrá alguno que ya no uses.');
	if (Number(uso?.total ?? 0) + bytes.byteLength > LIMITES_DOCUMENTO.total)
		error(507, 'Subir archivos está pausado por un rato. Probá más tarde.');
	const clave = `archivos/${duenio}/${crypto.randomUUID()}.${EXT_DOCUMENTO[tipo]}`;
	await destino.put(clave, bytes, {
		httpMetadata: { contentType: tipo, cacheControl: 'public, max-age=31536000, immutable' }
	});
	await db
		.insert(archivos)
		.values({ clave, duenio, sandboxId: sandbox, bytes: bytes.byteLength, creadoEn: Date.now() });
	return clave;
}

/** Borra las fotos y los archivos de los sandboxes vencidos de la demo. */
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
