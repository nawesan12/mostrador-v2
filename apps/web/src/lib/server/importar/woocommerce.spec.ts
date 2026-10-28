import { describe, expect, it } from 'vitest';
import casos from '@mostrador/contrato/casos/importar/woocommerce.json';
import esperado from '@mostrador/contrato/casos/importar/woocommerce.esperado.json';
import { conCabeceras, fetchFalso } from './fetch-falso.ts';
import { descubrirWoo, leerPaginaWoo, resultadoWoo, type ProductoWoo } from './woocommerce.ts';

const API = '/wp-json/wc/store/v1/products';
const productos = casos.productos as ProductoWoo[];
const variaciones = casos.variaciones as ProductoWoo[];

describe('importar: WooCommerce (mismos casos que Go)', () => {
	it('Store API → productos de Mostrador, con las variaciones', () => {
		expect(resultadoWoo({ productos, variaciones })).toEqual(esperado);
	});

	it('el total sale de X-WP-Total', async () => {
		const f = fetchFalso({
			[`${API}?per_page=1&page=1`]: conCabeceras([productos[0]], { 'x-wp-total': '12' })
		});
		expect(await descubrirWoo('https://tienda.com.ar', f)).toEqual({ total: 12 });
		expect(await descubrirWoo('https://tienda.com.ar', fetchFalso({}))).toBeNull();
	});

	it('una página pide las variaciones de todos sus variables de una vez', async () => {
		const pedidos: string[] = [];
		const f = fetchFalso(
			{
				[`${API}?per_page=5&page=1`]: productos,
				[`${API}?type=variation&parent=40&per_page=100&page=1`]: conCabeceras(variaciones, {
					'x-wp-totalpages': '1'
				})
			},
			pedidos
		);
		const p = await leerPaginaWoo('https://tienda.com.ar', 1, f);
		expect(p?.items.map((i) => i.producto)).toEqual(esperado.productos);
		expect(pedidos).toHaveLength(2);
	});

	it('si las variaciones no responden, quedan con el precio y el stock del producto', async () => {
		const f = fetchFalso({ [`${API}?per_page=5&page=1`]: [productos[1]] });
		const p = await leerPaginaWoo('https://tienda.com.ar', 1, f);
		const remera = p?.items[0];
		expect(remera?.estimado).toBe(true);
		expect(remera?.producto.variantes).toHaveLength(6);
		expect(remera?.producto.variantes?.every((v) => v.stock === 1 && v.precio === null)).toBe(true);
	});
});
