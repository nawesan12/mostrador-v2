/**
 * El HTML de un correo, listo para mostrar: sin scripts, formularios, iframes ni manejadores, con
 * los links abriendo afuera y las imágenes remotas apagadas hasta que la persona las pide (avisan
 * a quien mandó que el correo se abrió y desde dónde). Las cid: se sirven desde /adjunto.
 * Igual se muestra en un iframe sin scripts: esto es la primera barrera, no la única.
 */
import sanitizeHtml from 'sanitize-html';

const PERMITIDAS = [
	...sanitizeHtml.defaults.allowedTags,
	'img',
	'span',
	'font',
	'center',
	'style',
	'table',
	'thead',
	'tbody',
	'tfoot',
	'tr',
	'td',
	'th',
	'colgroup',
	'col',
	'h1',
	'h2',
	'h3',
	'h4',
	'h5',
	'h6',
	'u',
	's',
	'del',
	'ins',
	'sub',
	'sup',
	'hr'
];

export function sanear(html: string, o: { cids: Map<string, string>; remotas: boolean }) {
	let remotas = false;
	const usados = new Set<string>();
	const out = sanitizeHtml(html, {
		allowedTags: PERMITIDAS,
		allowVulnerableTags: true, // <style>: el CSS del correo; queda adentro del iframe.
		allowedAttributes: {
			'*': [
				'style',
				'align',
				'valign',
				'width',
				'height',
				'bgcolor',
				'color',
				'dir',
				'class',
				'border',
				'cellpadding',
				'cellspacing',
				'colspan',
				'rowspan',
				'face',
				'size'
			],
			a: ['href', 'title', 'name', 'target', 'rel'],
			img: ['src', 'alt', 'title', 'width', 'height']
		},
		allowedSchemes: ['http', 'https', 'mailto', 'tel', 'cid', 'data'],
		allowedSchemesByTag: {
			img: ['http', 'https', 'cid', 'data'],
			a: ['http', 'https', 'mailto', 'tel']
		},
		transformTags: {
			a: (tag, attribs) => ({
				tagName: 'a',
				attribs: { ...attribs, target: '_blank', rel: 'noopener noreferrer nofollow' }
			}),
			img: (tag, attribs) => {
				const src = attribs.src ?? '';
				if (src.startsWith('cid:')) {
					const id = src.slice(4).replace(/^<|>$/g, '');
					usados.add(id);
					return { tagName: 'img', attribs: { ...attribs, src: o.cids.get(id) ?? '' } };
				}
				if (/^https?:/i.test(src) && !o.remotas) {
					remotas = true;
					return {
						tagName: 'img',
						attribs: {
							alt: attribs.alt ?? '',
							'data-remota': '1',
							width: attribs.width ?? '',
							height: attribs.height ?? ''
						}
					};
				}
				return { tagName: 'img', attribs };
			}
		}
	});
	// CSS que trae recursos de afuera (background:url(http…), @import) también delata la apertura.
	const limpio = out
		.replace(/@import[^;]*;?/gi, '')
		.replace(/url\(\s*['"]?https?:[^)]*\)/gi, (m) => (o.remotas ? m : ((remotas = true), 'none')));
	return { html: limpio, remotas, usados };
}
