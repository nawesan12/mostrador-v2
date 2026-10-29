/**
 * Genera src/lib/legales.ts desde los 8 legales del handoff (design/Legal *.dc.html): encabezado,
 * "En pocas palabras", secciones numeradas, los bloques propios de cada página y el cierre.
 * El markup del handoff es generado y muy regular, así que alcanza con expresiones regulares.
 * Uso: bun scripts/legales.ts
 */
import { readFileSync, writeFileSync } from 'node:fs';
import { conMarca } from './marca.ts';
import { join } from 'node:path';

const __dirname = import.meta.dirname;
const DISENO = join(__dirname, '../../../design');

const PAGINAS = [
	['terminos', 'Legal Terminos'],
	['privacidad', 'Legal Privacidad'],
	['cookies', 'Legal Cookies'],
	['arrepentimiento', 'Legal Arrepentimiento'],
	['defensa-del-consumidor', 'Legal Consumidor'],
	['tratamiento-de-datos', 'Legal Tratamiento de datos'],
	['uso-aceptable', 'Legal Uso aceptable'],
	['reembolsos', 'Legal Reembolsos']
] as const;

/** Links del handoff (archivos .dc.html) → rutas del sitio. */
const RUTAS: Record<string, string> = {
	'Mostrador%20Contacto.dc.html': '/contacto',
	'Mostrador%20Seguridad.dc.html': '/seguridad',
	'Mostrador%20Estado.dc.html': '/estado',
	'Mostrador%20Ayuda.dc.html': '/ayuda',
	'Mostrador%20Ingresar.dc.html': '/ingresar',
	'Onboarding%20Mostrador.dc.html': '/ingresar?m=registro',
	'Mostrador%20v4.dc.html': '/'
};
for (const [slug, archivo] of PAGINAS) RUTAS[encodeURI(archivo) + '.dc.html'] = `/legales/${slug}`;

const ruta = (href: string) => {
	if (/^https?:|^mailto:|^#/.test(href)) return href;
	const [f, ancla] = href.split('#');
	const r = RUTAS[f];
	if (!r) throw new Error(`Link sin ruta: ${href}`);
	return ancla ? `${r}#${ancla}` : r;
};

/** Deja sólo <b> y <a href> (con la ruta del sitio); el resto del markup del handoff se va. */
function limpiar(html: string): string {
	return html
		.replace(/<a href="([^"]*)"[^>]*>/g, (_, h: string) => `<a href="${ruta(h)}">`)
		.replace(/<b [^>]*>/g, '<b>')
		.replace(/<(?!\/?(?:b|a)\b)[^>]+>/g, '')
		.replace(/\s+/g, ' ')
		.trim();
}

const texto = (html: string) => limpiar(html).replace(/<[^>]+>/g, '');
/** 'ph:bank-fill' → 'phBankFill', como los exports de #lib/icons.ts. */
const icono = (s: string) =>
	s.replace(/^ph:/, 'ph-').replace(/-(\w)/g, (_, c: string) => c.toUpperCase());

const vinetas = (ul: string) =>
	[...ul.matchAll(/<li[^>]*><span[^>]*><\/span><span>([\s\S]*?)<\/span><\/li>/g)].map((m) =>
		limpiar(m[1])
	);

type Bloque =
	| { t: 'resumen'; items: string[] }
	| {
			t: 'seccion';
			id: string;
			n: string;
			titulo: string;
			cuerpo: ({ p: string } | { ul: string[] })[];
	  }
	| { t: 'organismos'; items: { href: string; icono: string; titulo: string; texto: string }[] }
	| { t: 'creditos'; items: { etiqueta: string; umbral: string; credito: string; pie: string }[] }
	| {
			t: 'cookies';
			filas: { nombre: string; tipo: string; color: string; para: string; dura: string }[];
	  }
	| { t: 'configurar-cookies' }
	| { t: 'arrepentimiento' };

function pagina(slug: string, archivo: string) {
	const s = readFileSync(join(DISENO, `${archivo}.dc.html`), 'utf8');
	const desc = s.match(/<meta name="description" content="([^"]*)"/)![1];
	const hero = s.slice(s.indexOf('<main'), s.indexOf('<article'));
	const titulo = texto(hero.match(/<h1[^>]*>([\s\S]*?)<\/h1>/)![1]);
	const miga = texto([...hero.matchAll(/<span style="color:#2A140C">([^<]*)<\/span>/g)][0][1]);
	const chips = [...hero.matchAll(/<span style="padding:6px 11px[^>]*>([^<]*)<\/span>/g)].map(
		(m) => m[1]
	);
	const ilustracion = hero.match(/assets\/hornero\/([\w-]+)\.png/)![1];
	// Nombre corto de la página en la lista lateral "Legales" (no siempre es el título).
	const corto = [
		...hero.matchAll(/<a href="Legal%20[^"]*" style="padding:9px 12px[^>]*>([^<]*)<\/a>/g)
	].map((m) => m[1])[PAGINAS.findIndex(([sl]) => sl === slug)];
	const indice = [...hero.matchAll(/<a href="#(s\d+)"[^>]*>([^<]*)<\/a>/g)].map((m) => [
		m[1],
		m[2]
	]);

	const art = s.slice(s.indexOf('<article'), s.indexOf('</article>'));
	const bloques: [number, Bloque][] = [];

	const resumen = art.match(/En pocas palabras<\/div>(<ul[\s\S]*?<\/ul>)/);
	if (resumen)
		bloques.push([art.indexOf('En pocas palabras'), { t: 'resumen', items: vinetas(resumen[1]) }]);

	for (const m of art.matchAll(/<section id="(s\d+)"[^>]*>([\s\S]*?)<\/section>/g)) {
		const [, id, cuerpo] = m;
		const h2 = cuerpo.match(/<h2[^>]*><span[^>]*>(\d+)<\/span>([\s\S]*?)<\/h2>/)!;
		const partes = [...cuerpo.matchAll(/<p [^>]*>([\s\S]*?)<\/p>|(<ul [\s\S]*?<\/ul>)/g)].map(
			(p) => (p[1] !== undefined ? { p: limpiar(p[1]) } : { ul: vinetas(p[2]) })
		);
		bloques.push([m.index!, { t: 'seccion', id, n: h2[1], titulo: texto(h2[2]), cuerpo: partes }]);
	}

	const org = art.indexOf('repeat(auto-fit,minmax(200px,1fr));gap:8px');
	if (org >= 0) {
		const zona = art.slice(org, art.indexOf('<section', org));
		const items = [
			...zona.matchAll(
				/<a href="([^"]*)"[^>]*><iconify-icon icon="([^"]*)"[^>]*><\/iconify-icon><b[^>]*>([^<]*)<\/b><span[^>]*>([^<]*)<\/span>/g
			)
		].map((m) => ({ href: ruta(m[1]), icono: icono(m[2]), titulo: m[3], texto: m[4] }));
		bloques.push([org, { t: 'organismos', items }]);
	}

	const cred = art.indexOf('repeat(auto-fit,minmax(150px,1fr));gap:8px');
	if (cred >= 0) {
		const zona = art.slice(cred, art.indexOf('<section', cred));
		const items = [
			...zona.matchAll(
				/<div style="padding:16px[^>]*><div[^>]*>([^<]*)<\/div><div[^>]*>([^<]*)<\/div><div[^>]*>([^<]*)<\/div><div[^>]*>([^<]*)<\/div><\/div>/g
			)
		].map((m) => ({ etiqueta: m[1], umbral: m[2], credito: m[3], pie: m[4] }));
		bloques.push([cred, { t: 'creditos', items }]);
	}

	const tabla = art.indexOf('minmax(0,1.3fr) minmax(0,.8fr)');
	if (tabla >= 0) {
		const filas = [
			...art.matchAll(
				/<div style="display:grid;grid-template-columns:minmax\(0,1\.3fr\)[^"]*border-top[^>]*><span[^>]*>([^<]*)<\/span><span[^>]*color:(#\w+)[^>]*>([^<]*)<\/span><span[^>]*>([^<]*)<\/span><span[^>]*>([^<]*)<\/span><\/div>/g
			)
		].map((m) => ({ nombre: m[1], color: m[2], tipo: m[3], para: m[4], dura: m[5] }));
		bloques.push([tabla, { t: 'cookies', filas }]);
	}
	const cfg = art.indexOf('Configurar mis cookies');
	if (cfg >= 0) bloques.push([cfg, { t: 'configurar-cookies' }]);
	const arr = art.indexOf('Revocar mi contratación');
	if (arr >= 0) bloques.push([arr, { t: 'arrepentimiento' }]);

	const cierre = art.match(/<b style="font-size:15\.5px">([^<]*)<\/b><div[^>]*>([^<]*)<\/div>/)!;
	return {
		slug,
		titulo,
		corto,
		miga,
		descripcion: desc,
		chips,
		ilustracion,
		indice,
		bloques: bloques.sort((a, b) => a[0] - b[0]).map((b) => b[1]),
		cierre: { titulo: cierre[1], texto: cierre[2] }
	};
}

const datos = PAGINAS.map(([slug, archivo]) => pagina(slug, archivo));
const iconos = [
	...new Set(
		datos.flatMap((d) =>
			d.bloques.flatMap((b) => (b.t === 'organismos' ? b.items.map((i) => i.icono) : []))
		)
	)
];

const salida = `// Generado por scripts/legales.ts desde design/Legal *.dc.html. No editar a mano.
import { ${iconos.join(', ')} } from '#lib/icons.ts';
import type { IconData } from '#lib/icons.ts';

export type BloqueLegal =
	| { t: 'resumen'; items: string[] }
	| { t: 'seccion'; id: string; n: string; titulo: string; cuerpo: ({ p: string } | { ul: string[] })[] }
	| { t: 'organismos'; items: { href: string; icono: IconData; titulo: string; texto: string }[] }
	| { t: 'creditos'; items: { etiqueta: string; umbral: string; credito: string; pie: string }[] }
	| { t: 'cookies'; filas: { nombre: string; tipo: string; color: string; para: string; dura: string }[] }
	| { t: 'configurar-cookies' }
	| { t: 'arrepentimiento' };

export type Legal = {
	slug: string;
	titulo: string;
	/** Cómo figura en la lista lateral "Legales". */
	corto: string;
	miga: string;
	descripcion: string;
	chips: string[];
	ilustracion: string;
	indice: [string, string][];
	bloques: BloqueLegal[];
	cierre: { titulo: string; texto: string };
};

export const LEGALES: Legal[] = ${JSON.stringify(datos, null, '\t').replace(/"icono": "(\w+)"/g, '"icono": $1')};
`;
writeFileSync(join(__dirname, '../src/lib/legales.ts'), conMarca(salida));
console.log(
	datos
		.map(
			(d) => `${d.slug}: ${d.bloques.map((b) => b.t[0]).join('')} (${d.indice.length} en el índice)`
		)
		.join('\n')
);
