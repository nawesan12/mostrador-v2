import { describe, expect, it } from 'vitest';
import empre from '@mostrador/contrato/casos/importar/empretienda.json';
import shopify from '@mostrador/contrato/casos/importar/shopify.json';
import woo from '@mostrador/contrato/casos/importar/woocommerce.json';
import { fetchFalso } from '../importar/fetch-falso.ts';
import { ErrorDefinitivo, type Origen } from './fuente.ts';
import { fuentes } from './fuentes.ts';

const origen = (plataforma: Origen['plataforma'], url: string): Origen => ({
	plataforma,
	url,
	tiendaId: 'sb',
	credencial: null
});

describe('fuentes de la demo', () => {
	it('Shopify: de a 5, con el id y el path viejo de cada producto', async () => {
		const pedidos: string[] = [];
		const f = fuentes({ fetch: fetchFalso({ '/products.json': shopify }, pedidos) });
		const fuente = await f.shopify!(origen('shopify', 'https://tienda.com'));
		const p = await fuente.pagina('productos', null);
		expect(pedidos).toEqual(['https://tienda.com/products.json?limit=5&page=1']);
		expect(p.productos?.[0]).toMatchObject({
			externo: String(shopify.products[0].id),
			url: `/products/${shopify.products[0].handle}`
		});
		// Menos de 5: no hay otra página. Las etapas que Shopify no tiene terminan enseguida.
		expect(p.siguiente).toBeNull();
		expect(await fuente.pagina('clientes', null)).toEqual({ siguiente: null });
	});

	it('WooCommerce: el path del permalink; una tienda que no responde corta', async () => {
		const f = fuentes({
			fetch: fetchFalso({ '/wp-json/wc/store/v1/products': woo.productos })
		});
		const p = await (
			await f.woocommerce!(origen('woocommerce', 'https://tienda.com'))
		).pagina('productos', null);
		expect(p.productos?.length).toBeGreaterThan(0);
		expect(p.productos?.every((x) => x.externo)).toBe(true);
		const nada = fuentes({ fetch: fetchFalso({}) });
		await expect(
			(await nada.woocommerce!(origen('woocommerce', 'https://tienda.com'))).pagina(
				'productos',
				null
			)
		).rejects.toBeInstanceOf(ErrorDefinitivo);
	});

	it('Empretienda: las fichas del sitemap de a 5, con el cursor en la posición', async () => {
		// Las 4 fichas reales del caso (dos son de otra tienda), todas en la misma y repetidas con otro
		// path para que haya dos pasos.
		const base = 'https://joliegarcia.empretienda.com.ar';
		const enBase = empre.fichas.map((x) => ({ ...x, url: base + new URL(x.url).pathname }));
		const fichas = [...enBase, ...enBase.map((x) => ({ ...x, url: `${x.url}-2` }))];
		const html = (x: (typeof fichas)[number]) =>
			['p_nombre', 'cdn_products', 'imagenes', 'categorias_flatten', 'stock']
				.map((k) => `var ${k} = ${JSON.stringify(x[k as keyof typeof x])};\n`)
				.join('') + `<div class="product-vip__description">${x.descripcion}</div>`;
		const sitemap = fichas
			.map((x) => `<url><loc>${x.url}</loc><priority>0.8</priority></url>`)
			.concat(`<url><loc>${base}/</loc><priority>1</priority></url>`)
			// Una ficha de otra tienda no se lee.
			.concat('<url><loc>https://otra.com/x</loc><priority>0.8</priority></url>')
			.join('');
		const rutas = Object.fromEntries([
			['/sitemap.xml', sitemap],
			...fichas.map((x) => [new URL(x.url).pathname, html(x)])
		]);
		const f = fuentes({ fetch: fetchFalso(rutas) });
		const fuente = await f.empretienda!(origen('empretienda', 'joliegarcia'));
		const p1 = await fuente.pagina('productos', null);
		expect(p1.productos).toHaveLength(5);
		expect(p1.productos?.[0]).toMatchObject({
			externo: new URL(fichas[0].url).pathname,
			url: new URL(fichas[0].url).pathname
		});
		expect(p1.siguiente).toEqual({ desde: 5 });
		const p2 = await fuente.pagina('productos', p1.siguiente);
		expect(p2.productos).toHaveLength(3);
		expect(p2.siguiente).toBeNull();
		// Sin sitemap no es una tienda de Empretienda.
		const nada = await fuentes({ fetch: fetchFalso({}) }).empretienda!(
			origen('empretienda', 'joliegarcia')
		);
		await expect(nada.pagina('productos', null)).rejects.toBeInstanceOf(ErrorDefinitivo);
	});

	it('sin una dirección pública o sin credencial no se abre', () => {
		const f = fuentes({ fetch: fetchFalso({}) });
		expect(() => f.shopify!(origen('shopify', 'http://127.0.0.1:8080'))).toThrow(ErrorDefinitivo);
		expect(() => f.tiendanube!(origen('tiendanube', ''))).toThrow(ErrorDefinitivo);
	});
});
