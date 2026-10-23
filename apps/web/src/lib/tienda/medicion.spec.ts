import { describe, expect, it } from 'vitest';
import {
	contenidoMeta,
	itemsGa,
	pideConsentimiento,
	validarGa4,
	validarPixel
} from './medicion.ts';
import type { ProductoPublico } from './tipos.ts';

describe('validación de IDs', () => {
	it('GA4: G- y de 6 a 12 letras o números, en mayúsculas', () => {
		expect(validarGa4('G-AB12CD34EF')).toEqual({ ok: true, valor: 'G-AB12CD34EF' });
		expect(validarGa4('  g-ab12cd34ef ')).toEqual({ ok: true, valor: 'G-AB12CD34EF' });
		expect(validarGa4('')).toEqual({ ok: true, valor: null });
		expect(validarGa4(null)).toEqual({ ok: true, valor: null });
		for (const malo of ['UA-123456-1', 'G-', 'G-ABC', 'AB12CD34EF', 'G-AB12CD34EF1234', 'G-AB"<x>'])
			expect(validarGa4(malo).ok, malo).toBe(false);
		expect(validarGa4('UA-123456-1')).toMatchObject({ error: expect.stringContaining('GA4') });
	});

	it('píxel de Meta: sólo dígitos, 15 o 16', () => {
		expect(validarPixel('123456789012345')).toEqual({ ok: true, valor: '123456789012345' });
		expect(validarPixel(' 1234567890123456 ')).toEqual({ ok: true, valor: '1234567890123456' });
		expect(validarPixel('')).toEqual({ ok: true, valor: null });
		for (const malo of [
			'12345678901234',
			'12345678901234567',
			'12345678901234a',
			'1234 5678 9012 345'
		])
			expect(validarPixel(malo).ok, malo).toBe(false);
	});
});

describe('eventos', () => {
	const productos = [
		{ id: 'a', nombre: 'Mate', categoria: 'Mates', precio: 1000 },
		{ id: 'b', nombre: 'Bombilla', categoria: null, precio: 500 }
	] as ProductoPublico[];

	it('ítems de GA4 y contenido de Meta con valor y moneda', () => {
		const items = itemsGa(
			[
				{ id: 'a', cantidad: 2 },
				{ id: 'b', cantidad: 1 },
				{ id: 'ya-no-esta', cantidad: 1 }
			],
			productos
		);
		expect(items).toEqual([
			{ item_id: 'a', item_name: 'Mate', item_category: 'Mates', price: 1000, quantity: 2 },
			{ item_id: 'b', item_name: 'Bombilla', price: 500, quantity: 1 }
		]);
		expect(contenidoMeta(items, 2500)).toEqual({
			content_ids: ['a', 'b'],
			contents: [
				{ id: 'a', quantity: 2, item_price: 1000 },
				{ id: 'b', quantity: 1, item_price: 500 }
			],
			content_type: 'product',
			num_items: 3,
			value: 2500,
			currency: 'ARS'
		});
	});

	it('el aviso de cookies sólo si hay IDs y no es una vitrina', () => {
		expect(
			pideConsentimiento({ vitrina: false, medicion: { ga4: 'G-AB12CD34EF', pixel: null } })
		).toBe(true);
		expect(
			pideConsentimiento({ vitrina: false, medicion: { ga4: null, pixel: '123456789012345' } })
		).toBe(true);
		expect(pideConsentimiento({ vitrina: false, medicion: { ga4: null, pixel: null } })).toBe(
			false
		);
		expect(
			pideConsentimiento({ vitrina: true, medicion: { ga4: 'G-AB12CD34EF', pixel: null } })
		).toBe(false);
	});
});
