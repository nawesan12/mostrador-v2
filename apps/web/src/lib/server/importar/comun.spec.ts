import { describe, expect, it } from 'vitest';
import { ajustar, origenDe, pesos, textoRico, type Crudo } from './comun.ts';

const crudo = (c: Partial<Crudo> = {}): Crudo => ({
	nombre: 'Producto',
	descripcion: '',
	categoria: null,
	precio: 100,
	sku: '',
	stock: 2,
	estimado: false,
	fotos: [],
	opciones: [],
	variantes: [],
	...c
});

describe('importar: lo común', () => {
	it('sólo dominios públicos por https', () => {
		expect(origenDe('mitienda', '.myshopify.com')).toBe('https://mitienda.myshopify.com');
		expect(origenDe('https://tienda.com.ar/productos?x=1')).toBe('https://tienda.com.ar');
		expect(origenDe('mitienda')).toBeNull();
		expect(origenDe('http://localhost:8080')).toBeNull();
		expect(origenDe('https://127.0.0.1')).toBeNull();
		expect(origenDe('https://[::1]')).toBeNull();
		expect(origenDe('https://tienda.com.ar:8443')).toBeNull();
	});

	it('precios a pesos enteros, la mitad para arriba', () => {
		expect(pesos('50.00')).toBe(50);
		expect(pesos('19.99')).toBe(20);
		expect(pesos('12499.50')).toBe(12500);
		expect(pesos('1250049', 2)).toBe(12500);
		expect(pesos('1250050', 2)).toBe(12501);
		expect(pesos('900', 0)).toBe(900);
		expect(pesos('')).toBeNull();
		expect(pesos('-5')).toBeNull();
	});

	it('HTML de catálogo a texto', () => {
		expect(textoRico('<p>Hola\n   mundo</p><ul><li>Uno</li><li>Dos&#8217;s</li></ul>')).toBe(
			'Hola mundo\nUno\nDos’s'
		);
		expect(textoRico('<p> </p>')).toBe('');
	});

	it('recorta lo que no entra y lo marca', () => {
		const r = ajustar(
			crudo({
				nombre: 'x'.repeat(130),
				fotos: Array.from({ length: 9 }, (_, i) => `https://f/${i}.jpg`)
			})
		);
		expect(r?.recortado).toBe(true);
		expect(r?.producto.nombre).toHaveLength(120);
		expect(r?.producto.fotos).toHaveLength(8);
		expect(ajustar(crudo({ nombre: ' a ' }))).toBeNull();
		expect(ajustar(crudo())?.recortado).toBe(false);
	});

	it('más de 100 combinaciones: saca valores de la opción más larga', () => {
		const valores = (n: number) => Array.from({ length: n }, (_, i) => `v${i}`);
		const r = ajustar(
			crudo({
				opciones: [
					{ nombre: 'Talle', valores: valores(12) },
					{ nombre: 'Color', valores: valores(10) },
					{ nombre: 'Color', valores: ['x'] },
					{ nombre: 'Otra', valores: ['y'] },
					{ nombre: 'Cuarta', valores: ['z'] }
				],
				variantes: [
					{ valores: ['v0', 'v0', 'x', 'y', 'z'], sku: 'A', precio: 90, stock: 3, foto: null },
					{ valores: ['v11', 'v0', 'x', 'y', 'z'], sku: 'B', precio: 80, stock: 1, foto: null },
					{ valores: ['v1', 'V0', 'x', 'y', 'z'], sku: 'C', precio: 95, stock: 1, foto: null }
				]
			})
		);
		expect(r?.recortado).toBe(true);
		expect(r?.producto.opciones?.map((o) => [o.nombre, o.valores.length])).toEqual([
			['Talle', 10],
			['Color', 10],
			['Otra', 1]
		]);
		// v11 quedó afuera; "V0" es el mismo valor que "v0".
		expect(r?.producto.variantes?.map((v) => v.sku)).toEqual(['A', 'C']);
		expect(r?.producto.precio).toBe(90);
		expect(r?.producto.variantes?.map((v) => v.precio)).toEqual([null, 95]);
	});
});
