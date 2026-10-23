import { describe, expect, it } from 'vitest';
import { fichasDelSitemap, origenDe, productoDeFicha, STOCK_ILIMITADO } from './empretienda.ts';

const O = 'https://mitienda.empretienda.com.ar';
const ficha = (stock: unknown[], extra = '') => `<html><script>
        var p_nombre = "Medias  Floyd";
        var cdn_products = "https://cdn.ejemplo/";
        var imagenes = [{"idImagenes":1,"i_link":"a.jpg"}];
        var stock = ${JSON.stringify(stock)};
        var categorias_flatten = [{"c_link_full":"\\/medias","c_nombre":"Medias"},{"c_link_full":"\\/medias\\/lisas","c_nombre":"Lisas"}];
</script><div class="product-vip__description ql-container"><p>Algodón&nbsp;peinado</p><p>Talle único</p></div>${extra}</html>`;
const s = (cant: number, precio: number, valor?: string, ilimitado = 0) => ({
	s_cantidad: cant,
	s_ilimitado: ilimitado,
	s_precio: precio,
	s_oferta: 0,
	s_precio_oferta: 0,
	s_mostrar_precio: 1,
	valoratributo: valor ? [{ valor: { vat_valor: valor } }] : []
});

describe('empretienda', () => {
	it('entiende el link de la tienda de varias formas', () => {
		expect(origenDe('mitienda')).toBe(O);
		expect(origenDe('mitienda.empretienda.com.ar')).toBe(O);
		expect(origenDe('https://mitienda.empretienda.com.ar/medias?x=1')).toBe(O);
		expect(origenDe('http://localhost:8080')).toBeNull();
		expect(origenDe('')).toBeNull();
	});

	it('del sitemap toma sólo las fichas de esa tienda', () => {
		const xml = `<urlset><url><loc>${O}</loc><priority>1.0</priority></url>
			<url><loc>${O}/medias</loc><priority>0.9</priority></url>
			<url><loc>${O}/medias/floyd</loc><priority>0.8</priority></url>
			<url><loc>https://otra.com/x</loc><priority>0.8</priority></url></urlset>`;
		expect(fichasDelSitemap(xml, O)).toEqual([`${O}/medias/floyd`]);
	});

	it('arma el producto: variantes, stock sumado, categoría más específica', () => {
		const p = productoDeFicha(
			ficha([s(2, 3800, 'Negro'), s(1, 3500, 'Fucsia')]),
			`${O}/medias/lisas/floyd`
		);
		expect(p).toMatchObject({
			nombre: 'Medias Floyd',
			precio: 3500,
			stock: 3,
			categoria: 'Lisas',
			foto: 'https://cdn.ejemplo/a.jpg'
		});
		expect(p?.descripcion).toBe(
			'Algodón peinado\nTalle único\n\nVariantes: Negro (2), Fucsia (1).'
		);
	});

	it('stock ilimitado y páginas que no son fichas', () => {
		expect(productoDeFicha(ficha([s(0, 100, undefined, 1)]), `${O}/medias/x`)?.stock).toBe(
			STOCK_ILIMITADO
		);
		expect(productoDeFicha('<html>portada</html>', O)).toBeNull();
	});
});
