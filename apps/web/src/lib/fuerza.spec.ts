import { describe, expect, it } from 'vitest';
import { fuerza } from './fuerza.ts';

describe('fuerza', () => {
	it('sube con largo, mayúsculas y números', () => {
		expect(fuerza('').nivel).toBe(0);
		expect(fuerza('abc').nivel).toBeLessThanOrEqual(1);
		expect(fuerza('abcdefgh').nivel).toBe(1);
		expect(fuerza('Abcdefgh').nivel).toBe(2);
		expect(fuerza('Abcdefghijkl').nivel).toBe(3);
		expect(fuerza('Abcdefghijk1').nivel).toBe(4);
		expect(fuerza('Abcdefghijk1').texto).toBe('Muy segura');
	});
});
