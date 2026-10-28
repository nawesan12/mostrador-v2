import { describe, expect, it } from 'vitest';
import type { ProductoPublico, TiendaPublica } from '#lib/tienda/tipos.ts';
import { feedXml, fotoFeed, precioFeed, xml } from './feed.ts';

const producto = (p: Partial<ProductoPublico>): ProductoPublico => ({
	id: 'p1',
	nombre: 'Mate',
	descripcion: 'Calabaza curada',
	categoria: 'Mates',
	precio: 12345,
	anterior: null,
	foto: 'https://cdn.x.com/p.jpg',
	etiqueta: null,
	stock: 3,
	fotos: [],
	opciones: [],
	variantes: [],
	...p
});

const tienda = (productos: ProductoPublico[]): TiendaPublica =>
	({
		slug: 'mate-y-co',
		nombre: 'Mate & Co',
		productos,
		envios: { gratisDesde: 100_000, retiro: true, moto: false, correo: true }
	}) as TiendaPublica;

/** Lo que va dentro de <item> de un producto, campo por campo. */
function campos(x: string, id: string) {
	const item = x.split('<item>').find((i) => i.includes(`<g:id>${id}</g:id>`))!;
	return Object.fromEntries(
		[...item.matchAll(/<(g:[a-z_]+)>([^<]*)<\/\1>/g)].map((m) => [m[1], m[2]])
	);
}

describe('feed de productos', () => {
	it('arma RSS 2.0 con el espacio de nombres de Google', () => {
		const x = feedXml(tienda([producto({})]), 'https://mostrador.iamnaw.dev');
		expect(x.startsWith('<?xml version="1.0" encoding="UTF-8"?>\n<rss version="2.0"')).toBe(true);
		expect(x).toContain('xmlns:g="http://base.google.com/ns/1.0"');
		expect(x).toContain('<title>Mate &amp; Co</title>');
		expect(x).toContain('<link>https://mostrador.iamnaw.dev/t/mate-y-co</link>');
		expect(x.match(/<item>/g)).toHaveLength(1);
		expect(campos(x, 'p1')).toEqual({
			'g:id': 'p1',
			'g:title': 'Mate',
			'g:description': 'Calabaza curada',
			'g:link': 'https://mostrador.iamnaw.dev/t/mate-y-co/producto/p1',
			'g:image_link': 'https://cdn.x.com/p.jpg',
			'g:availability': 'in_stock',
			'g:price': '12345.00 ARS',
			'g:brand': 'Mate &amp; Co',
			'g:condition': 'new',
			'g:identifier_exists': 'no',
			'g:product_type': 'Mates'
		});
	});

	it('escapa XML y saca caracteres de control', () => {
		expect(xml(`<a href="x">Tom & 'Jerry'</a>\u0001`)).toBe(
			'&lt;a href=&quot;x&quot;&gt;Tom &amp; &apos;Jerry&apos;&lt;/a&gt;'
		);
		const x = feedXml(
			tienda([producto({ nombre: 'Yerba <Premium> & "Suave"', descripcion: 'Sin\u0008 TACC' })]),
			'https://x.dev'
		);
		const c = campos(x, 'p1');
		expect(c['g:title']).toBe('Yerba &lt;Premium&gt; &amp; &quot;Suave&quot;');
		expect(c['g:description']).toBe('Sin TACC');
		expect([...x].some((ch) => ch.charCodeAt(0) < 9)).toBe(false);
	});

	it('disponibilidad según el stock', () => {
		const x = feedXml(
			tienda([producto({ id: 'hay', stock: 1 }), producto({ id: 'no', stock: 0 })]),
			'https://x.dev'
		);
		expect(campos(x, 'hay')['g:availability']).toBe('in_stock');
		expect(campos(x, 'no')['g:availability']).toBe('out_of_stock');
	});

	it('precios en ARS: oferta con precio tachado como price y el actual como sale_price', () => {
		expect(precioFeed(12345)).toBe('12345.00 ARS');
		expect(precioFeed(0)).toBe('0.00 ARS');
		expect(precioFeed(99.5)).toBe('99.50 ARS');
		const x = feedXml(
			tienda([
				producto({ id: 'oferta', precio: 8000, anterior: 10000 }),
				producto({ id: 'raro', precio: 8000, anterior: 5000 })
			]),
			'https://x.dev'
		);
		expect(campos(x, 'oferta')['g:price']).toBe('10000.00 ARS');
		expect(campos(x, 'oferta')['g:sale_price']).toBe('8000.00 ARS');
		expect(campos(x, 'raro')['g:price']).toBe('8000.00 ARS');
		expect(campos(x, 'raro')['g:sale_price']).toBeUndefined();
	});

	it('envío gratis desde el mínimo de la tienda, foto absoluta y descripción por defecto', () => {
		const x = feedXml(
			tienda([
				producto({ id: 'caro', precio: 150_000, foto: '1573612664822-d7d347da7b80' }),
				producto({ id: 'barato', descripcion: '  ', foto: 'https://cdn.x.com/a.jpg?b=1&c=2' })
			]),
			'https://x.dev'
		);
		const caro = x.split('<item>').find((i) => i.includes('<g:id>caro<'))!;
		expect(caro).toContain('<g:shipping>');
		expect(caro).toContain('<g:country>AR</g:country>');
		expect(caro).toContain('<g:price>0.00 ARS</g:price>');
		expect(campos(x, 'caro')['g:image_link']).toMatch(
			/^https:\/\/images\.unsplash\.com\/photo-1573612664822-d7d347da7b80\?/
		);
		expect(x.split('<item>').find((i) => i.includes('<g:id>barato<'))).not.toContain('g:shipping');
		expect(campos(x, 'barato')['g:image_link']).toBe('https://cdn.x.com/a.jpg?b=1&amp;c=2');
		expect(campos(x, 'barato')['g:description']).toBe('Mate');
		expect(fotoFeed('javascript:alert(1)')).toBeNull();
		// Sin foto válida no se publica: Google y Meta rechazan el artículo.
		const sin = feedXml(tienda([producto({ id: 'sinfoto', foto: null })]), 'https://m.x');
		expect(sin).not.toContain('sinfoto');
	});

	it('una tienda sin productos es un feed vacío pero válido', () => {
		const x = feedXml(tienda([]), 'https://x.dev');
		expect(x).not.toContain('<item>');
		expect(x).toContain('</channel>\n</rss>');
	});
});
