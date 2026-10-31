import { describe, expect, it } from 'vitest';
import { faltantes, textoFaltantes } from './faltantes.ts';
import type { ProductoPublico } from './tipos.ts';

const producto = (
	id: string,
	stock: number,
	extra: Partial<ProductoPublico> = {}
): ProductoPublico => ({
	id,
	nombre: id === 'mate' ? 'Mate' : 'Bombilla',
	descripcion: '',
	categoria: null,
	precio: 1000,
	anterior: null,
	foto: null,
	etiqueta: null,
	stock,
	fotos: [],
	opciones: [],
	variantes: [],
	...extra
});

describe('faltantes', () => {
	it('nada si alcanza el stock', () => {
		expect(
			faltantes([{ id: 'mate', varianteId: null, cantidad: 2 }], [producto('mate', 2)])
		).toEqual([]);
	});
	it('agotado, quedan menos y ya no se vende', () => {
		const fs = faltantes(
			[
				{ id: 'mate', varianteId: null, cantidad: 1 },
				{ id: 'bombilla', varianteId: null, cantidad: 3 },
				{ id: 'termo', varianteId: null, cantidad: 1 }
			],
			[producto('mate', 0), producto('bombilla', 2)]
		);
		expect(fs).toEqual([
			{ clave: 'mate', nombre: 'Mate', quedan: 0, pedidas: 1 },
			{ clave: 'bombilla', nombre: 'Bombilla', quedan: 2, pedidas: 3 },
			{ clave: 'termo', nombre: 'Un producto', quedan: 0, pedidas: 1 }
		]);
	});
	it('mira el stock de la variante', () => {
		const p = producto('mate', 5, {
			opciones: [{ nombre: 'Color', valores: ['Negro', 'Rojo'] }],
			variantes: [
				{ id: 'n', valores: { Color: 'Negro' }, precio: null, stock: 0, foto: null },
				{ id: 'r', valores: { Color: 'Rojo' }, precio: null, stock: 5, foto: null }
			]
		});
		const fs = faltantes(
			[
				{ id: 'mate', varianteId: 'n', cantidad: 1 },
				{ id: 'mate', varianteId: 'r', cantidad: 1 }
			],
			[p]
		);
		expect(fs).toHaveLength(1);
		expect(fs[0]).toMatchObject({ clave: 'mate~n', quedan: 0 });
		expect(fs[0].nombre).toContain('Negro');
	});
});

describe('textoFaltantes', () => {
	it('uno agotado, uno con menos y varios', () => {
		expect(textoFaltantes([{ clave: 'm', nombre: 'Mate', quedan: 0, pedidas: 1 }]).titulo).toBe(
			'Se agotó Mate'
		);
		expect(textoFaltantes([{ clave: 'm', nombre: 'Mate', quedan: 2, pedidas: 3 }])).toEqual({
			titulo: 'De Mate quedan 2',
			detalle: 'Tenés 3 en el carrito: ajustá la cantidad para seguir.'
		});
		const varios = textoFaltantes([
			{ clave: 'm', nombre: 'Mate', quedan: 0, pedidas: 1 },
			{ clave: 'b', nombre: 'Bombilla', quedan: 1, pedidas: 2 }
		]);
		expect(varios.titulo).toBe('Cambió el stock de tu carrito');
		expect(varios.detalle).toBe('Mate: agotado · Bombilla: quedan 1.');
	});
});
