import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { afterEach, describe, expect, it } from 'vitest';
import caso from '@mostrador/contrato/casos/tienda.json';
import { d1Local } from './d1-local.ts';
import { crear, DatosSandbox, dbSandbox, leer, purgar, VIDA_MS } from './sandbox.ts';
import { cambios } from './schema.ts';
import { slugDe } from './rubros.ts';

const dir = join(import.meta.dirname, '../../../../migraciones-sandbox');
const migraciones = readdirSync(dir)
	.filter((f) => f.endsWith('.sql'))
	.sort()
	.map((f) => readFileSync(join(dir, f), 'utf8'));

let d1: ReturnType<typeof d1Local> | undefined;
const nueva = () => {
	d1 = d1Local(migraciones);
	return dbSandbox(d1);
};
afterEach(() => {
	d1?.cerrar();
	d1 = undefined;
});

const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-8[0-9a-f]{3}-[0-9a-f]{12}$/;

describe('ciclo de vida del sandbox', () => {
	it('se crea con un rubro válido y vence a las 24 h', async () => {
		const db = nueva();
		const ahora = Date.UTC(2026, 9, 20, 12);
		const sb = await crear(db, 'moda', ahora);
		expect(sb?.venceEn).toBe(ahora + VIDA_MS);
		expect(await leer(db, sb!.id, ahora + VIDA_MS - 1)).not.toBeNull();
		expect(await leer(db, sb!.id, ahora + VIDA_MS + 1)).toBeNull();
	});

	it('un rubro inventado no crea nada', async () => {
		const db = nueva();
		expect(await crear(db, 'casino')).toBeNull();
	});

	it('cookies vacías, largas o desconocidas no encuentran nada', async () => {
		const db = nueva();
		expect(await leer(db, undefined)).toBeNull();
		expect(await leer(db, 'x'.repeat(65))).toBeNull();
		expect(await leer(db, 'no-existe')).toBeNull();
	});

	it('la purga borra sólo lo vencido, con sus cambios', async () => {
		const db = nueva();
		const t0 = Date.UTC(2026, 9, 20);
		const viejo = (await crear(db, 'moda', t0))!;
		const nuevo = (await crear(db, 'moda', t0 + VIDA_MS))!;
		await db.insert(cambios).values({
			sandboxId: viejo.id,
			entidad: 'tienda',
			clave: 'tienda',
			datos: '{}',
			actualizadoEn: t0
		});
		expect(await purgar(db, t0 + VIDA_MS + 1)).toBe(1);
		expect(await leer(db, nuevo.id, t0 + VIDA_MS + 1)).not.toBeNull();
		expect(await db.select().from(cambios)).toHaveLength(0);
		expect(await purgar(db, t0 + VIDA_MS + 1)).toBe(0);
	});
});

describe('contrato: tienda', () => {
	it(caso.descripcion, async () => {
		const db = nueva();
		const sb = (await crear(db, caso.siembra.rubro))!;
		const t = await new DatosSandbox(db, sb).tienda();
		expect(t).toMatchObject(caso.esperado);
		expect(t.id).toMatch(UUID);
		expect(Number.isNaN(Date.parse(t.creadaEn))).toBe(false);
		// Lo que el contrato no conoce, no viaja.
		expect(Object.keys(t).sort()).toEqual(
			[...Object.keys(caso.esperado), ...Object.keys(caso.forma)].sort()
		);
	});

	it('el id es estable para el mismo sandbox y distinto entre sandboxes', async () => {
		const db = nueva();
		const a = (await crear(db, 'moda'))!;
		const b = (await crear(db, 'moda'))!;
		const ta = await new DatosSandbox(db, a).tienda();
		expect((await new DatosSandbox(db, a).tienda()).id).toBe(ta.id);
		expect((await new DatosSandbox(db, b).tienda()).id).not.toBe(ta.id);
	});

	it('lo que cambió el visitante pisa al ejemplo, sólo en su sandbox', async () => {
		const db = nueva();
		const a = (await crear(db, 'gastronomia'))!;
		const b = (await crear(db, 'gastronomia'))!;
		await db.insert(cambios).values({
			sandboxId: a.id,
			entidad: 'tienda',
			clave: 'tienda',
			datos: JSON.stringify({ nombre: 'Café La Esquina' }),
			actualizadoEn: Date.now()
		});
		expect((await new DatosSandbox(db, a).tienda()).nombre).toBe('Café La Esquina');
		expect((await new DatosSandbox(db, b).tienda()).nombre).toBe('Café Bruma');
	});
});

describe('slugDe', () => {
	it('saca tildes, símbolos y espacios', () => {
		expect(slugDe('Café Bruma')).toBe('cafe-bruma');
		expect(slugDe('Mate & Co')).toBe('mate-y-co');
		expect(slugDe('  Ópt!ca   Faro ')).toBe('opt-ca-faro');
	});
});
