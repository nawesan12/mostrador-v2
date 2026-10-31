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

describe('la dirección de la tienda en la demo (decisiones §38)', () => {
	it('dice si se puede usar: la propia, inválida, reservada (las vitrinas) o libre', async () => {
		const d = await tienda();
		const { slug } = await d.tienda();
		expect((await d.disponibilidadDireccion(slug)).motivo).toBe('actual');
		expect(await d.disponibilidadDireccion('Mi Local')).toMatchObject({
			disponible: false,
			motivo: 'invalido'
		});
		expect(await d.disponibilidadDireccion('mate-y-co')).toMatchObject({
			disponible: false,
			motivo: 'reservado'
		});
		expect(await d.disponibilidadDireccion('cafe-de-ana')).toMatchObject({
			disponible: true,
			motivo: 'libre'
		});
	});

	it('cambia, la vieja lleva a la nueva y se puede volver', async () => {
		const d = await tienda();
		const vieja = (await d.tienda()).slug;
		expect(await d.cambiarDireccion('cafe-de-ana')).toEqual({
			ok: true,
			slug: 'cafe-de-ana',
			anterior: vieja
		});
		expect((await d.tienda()).slug).toBe('cafe-de-ana');
		expect(await d.slugVigente(vieja)).toBe('cafe-de-ana');
		expect(await d.slugVigente('nunca-fue')).toBeNull();
		// La de antes es una vitrina (reservada), pero fue suya: puede volver.
		expect((await d.disponibilidadDireccion(vieja)).motivo).toBe('propio');
		expect((await d.cambiarDireccion(vieja)).ok).toBe(true);
		expect(await d.slugVigente('cafe-de-ana')).toBe(vieja);
		expect(await d.slugVigente(vieja)).toBeNull();
	});

	it('no deja una inválida ni reservada, y frena a los 5 cambios por día', async () => {
		const d = await tienda();
		expect(await d.cambiarDireccion('-mal')).toMatchObject({ ok: false, motivo: 'invalido' });
		expect(await d.cambiarDireccion('panel')).toMatchObject({ ok: false, motivo: 'reservado' });
		const ahora = Date.now();
		for (const [i, s] of ['uno', 'dos', 'tres', 'cuatro', 'cinco'].entries())
			expect((await d.cambiarDireccion(s, ahora + i)).ok).toBe(true);
		expect(await d.cambiarDireccion('seis', ahora + 10)).toMatchObject({
			ok: false,
			motivo: 'tope'
		});
		// Al otro día, sí.
		expect((await d.cambiarDireccion('seis', ahora + 86_400_001 + 10)).ok).toBe(true);
	});
});
