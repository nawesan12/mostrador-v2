import { describe, expect, it } from 'vitest';
import { costoPorUso, cuando, grande, lunes9, vigenciaDe } from './cupones.ts';

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
	it('el costo de cada uso sale del ticket real', () => {
		expect(costoPorUso({ tipo: 'pct', valor: 10, minimo: 0 }, 20_000)).toEqual({
			ticket: 20_000,
			costo: 2_000
		});
		// La compra mínima manda si es mayor que el ticket.
		expect(costoPorUso({ tipo: 'pct', valor: 10, minimo: 30_000 }, 20_000)?.costo).toBe(3_000);
		expect(costoPorUso({ tipo: 'fijo', valor: 3000, minimo: 0 }, 0)?.costo).toBe(3000);
		// Sin ventas no hay ticket: no se inventa.
		expect(costoPorUso({ tipo: 'pct', valor: 10, minimo: 0 }, 0)).toBeNull();
		expect(costoPorUso({ tipo: 'envio', valor: 0, minimo: 0 }, 20_000)).toBeNull();
	});
});
