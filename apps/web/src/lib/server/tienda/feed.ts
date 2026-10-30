/**
 * El feed de productos de la tienda (/t/{slug}/feed.xml): RSS 2.0 con el espacio de nombres `g:` de
 * Google Merchant Center. Meta Commerce Manager acepta el mismo archivo. Sólo lo publicado en la
 * tienda online (TiendaPublica ya viene filtrada), con precios en pesos.
 */
import type { ProductoPublico, TiendaPublica } from '#lib/tienda/tipos.ts';

/** Escapa para XML y saca los caracteres de control que XML 1.0 no admite. */
export function xml(s: string): string {
	return (
		s
			// eslint-disable-next-line no-control-regex -- justamente los que hay que sacar
			.replace(/[\u0000-\u0008\u000B\u000C\u000E-\u001F￾￿]/g, '')
			.replace(/&/g, '&amp;')
			.replace(/</g, '&lt;')
			.replace(/>/g, '&gt;')
			.replace(/"/g, '&quot;')
			.replace(/'/g, '&apos;')
	);
}

/** "12345.00 ARS" (los precios de la tienda son pesos enteros). */
export function precioFeed(pesos: number): string {
	return `${(Math.round(pesos * 100) / 100).toFixed(2)} ARS`;
}

/** La foto como URL absoluta (en la demo puede ser el id de una foto de Unsplash). */
export function fotoFeed(f: string | null): string | null {
	if (!f) return null;
	if (f.startsWith('https://')) return f;
	if (/^[\w-]+$/.test(f))
		return `https://images.unsplash.com/photo-${f}?auto=format&fit=crop&w=1000&q=80`;
	return null;
}

const corto = (s: string, max: number) => (s.length > max ? s.slice(0, max - 1) + '…' : s);

function item(p: ProductoPublico, t: TiendaPublica, base: string): string {
	const enOferta = p.anterior != null && p.anterior > p.precio;
	const foto = fotoFeed(p.foto);
	const gratis = t.envios.gratisDesde != null && p.precio >= t.envios.gratisDesde;
	const campos: [string, string][] = [
		['g:id', p.id],
		['g:title', corto(p.nombre, 150)],
		['g:description', corto(p.descripcion.trim() || p.nombre, 5000)],
		['g:link', `${base}/producto/${encodeURIComponent(p.slug ?? p.id)}`],
		['g:image_link', foto!],
		['g:availability', p.stock > 0 ? 'in_stock' : 'out_of_stock'],
		['g:price', precioFeed(enOferta ? p.anterior! : p.precio)],
		...(enOferta ? ([['g:sale_price', precioFeed(p.precio)]] as [string, string][]) : []),
		['g:brand', t.nombre],
		['g:condition', 'new'],
		// Sin GTIN ni código del fabricante: productos propios o de producción chica.
		['g:identifier_exists', 'no'],
		...(p.categoria ? ([['g:product_type', p.categoria]] as [string, string][]) : [])
	];
	const lineas = campos.map(([k, v]) => `      <${k}>${xml(v)}</${k}>`);
	if (gratis)
		lineas.push(
			'      <g:shipping>',
			'        <g:country>AR</g:country>',
			`        <g:price>${precioFeed(0)}</g:price>`,
			'      </g:shipping>'
		);
	return `    <item>\n${lineas.join('\n')}\n    </item>`;
}

/** origen: https://tulocalweb.com.ar (sin barra al final). */
export function feedXml(t: TiendaPublica, origen: string): string {
	const base = `${origen}/t/${t.slug}`;
	return [
		'<?xml version="1.0" encoding="UTF-8"?>',
		'<rss version="2.0" xmlns:g="http://base.google.com/ns/1.0">',
		'  <channel>',
		`    <title>${xml(t.nombre)}</title>`,
		`    <link>${xml(base)}</link>`,
		`    <description>${xml(`Productos de ${t.nombre}`)}</description>`,
		// Google y Meta rechazan un artículo sin imagen: los productos sin foto válida no se publican.
		...t.productos.filter((p) => fotoFeed(p.foto)).map((p) => item(p, t, base)),
		'  </channel>',
		'</rss>',
		''
	].join('\n');
}
