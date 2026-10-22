import { describe, expect, it } from 'vitest';
import { ARTICULOS, CATEGORIAS, buscar } from './articulos.ts';

describe('ayuda', () => {
	it('cada artículo es de una categoría que existe y los slugs no se repiten', () => {
		for (const a of ARTICULOS) expect(CATEGORIAS.some((c) => c.slug === a.categoria)).toBe(true);
		expect(new Set(ARTICULOS.map((a) => a.slug)).size).toBe(ARTICULOS.length);
	});
	it('el buscador encuentra lo que se pregunta', () => {
		expect(buscar('¿cómo pongo mi dominio?')[0].slug).toBe('dominio-propio');
		expect(buscar('factura A')[0].slug).toBe('factura-a-b-c');
		expect(buscar('cuotas sin interés')[0].slug).toBe('mercado-pago-y-cuotas');
		expect(buscar('zzzz')).toEqual([]);
	});
});
