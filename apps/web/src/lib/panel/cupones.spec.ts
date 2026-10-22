import { describe, expect, it } from 'vitest';
import { cuando, estimar, grande, lunes9, vigenciaDe } from './cupones.ts';

const JUEVES = Date.parse('2026-10-22T18:00:00Z');

describe('cupones', () => {
	it('vigencias', () => {
		expect(new Date(vigenciaDe('finde', JUEVES).termina!).toISOString()).toBe(
			'2026-10-26T02:59:00.000Z'
		);
		expect(new Date(vigenciaDe('hotsale', JUEVES).empieza!).toISOString()).toBe(
			'2027-05-11T03:00:00.000Z'
		);
		expect(vigenciaDe('sin', JUEVES).termina).toBeNull();
		expect(new Date(lunes9(JUEVES)).toISOString()).toBe('2026-10-26T12:00:00.000Z');
	});
	it('la línea de estado', () => {
		const termina = new Date(JUEVES + 12 * 86_400_000).toISOString();
		expect(cuando({ estado: 'activo', empieza: null, termina }, JUEVES)).toBe('Vence en 12 días');
		expect(
			cuando({ estado: 'activo', empieza: null, termina: '2026-10-26T02:59:00.000Z' }, JUEVES)
		).toBe('Vence el domingo');
		expect(cuando({ estado: 'pausado', empieza: null, termina }, JUEVES)).toBe('Pausado por vos');
		expect(grande({ tipo: 'fijo', valor: 3000 })).toBe('$ 3k');
	});
	it('la estimación del handoff', () => {
		const e = estimar({ tipo: 'fijo', valor: 3000, quien: 'nuevos', minimo: 15_000, limite: 200 });
		expect(e).toEqual({ ventas: 74 * 38_000, costo: 74 * 3000, margen: 36 });
		expect(estimar({ tipo: 'pct', valor: 30, quien: 'todos', minimo: 0, limite: 300 }).margen).toBe(
			14
		);
	});
});
