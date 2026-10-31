import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { afterEach, describe, expect, it } from 'vitest';
import { d1Local } from './d1-local.ts';
import { crear, DatosSandbox, dbSandbox } from './sandbox.ts';

const dir = join(import.meta.dirname, '../../../../migraciones-sandbox');
const migraciones = readdirSync(dir)
	.filter((f) => f.endsWith('.sql'))
	.sort()
	.map((f) => readFileSync(join(dir, f), 'utf8'));

let d1: ReturnType<typeof d1Local> | undefined;
afterEach(() => {
	d1?.cerrar();
	d1 = undefined;
});

async function tienda() {
	d1 = d1Local(migraciones);
	const db = dbSandbox(d1);
	const sb = await crear(db, 'gastronomia', Date.now());
	return new DatosSandbox(db, sb!);
}

describe('la campanita en la demo (decisiones §45)', () => {
	it('lista las suyas con el número sin leer, tocar una la marca y «todo» las marca todas', async () => {
		const d = await tienda();
		const p = await d.notificaciones();
		expect(p.siguiente).toBeNull();
		expect(p.notificaciones.length).toBeGreaterThan(0);
		expect(p.sinLeer).toBe(p.notificaciones.filter((n) => !n.leida).length);
		const una = p.notificaciones.find((n) => !n.leida)!;
		await d.marcarNotificacionLeida(una.id);
		await d.marcarNotificacionLeida(una.id);
		const tras = await d.notificaciones();
		expect(tras.sinLeer).toBe(p.sinLeer - 1);
		expect(tras.notificaciones.find((n) => n.id === una.id)?.leida).toBe(true);
		await d.marcarNotificacionesLeidas();
		expect((await d.notificaciones()).sinLeer).toBe(0);
	});
});
