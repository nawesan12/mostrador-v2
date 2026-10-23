import { describe, expect, it } from 'vitest';
import { EMPRESA, anioHotSale, fechaIso } from './datos.ts';

describe('datos de las piezas de marca', () => {
	it('la fecha es la de Argentina, no la de UTC', () => {
		// 01:30 UTC del 1/11 son las 22:30 del 31/10 en Argentina.
		expect(fechaIso(new Date(Date.UTC(2026, 10, 1, 1, 30)))).toBe('2026-10-31');
	});

	it('el Hot Sale es el del año que viene pasado mayo', () => {
		expect(anioHotSale(new Date(Date.UTC(2026, 3, 10, 15)))).toBe(2026);
		expect(anioHotSale(new Date(Date.UTC(2026, 4, 31, 15)))).toBe(2026);
		expect(anioHotSale(new Date(Date.UTC(2026, 9, 23, 15)))).toBe(2027);
	});

	it('titular, CUIT y domicilio salen de legales.ts', () => {
		expect(EMPRESA.titular).toBe('Nahuel Santillán');
		expect(EMPRESA.cuit).toBe('20-44859234-1');
		expect(EMPRESA.ciudad).toBe('Mar del Plata');
		expect(EMPRESA.region).toBe('Provincia de Buenos Aires');
	});
});
