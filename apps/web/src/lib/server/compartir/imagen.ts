/**
 * La imagen al compartir: satori arma un SVG con el dibujo (las letras como trazos, así el
 * contenedor no necesita fuentes instaladas) y vips, el mismo que achica las fotos que se suben,
 * lo pasa a JPEG. Las fotos (portada, productos, logo) se bajan y se achican con vips antes de entrar
 * al SVG; una que no se puede leer se saltea (nunca rompe la imagen).
 */
import satori from 'satori';
import { unsplash } from '#lib/landing/formato.ts';
import type { Tarjeta } from '#lib/tienda/compartir.ts';
import { ALTO, ANCHO, dibujar, sinViudas, type Foto, type Medir } from './dibujo.ts';
import { estiloDe, fuentesDe, type Tipografia } from './estilos.ts';

/** Topes: lo que se baja por foto, lo que espera, y el peso máximo del JPEG (WhatsApp: 300 KB). */
const TOPES = { bajada: 12 * 1024 * 1024, espera: 6000, jpeg: 290 * 1024 };

export type Fuentes = (archivo: string) => Promise<ArrayBuffer>;
/** Los bytes de una foto de la tienda (null si no se puede). */
export type Bytes = (src: string) => Promise<Uint8Array | null>;

async function nodo() {
	const [{ execFile }, { promisify }, fs, os, path] = await Promise.all([
		import('node:child_process'),
		import('node:util'),
		import('node:fs/promises'),
		import('node:os'),
		import('node:path')
	]);
	return { exec: promisify(execFile), fs, os, path };
}

/** vips dentro de una carpeta temporal que se borra al terminar. */
async function conVips<T>(
	fn: (vips: (args: string[]) => Promise<void>, dir: string) => Promise<T>
) {
	const { exec, fs, os, path } = await nodo();
	const dir = await fs.mkdtemp(path.join(os.tmpdir(), 'compartir-'));
	try {
		return await fn(async (args) => {
			await exec('vips', args, { timeout: 20_000 });
		}, dir);
	} finally {
		await fs.rm(dir, { recursive: true, force: true });
	}
}

/** Ancho y alto de un JPEG o PNG, leyendo su cabecera. */
export function medidas(b: Uint8Array): { ancho: number; alto: number } | null {
	if (b[0] === 0x89 && b[1] === 0x50) {
		const v = new DataView(b.buffer, b.byteOffset, b.byteLength);
		return { ancho: v.getUint32(16), alto: v.getUint32(20) };
	}
	if (b[0] === 0xff && b[1] === 0xd8) {
		let i = 2;
		while (i + 9 < b.length) {
			if (b[i] !== 0xff) return null;
			const m = b[i + 1];
			const largo = (b[i + 2] << 8) | b[i + 3];
			if (m >= 0xc0 && m <= 0xcf && m !== 0xc4 && m !== 0xc8 && m !== 0xcc)
				return { alto: (b[i + 5] << 8) | b[i + 6], ancho: (b[i + 7] << 8) | b[i + 8] };
			i += 2 + largo;
		}
	}
	return null;
}

const base64 = (b: Uint8Array) => Buffer.from(b).toString('base64');

/**
 * Una foto lista para el dibujo: JPEG de hasta `lado` px (sin agrandar), o PNG si es el logo (para
 * no perder la transparencia).
 */
async function preparar(bytes: Uint8Array, lado: number, logo: boolean): Promise<Foto | null> {
	try {
		return await conVips(async (vips, dir) => {
			const { fs, path } = await nodo();
			const entra = path.join(dir, 'entra');
			const sale = path.join(dir, logo ? 'sale.png' : 'sale.jpg');
			await fs.writeFile(entra, bytes);
			await vips([
				'thumbnail',
				entra,
				logo ? `${sale}[keep=none]` : `${sale}[Q=84,keep=none]`,
				String(lado),
				'--size',
				'down'
			]);
			const out = new Uint8Array(await fs.readFile(sale));
			const m = medidas(out);
			if (!m) return null;
			return { uri: `data:image/${logo ? 'png' : 'jpeg'};base64,${base64(out)}`, ...m };
		});
	} catch {
		return null;
	}
}

/** Host de internet (no la máquina misma ni la red interna). */
function publica(u: URL) {
	const h = u.hostname;
	return (
		u.protocol === 'https:' &&
		!/^(localhost|.*\.local|.*\.internal|\d+\.\d+\.\d+\.\d+|\[.*\])$/i.test(h) &&
		h.includes('.')
	);
}

/** Baja una foto de internet con tope de tiempo y de peso. */
export async function bajar(src: string): Promise<Uint8Array | null> {
	let u: URL;
	try {
		u = new URL(/^https?:\/\//.test(src) ? src : unsplash(src, 1400));
	} catch {
		return null;
	}
	if (!publica(u)) return null;
	try {
		const r = await fetch(u, { signal: AbortSignal.timeout(TOPES.espera), redirect: 'follow' });
		if (!r.ok || !r.body) return null;
		if (Number(r.headers.get('content-length') ?? 0) > TOPES.bajada) return null;
		const b = new Uint8Array(await r.arrayBuffer());
		return b.byteLength > TOPES.bajada ? null : b;
	} catch {
		return null;
	}
}

/** Las fuentes de la plantilla, para satori. */
async function fuentes(cargar: Fuentes, estilo: ReturnType<typeof estiloDe>) {
	return Promise.all(
		fuentesDe(estilo).map(async (f) => ({
			name: f.familia,
			data: await cargar(f.archivo),
			weight: f.peso as 400 | 500 | 600 | 700 | 800 | 900,
			style: (f.italica ? 'italic' : 'normal') as 'italic' | 'normal'
		}))
	);
}

const claveTipo = (t: Tipografia) =>
	`${t.familia}|${t.peso}|${t.italica}|${t.mayusculas}|${t.tracking}`;

/**
 * Mide con las fuentes de verdad cada palabra de los textos grandes (en una sola pasada de satori:
 * cada palabra es una caja con un color distinto y se lee su ancho en el SVG). Así el tamaño de
 * letra se elige para que el nombre entre justo, sin desbordar ni quedar chico.
 */
export async function medidor(
	textos: { texto: string; tipo: Tipografia }[],
	fonts: Awaited<ReturnType<typeof fuentes>>
): Promise<Medir> {
	const medidas = new Map<string, number>();
	const items: { clave: string; texto: string; tipo: Tipografia }[] = [];
	for (const { texto, tipo } of textos) {
		if (!texto) continue;
		for (const p of [...sinViudas(texto).split(' '), 'x\u00a0x', 'xx'])
			items.push({ clave: `${claveTipo(tipo)}#${p}`, texto: p, tipo });
	}
	const unicos = [...new Map(items.map((i) => [i.clave, i])).values()];
	if (!unicos.length) return () => null;
	const color = (i: number) => `#${(i + 1).toString(16).padStart(6, '0')}`;
	const svg = await satori(
		{
			type: 'div',
			props: {
				style: { display: 'flex', flexDirection: 'column', alignItems: 'flex-start' },
				children: unicos.map((u, i) => ({
					type: 'div',
					props: {
						style: {
							display: 'flex',
							background: color(i),
							whiteSpace: 'pre',
							fontFamily: u.tipo.familia,
							fontWeight: u.tipo.peso,
							fontStyle: u.tipo.italica ? 'italic' : 'normal',
							textTransform: u.tipo.mayusculas ? 'uppercase' : 'none',
							letterSpacing: `${u.tipo.tracking}em`,
							fontSize: 100,
							lineHeight: 1
						},
						children: u.texto
					}
				}))
			}
		} as Parameters<typeof satori>[0],
		{ width: 8000, height: unicos.length * 110 + 10, fonts }
	);
	for (const m of svg.matchAll(/<rect [^>]*?width="([\d.]+)"[^>]*?fill="(#[0-9a-f]{6})"/g)) {
		const i = parseInt(m[2].slice(1), 16) - 1;
		if (unicos[i]) medidas.set(unicos[i].clave, Number(m[1]) / 100);
	}
	return (texto, tipo) => {
		const k = claveTipo(tipo);
		if (texto === ' ') {
			const a = medidas.get(`${k}#x\u00a0x`);
			const b = medidas.get(`${k}#xx`);
			return a != null && b != null ? a - b : null;
		}
		return medidas.get(`${k}#${texto}`) ?? null;
	};
}

/** El SVG de la tarjeta (sin pasar a JPEG): sirve también para las pruebas. */
export async function svgDe(tarjeta: Tarjeta, o: { fuentes: Fuentes; bytes: Bytes }) {
	const estilo = estiloDe(tarjeta.plantilla, tarjeta.paleta);
	const fuentes_ = fuentes(o.fuentes, estilo);
	const fuentesFotos =
		tarjeta.tipo === 'producto'
			? [tarjeta.foto]
			: tarjeta.propia
				? [tarjeta.propia]
				: tarjeta.fotos;
	const lado = tarjeta.tipo === 'tienda' && tarjeta.propia ? 1400 : 900;
	const [fotos, logo] = await Promise.all([
		Promise.all(
			fuentesFotos.filter(Boolean).map(async (src) => {
				const b = await o.bytes(src);
				return b ? preparar(b, lado, false) : null;
			})
		),
		tarjeta.logo
			? o.bytes(tarjeta.logo).then((b) => (b ? preparar(b, 640, true) : null))
			: Promise.resolve(null)
	]);
	const fonts = await fuentes_;
	const medir = await medidor(
		[
			{ texto: tarjeta.nombre, tipo: estilo.titulo },
			{ texto: tarjeta.tipo === 'tienda' ? tarjeta.frase : '', tipo: estilo.frase }
		],
		fonts
	).catch(() => undefined);
	const arbol = dibujar(tarjeta, estilo, {
		fotos: fotos.filter((f): f is Foto => !!f),
		logo,
		medir
	});
	return satori(arbol as Parameters<typeof satori>[0], { width: ANCHO, height: ALTO, fonts });
}

/** La imagen en JPEG, de menos de 300 KB. */
export async function renderizar(tarjeta: Tarjeta, o: { fuentes: Fuentes; bytes: Bytes }) {
	const svg = await svgDe(tarjeta, o);
	return conVips(async (vips, dir) => {
		const { fs, path } = await nodo();
		const entra = path.join(dir, 'tarjeta.svg');
		await fs.writeFile(entra, svg);
		for (const q of [86, 76, 64]) {
			const sale = path.join(dir, `tarjeta-${q}.jpg`);
			await vips(['copy', entra, `${sale}[Q=${q},keep=none,optimize-coding,interlace]`]);
			const out = new Uint8Array(await fs.readFile(sale));
			if (out.byteLength <= TOPES.jpeg || q === 64) return out;
		}
		throw new Error('sin imagen');
	});
}
