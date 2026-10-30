/**
 * `cloudflare:workers` en el VPS: el mismo `env` que dan los bindings del Worker, armado con piezas
 * locales. En el build para el VPS (DESTINO=vps, vite.config.ts) este módulo reemplaza al de
 * Cloudflare y la app no se entera:
 * - D1 (SANDBOX, ESTADO): SQLite con node:sqlite, un archivo por base, con sus migraciones.
 * - R2 (FILES): una carpeta en disco, con la metadata al lado.
 * - AI: la API de Workers AI por HTTP (Tincho sigue en Cloudflare: es la única excepción).
 * - El resto (texto): variables de entorno.
 */
import { DatabaseSync, type StatementSync } from 'node:sqlite';
import { createHash } from 'node:crypto';
import {
	createReadStream,
	existsSync,
	mkdirSync,
	readdirSync,
	readFileSync,
	rmSync,
	statSync,
	writeFileSync
} from 'node:fs';
import { dirname, join } from 'node:path';
import { Readable } from 'node:stream';

const DATOS = process.env.DATOS_DIR ?? '/data';
const MIGRACIONES = process.env.MIGRACIONES_DIR ?? join(process.cwd(), 'migraciones');

// ---- D1 sobre SQLite ----

type Valor = string | number | bigint | null | Uint8Array;
const valor = (v: unknown): Valor => {
	if (v === undefined || v === null) return null;
	if (typeof v === 'boolean') return v ? 1 : 0;
	if (v instanceof ArrayBuffer) return new Uint8Array(v);
	if (ArrayBuffer.isView(v)) return new Uint8Array(v.buffer, v.byteOffset, v.byteLength);
	if (typeof v === 'object') return JSON.stringify(v);
	return v as Valor;
};

class Sentencia {
	constructor(
		private db: DatabaseSync,
		private sql: string,
		private params: Valor[] = []
	) {}
	bind(...p: unknown[]) {
		return new Sentencia(this.db, this.sql, p.map(valor));
	}
	private preparar(): StatementSync {
		return this.db.prepare(this.sql);
	}
	async first<T = Record<string, unknown>>(col?: string): Promise<T | null> {
		const r = this.preparar().get(...this.params) as Record<string, unknown> | undefined;
		if (!r) return null;
		return (col ? (r[col] as T) : (r as T)) ?? null;
	}
	async all<T = Record<string, unknown>>() {
		const results = this.preparar().all(...this.params) as T[];
		return { results, success: true, meta: { rows_read: results.length, changes: 0 } };
	}
	async raw<T = unknown[]>(o?: { columnNames?: boolean }): Promise<T[]> {
		const s = this.preparar() as StatementSync & {
			setReturnArrays?: (b: boolean) => void;
			columns?: () => { name: string }[];
		};
		// Con setReturnArrays (Node 24) las columnas repetidas de un join no se pisan.
		const arrays = typeof s.setReturnArrays === 'function';
		if (arrays) s.setReturnArrays!(true);
		const filas = s.all(...this.params) as unknown[];
		const arr = (arrays ? filas : filas.map((f) => Object.values(f as object))) as T[];
		if (o?.columnNames) return [(s.columns?.() ?? []).map((c) => c.name) as unknown as T, ...arr];
		return arr;
	}
	async run() {
		const r = this.preparar().run(...this.params);
		return {
			results: [],
			success: true,
			meta: {
				changes: Number(r.changes),
				last_row_id: Number(r.lastInsertRowid),
				rows_written: Number(r.changes)
			}
		};
	}
	/** Para batch: corre y devuelve como all(), con los cambios. */
	correr() {
		const s = this.preparar();
		const results = s.all(...this.params);
		return { results, success: true, meta: { changes: 0 } };
	}
}

class BaseLocal {
	readonly db: DatabaseSync;
	constructor(archivo: string, migraciones: string) {
		mkdirSync(dirname(archivo), { recursive: true });
		this.db = new DatabaseSync(archivo);
		this.db.exec(
			'PRAGMA journal_mode = WAL; PRAGMA busy_timeout = 5000; PRAGMA foreign_keys = ON;'
		);
		migrar(this.db, migraciones);
	}
	prepare(sql: string) {
		return new Sentencia(this.db, sql);
	}
	async batch(sentencias: Sentencia[]) {
		this.db.exec('BEGIN');
		try {
			const out = sentencias.map((s) => s.correr());
			this.db.exec('COMMIT');
			return out;
		} catch (e) {
			this.db.exec('ROLLBACK');
			throw e;
		}
	}
	async exec(sql: string) {
		this.db.exec(sql);
		return { count: sql.split(';').filter((x) => x.trim()).length, duration: 0 };
	}
	withSession() {
		return this;
	}
}

/** Las migraciones de wrangler (carpeta con 0001_x.sql…), anotadas en d1_migrations como hace él. */
function migrar(db: DatabaseSync, carpeta: string) {
	if (!existsSync(carpeta)) return;
	db.exec(
		'CREATE TABLE IF NOT EXISTS d1_migrations (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT UNIQUE, applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP)'
	);
	const hechas = new Set(
		(db.prepare('SELECT name FROM d1_migrations').all() as { name: string }[]).map((r) => r.name)
	);
	for (const f of readdirSync(carpeta)
		.filter((x) => x.endsWith('.sql'))
		.sort()) {
		if (hechas.has(f)) continue;
		db.exec('BEGIN');
		try {
			db.exec(readFileSync(join(carpeta, f), 'utf8'));
			db.prepare('INSERT INTO d1_migrations (name) VALUES (?)').run(f);
			db.exec('COMMIT');
		} catch (e) {
			db.exec('ROLLBACK');
			throw new Error(`migración ${f}: ${e instanceof Error ? e.message : e}`, { cause: e });
		}
	}
}

// ---- R2 en disco ----

type Meta = { contentType?: string; cacheControl?: string; etag: string; size: number };

class ArchivosLocales {
	constructor(private raiz: string) {
		mkdirSync(raiz, { recursive: true });
	}
	private ruta(clave: string) {
		if (clave.includes('..') || clave.startsWith('/')) throw new Error('clave inválida');
		return join(this.raiz, clave);
	}
	async put(
		clave: string,
		datos: ArrayBuffer | ArrayBufferView | string,
		o?: { httpMetadata?: { contentType?: string; cacheControl?: string } }
	) {
		const f = this.ruta(clave);
		mkdirSync(dirname(f), { recursive: true });
		const b =
			typeof datos === 'string'
				? Buffer.from(datos)
				: ArrayBuffer.isView(datos)
					? Buffer.from(datos.buffer, datos.byteOffset, datos.byteLength)
					: Buffer.from(datos);
		writeFileSync(f, b);
		const meta: Meta = {
			...o?.httpMetadata,
			etag: createHash('md5').update(b).digest('hex'),
			size: b.length
		};
		writeFileSync(f + '.meta.json', JSON.stringify(meta));
		return { key: clave, etag: meta.etag, size: meta.size };
	}
	async get(clave: string, o?: { onlyIf?: Headers }) {
		const f = this.ruta(clave);
		if (!existsSync(f)) return null;
		const meta: Meta = existsSync(f + '.meta.json')
			? JSON.parse(readFileSync(f + '.meta.json', 'utf8'))
			: { etag: '', size: statSync(f).size };
		const base = {
			key: clave,
			size: meta.size,
			etag: meta.etag,
			httpEtag: `"${meta.etag}"`,
			writeHttpMetadata(h: Headers) {
				if (meta.contentType) h.set('content-type', meta.contentType);
				if (meta.cacheControl) h.set('cache-control', meta.cacheControl);
			}
		};
		const inm = o?.onlyIf?.get('if-none-match');
		if (inm && inm.replace(/W\//, '') === base.httpEtag) return base;
		return {
			...base,
			body: Readable.toWeb(createReadStream(f)) as ReadableStream,
			arrayBuffer: async () => new Uint8Array(readFileSync(f)).buffer,
			text: async () => readFileSync(f, 'utf8')
		};
	}
	async delete(claves: string | string[]) {
		for (const c of Array.isArray(claves) ? claves : [claves]) {
			const f = this.ruta(c);
			rmSync(f, { force: true });
			rmSync(f + '.meta.json', { force: true });
		}
	}
}

// ---- Workers AI por HTTP ----

const ia = {
	async run(modelo: string, entrada: Record<string, unknown>) {
		const r = await fetch(
			`https://api.cloudflare.com/client/v4/accounts/${process.env.CF_CUENTA}/ai/run/${modelo}`,
			{
				method: 'POST',
				headers: {
					authorization: `Bearer ${process.env.CF_IA_TOKEN}`,
					'content-type': 'application/json'
				},
				body: JSON.stringify(entrada)
			}
		);
		if (!r.ok) throw new Error(`Workers AI ${r.status}`);
		if (entrada.stream) return r.body;
		return ((await r.json()) as { result: unknown }).result;
	}
};

// ---- env ----

let sandbox: BaseLocal | undefined;
let estado: BaseLocal | undefined;
let archivos: ArchivosLocales | undefined;

const BINDINGS: Record<string, () => unknown> = {
	SANDBOX: () =>
		(sandbox ??= new BaseLocal(join(DATOS, 'sandbox.db'), join(MIGRACIONES, 'sandbox'))),
	ESTADO: () => (estado ??= new BaseLocal(join(DATOS, 'estado.db'), join(MIGRACIONES, 'estado'))),
	FILES: () => (archivos ??= new ArchivosLocales(join(DATOS, 'archivos'))),
	AI: () => (process.env.CF_IA_TOKEN ? ia : undefined)
};

export const env = new Proxy({} as Record<string, unknown>, {
	get(_, k: string) {
		return k in BINDINGS ? BINDINGS[k]() : process.env[k];
	},
	has(_, k: string) {
		return k in BINDINGS || k in process.env;
	}
}) as unknown as Cloudflare.Env;
