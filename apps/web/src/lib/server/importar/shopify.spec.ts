import { describe, expect, it } from 'vitest';
import casos from '@mostrador/contrato/casos/importar/shopify.json';
import esperado from '@mostrador/contrato/casos/importar/shopify.esperado.json';
import { fetchFalso as falso } from './fetch-falso.ts';
import { descubrirShopify, leerPaginaShopify, resultadoShopify } from './shopify.ts';

describe('importar: Shopify (mismos casos que Go)', () => {
	it('products.json → productos de Tu Local Web', () => {
		expect(resultadoShopify(casos)).toEqual(esperado);
	});

	it('descubre con meta.json y lee de a 5', async () => {
		const pedidos: string[] = [];
		const f = falso(
			{
				'/meta.json': { myshopify_domain: 'x.myshopify.com', published_products_count: 3 },
				'/products.json?limit=5&page=1': casos
			},
			pedidos
		);
		expect(await descubrirShopify('https://tienda.com', f)).toEqual({ total: 3 });
		const p = await leerPaginaShopify('https://tienda.com', 1, f);
		expect(p?.items).toHaveLength(3);
		expect(p?.leidos).toBe(3);
		expect(pedidos).toEqual([
			'https://tienda.com/meta.json',
			'https://tienda.com/products.json?limit=5&page=1'
		]);
	});

	it('sin meta.json prueba la primera página; si no hay nada, no es Shopify', async () => {
		const f = falso({ '/products.json?limit=1': { products: [] } });
		expect(await descubrirShopify('https://tienda.com', f)).toEqual({ total: null });
		expect(await descubrirShopify('https://tienda.com', falso({}))).toBeNull();
	});
});
