/**
 * D1 sobre node:sqlite, para tests: implementa lo que usan drizzle-orm/d1 y nuestro código
 * (prepare/bind/all/first/run/raw y batch). El SQL corre de verdad en SQLite; sólo cambia el
 * transporte. Activa las foreign keys como D1.
 */
import { DatabaseSync } from 'node:sqlite';

type Valor = string | number | bigint | null | Uint8Array;

class Sentencia {
	constructor(
		private db: DatabaseSync,
		private sql: string,
		private params: Valor[] = []
	) {}
	bind(...p: unknown[]) {
		return new Sentencia(
			this.db,
			this.sql,
			p.map((v) => (typeof v === 'boolean' ? Number(v) : (v ?? null)) as Valor)
		);
	}
	private st() {
		return this.db.prepare(this.sql);
	}
	async all<T>() {
		return { results: this.st().all(...this.params) as T[], success: true, meta: {} };
	}
	async first<T>(col?: string) {
		const fila = this.st().get(...this.params) as Record<string, unknown> | undefined;
		if (!fila) return null;
		return (col ? fila[col] : fila) as T;
	}
	async run() {
		const r = this.st().run(...this.params);
		return { success: true, results: [], meta: { changes: Number(r.changes) } };
	}
	async raw<T>() {
		const st = this.st();
		st.setReturnArrays(true);
		return st.all(...this.params) as T[];
	}
	/** Para batch: corre y devuelve lo mismo que all(). */
	ejecutar() {
		const st = this.st();
		return st.columns().length
			? { results: st.all(...this.params), success: true, meta: {} }
			: { results: [], success: true, meta: { changes: Number(st.run(...this.params).changes) } };
	}
}

export function d1Local(migraciones: string[] = []) {
	const db = new DatabaseSync(':memory:');
	db.exec('PRAGMA foreign_keys = ON');
	for (const m of migraciones) db.exec(m);
	return {
		prepare: (sql: string) => new Sentencia(db, sql),
		async batch(sts: Sentencia[]) {
			db.exec('BEGIN');
			try {
				const res = sts.map((s) => s.ejecutar());
				db.exec('COMMIT');
				return res;
			} catch (e) {
				db.exec('ROLLBACK');
				throw e;
			}
		},
		async exec(sql: string) {
			db.exec(sql);
			return { count: 0, duration: 0 };
		},
		cerrar: () => db.close()
	} as unknown as D1Database & { cerrar(): void };
}
