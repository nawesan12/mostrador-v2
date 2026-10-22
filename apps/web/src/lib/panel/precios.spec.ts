import { describe, expect, it } from 'vitest';
import { aumentar, proximoLunes, redondear } from './precios.ts';

describe('aumento de precios', () => {
	it('redondea como el handoff', () => {
		expect(redondear(9012, '90')).toBe(8990);
		expect(redondear(9012, '500')).toBe(9000);
		expect(redondear(9012.4, 'no')).toBe(9012);
		expect(aumentar(42900, 8, '90')).toBe(46290);
	});
	it('el lunes siguiente, a las 00 de Argentina', () => {
		// Jueves 22/10/2026 a las 15 h de Argentina.
		const lunes = new Date(proximoLunes(Date.parse('2026-10-22T18:00:00Z')));
		expect(lunes.toISOString()).toBe('2026-10-26T03:00:00.000Z');
		// Un lunes a la mañana programa para el lunes que viene.
		expect(new Date(proximoLunes(Date.parse('2026-10-26T12:00:00Z'))).toISOString()).toBe(
			'2026-11-02T03:00:00.000Z'
		);
	});
});
