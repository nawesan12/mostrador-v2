/**
 * Caché de las imágenes al compartir: en memoria (las últimas) y en disco (la carpeta temporal del
 * contenedor; si se borra, se vuelven a dibujar). La clave lleva la versión de lo dibujado, así que
 * nunca hay que invalidar: una tienda editada pide otra clave. Dos pedidos iguales al mismo tiempo
 * (WhatsApp pide la imagen varias veces) dibujan una sola vez.
 */
const EN_MEMORIA = 40;
const EN_DISCO = 2000;

const memoria = new Map<string, Uint8Array>();
const enCurso = new Map<string, Promise<Uint8Array>>();

function recordar(clave: string, b: Uint8Array) {
	memoria.delete(clave);
	memoria.set(clave, b);
	while (memoria.size > EN_MEMORIA) memoria.delete(memoria.keys().next().value!);
}

async function disco() {
	try {
		const [fs, os, path] = await Promise.all([
			import('node:fs/promises'),
			import('node:os'),
			import('node:path')
		]);
		const dir = path.join(os.tmpdir(), 'tulocalweb-compartir');
		await fs.mkdir(dir, { recursive: true });
		return { fs, dir, ruta: (c: string) => path.join(dir, `${c}.jpg`) };
	} catch {
		return null;
	}
}

/** La imagen de `clave` (letras, números y guiones), dibujándola con `hacer` si no está. */
export async function enCache(clave: string, hacer: () => Promise<Uint8Array>) {
	if (!/^[a-z0-9-]{1,120}$/.test(clave)) throw new Error('clave inválida');
	const ya = memoria.get(clave);
	if (ya) {
		recordar(clave, ya);
		return ya;
	}
	const corriendo = enCurso.get(clave);
	if (corriendo) return corriendo;
	const p = (async () => {
		const d = await disco();
		if (d) {
			const b = await d.fs.readFile(d.ruta(clave)).catch(() => null);
			if (b) return new Uint8Array(b);
		}
		const b = await hacer();
		if (d) {
			await d.fs.writeFile(d.ruta(clave), b).catch(() => {});
			void podar(d);
		}
		return b;
	})();
	enCurso.set(clave, p);
	try {
		const b = await p;
		recordar(clave, b);
		return b;
	} finally {
		enCurso.delete(clave);
	}
}

/** Con demasiados archivos, se borran los más viejos (de a poco: como mucho una vez por minuto). */
let ultimaPoda = 0;
async function podar(d: NonNullable<Awaited<ReturnType<typeof disco>>>) {
	if (Date.now() - ultimaPoda < 60_000) return;
	ultimaPoda = Date.now();
	try {
		const nombres = await d.fs.readdir(d.dir);
		if (nombres.length <= EN_DISCO) return;
		const con = await Promise.all(
			nombres.map(async (n) => ({
				n,
				t: (await d.fs.stat(`${d.dir}/${n}`).catch(() => null))?.mtimeMs ?? 0
			}))
		);
		con.sort((a, b) => a.t - b.t);
		await Promise.all(
			con
				.slice(0, con.length - EN_DISCO / 2)
				.map((x) => d.fs.rm(`${d.dir}/${x.n}`, { force: true }))
		);
	} catch {
		// Si no se puede podar, se intenta la próxima.
	}
}
