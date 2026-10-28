import { describe, expect, it } from 'vitest';
import { conCabeceras, fetchFalso } from './fetch-falso.ts';
import { descubrir } from './fuentes.ts';

const woo = { '/wp-json/wc/store/v1/products': conCabeceras([{ id: 1 }], { 'x-wp-total': '7' }) };

describe('importar: qué plataforma es', () => {
	it('el dominio manda: un .myshopify.com es Shopify aunque se eligiera otra', async () => {
		const pedidos: string[] = [];
		const f = fetchFalso(
			{ '/meta.json': { myshopify_domain: 'a.myshopify.com', published_products_count: 11 } },
			pedidos
		);
		expect(await descubrir('woocommerce', 'https://mitienda.myshopify.com/collections', f)).toEqual(
			{
				fuente: 'shopify',
				origen: 'https://mitienda.myshopify.com',
				total: 11,
				paginas: 3,
				omitidas: 0
			}
		);
		expect(pedidos).toEqual(['https://mitienda.myshopify.com/meta.json']);
	});

	it('con dominio propio prueba la elegida y después las otras', async () => {
		const r = await descubrir('shopify', 'tienda.com.ar', fetchFalso(woo));
		expect(r).toMatchObject({ fuente: 'woocommerce', total: 7, paginas: 2 });
	});

	it('un nombre suelto se completa con el dominio de la plataforma', async () => {
		const xml = `<urlset><url><loc>https://mitienda.empretienda.com.ar/a/b</loc><priority>0.8</priority></url></urlset>`;
		const r = await descubrir('empretienda', 'mitienda', fetchFalso({ '/sitemap.xml': xml }));
		expect(r).toMatchObject({
			fuente: 'empretienda',
			paginas: 1,
			fichas: ['https://mitienda.empretienda.com.ar/a/b']
		});
	});

	it('links inválidos, tiendas vacías y que no responden', async () => {
		expect(await descubrir('shopify', 'http://localhost:3000', fetchFalso({}))).toBe('invalida');
		expect(await descubrir('woocommerce', 'mitienda', fetchFalso({}))).toBe('invalida');
		expect(await descubrir('shopify', 'tienda.com.ar', fetchFalso({}))).toBe('no-encontrada');
		const vacia = { '/meta.json': { myshopify_domain: 'a', published_products_count: 0 } };
		expect(await descubrir('shopify', 'tienda.com.ar', fetchFalso(vacia))).toBe('vacia');
	});
});
