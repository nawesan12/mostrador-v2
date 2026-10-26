import { describe, expect, it } from 'vitest';
import { vendibles } from './vendibles.ts';

const base = { id: 'p1', nombre: 'Remera', sku: 'R1', precio: 1000, stock: 9 };
const variante = (id: string, talle: string, extra = {}) => ({
	id,
	texto: `Talle ${talle}`,
	sku: '',
	precio: null,
	stock: 2,
	activa: true,
	...extra
});

describe('vendibles', () => {
	it('sin opciones, el producto es un renglón', () => {
		expect(vendibles([base])).toEqual([{ ...base, clave: 'p1', varianteId: null, variante: '' }]);
	});

	it('con opciones, un renglón por variante activa con su precio, stock y SKU', () => {
		const p = {
			...base,
			variantes: [
				variante('a', 'S'),
				variante('b', 'M', { precio: 1200, sku: 'R1-M', stock: 5 }),
				variante('c', 'L', { activa: false })
			]
		};
		const r = vendibles([p]);
		expect(r.map((x) => [x.clave, x.nombre, x.precio, x.stock, x.sku])).toEqual([
			['p1:a', 'Remera · Talle S', 1000, 2, 'R1'],
			['p1:b', 'Remera · Talle M', 1200, 5, 'R1-M']
		]);
		expect(r[1]).toMatchObject({ id: 'p1', varianteId: 'b', variante: 'Talle M' });
	});
});
