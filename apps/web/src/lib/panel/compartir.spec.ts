import { describe, expect, it } from 'vitest';
import { linkProducto, textoProducto } from './compartir.ts';

describe('compartir un producto', () => {
	it('el link va por el slug y, si no hay, por el id', () => {
		expect(linkProducto('luna', { id: 'abc', slug: 'mate-de-calabaza' })).toBe(
			'https://tulocalweb.com.ar/t/luna/producto/mate-de-calabaza'
		);
		expect(linkProducto('luna', { id: 'abc-1', slug: '' })).toBe(
			'https://tulocalweb.com.ar/t/luna/producto/abc-1'
		);
	});

	it('el texto: nombre · $ precio — link', () => {
		expect(textoProducto('Mate de calabaza', 12000, 'https://x.ar/p')).toBe(
			'Mate de calabaza · $ 12.000 — https://x.ar/p'
		);
	});
});
