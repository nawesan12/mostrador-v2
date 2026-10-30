import { describe, expect, it } from 'vitest';
import { enOrden } from './secciones.ts';

describe('enOrden', () => {
	const malbec = ['hero', 'prods', 'cats', 'revs', 'ig'] as const;

	it('sin reordenar, va el orden del diseño de la plantilla', () => {
		expect(enOrden(['hero', 'cats', 'prods', 'revs', 'ig', 'wa'], malbec)).toEqual([
			'hero',
			'prods',
			'cats',
			'revs',
			'ig'
		]);
	});

	it('las ocultas no aparecen', () => {
		expect(enOrden(['hero', 'prods', 'wa'], malbec)).toEqual(['hero', 'prods']);
	});

	it('reordenadas en Mi tienda, va el orden de la tienda', () => {
		expect(enOrden(['prods', 'hero', 'cats', 'revs', 'ig', 'wa'], malbec)).toEqual([
			'prods',
			'hero',
			'cats',
			'revs',
			'ig',
			'wa'
		]);
	});
});
