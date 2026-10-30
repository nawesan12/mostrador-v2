import { describe, expect, it } from 'vitest';
import { mesesCalendario } from './crecer.ts';

describe('Informes: meses del eje', () => {
	it('el 31 de octubre no repite agosto ni saltea noviembre', () => {
		// 31/10 a las 23:30 en Argentina (ya es 1/11 en UTC).
		const ahora = Date.UTC(2026, 10, 1, 2, 30);
		expect(mesesCalendario(ahora)).toEqual([
			'nov',
			'dic',
			'ene',
			'feb',
			'mar',
			'abr',
			'may',
			'jun',
			'jul',
			'ago',
			'sept',
			'oct'
		]);
		expect(mesesCalendario(ahora, 'long')[10]).toBe('septiembre');
	});
});
