import { describe, expect, it } from 'vitest';
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

	it('sin una dirección pública o sin credencial no se abre', () => {
		const f = fuentes({ fetch: fetchFalso({}) });
		expect(() => f.shopify!(origen('shopify', 'http://127.0.0.1:8080'))).toThrow(ErrorDefinitivo);
		expect(() => f.tiendanube!(origen('tiendanube', ''))).toThrow(ErrorDefinitivo);
	});
});
