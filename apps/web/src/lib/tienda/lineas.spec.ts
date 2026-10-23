import { describe, expect, it } from 'vitest';
import { claveLinea, detalleLineas, migrarLineas } from './lineas.ts';
import type { ProductoPublico } from './tipos.ts';

const base = {
	descripcion: '',
	categoria: null,
	anterior: 30_000,
	etiqueta: null,
	fotos: []
};
const remera: ProductoPublico = {
	...base,
	id: 'remera',
	nombre: 'Remera',
	precio: 28_000,
	foto: 'f1',
	stock: 5,
	fotos: ['f1', 'f2'],
	opciones: [{ nombre: 'Talle', valores: ['M', 'XL'] }],
	variantes: [
		{ id: 'm', valores: { Talle: 'M' }, precio: null, stock: 3, foto: null },
		{ id: 'xl', valores: { Talle: 'XL' }, precio: 31_000, stock: 2, foto: 'f2' }
	]
};
const taza: ProductoPublico = {
	...base,
	id: 'taza',
	nombre: 'Taza',
	precio: 9_000,
	foto: null,
	stock: 4,
	opciones: [],
	variantes: []
};

describe('líneas del carrito', () => {
	it('migra los carritos viejos (sin varianteId) y descarta lo roto o repetido', () => {
		const viejo = [
			{ id: 'taza', cantidad: 2 },
			{ id: 'remera', varianteId: 'm', cantidad: 120 },
			{ id: 'taza', cantidad: 1 },
			{ id: '', cantidad: 1 },
			{ id: 'x', cantidad: 0 },
			null,
			'basura'
		];
		expect(migrarLineas(viejo)).toEqual([
			{ id: 'taza', varianteId: null, cantidad: 2 },
			{ id: 'remera', varianteId: 'm', cantidad: 99 }
		]);
		expect(migrarLineas({ no: 'es lista' })).toEqual([]);
	});

	it('una línea por producto y variante', () => {
		expect(claveLinea({ id: 'remera', varianteId: 'm' })).toBe('remera~m');
		expect(claveLinea({ id: 'taza', varianteId: null })).toBe('taza');
	});

	it('el detalle usa precio, foto y stock de la variante, con su texto', () => {
		const d = detalleLineas(
			[
				{ id: 'remera', varianteId: 'xl', cantidad: 2 },
				{ id: 'taza', varianteId: null, cantidad: 1 }
			],
			[remera, taza]
		);
		expect(d[0]).toMatchObject({ clave: 'remera~xl', variante: 'Talle XL', subtotal: 62_000 });
		expect(d[0].producto).toMatchObject({ precio: 31_000, foto: 'f2', stock: 2, anterior: null });
		expect(d[1]).toMatchObject({ clave: 'taza', variante: null, subtotal: 9_000 });
	});

	it('descarta variantes que ya no están y productos con opciones sin variante', () => {
		const d = detalleLineas(
			[
				{ id: 'remera', varianteId: null, cantidad: 1 },
				{ id: 'remera', varianteId: 'borrada', cantidad: 1 },
				{ id: 'taza', varianteId: 'm', cantidad: 1 },
				{ id: 'no-existe', varianteId: null, cantidad: 1 }
			],
			[remera, taza]
		);
		expect(d).toEqual([]);
	});
});
