/**
 * Genera src/lib/landing/datos.ts con los datos de la landing que están en el script del handoff
 * (design/Mostrador v4.dc.html): feed en vivo, integraciones, categorías, planes, comparativa con la
 * competencia y el chat de Hornero. Uso: bun scripts/datos-landing.ts
 */
import { readFileSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';
import { runInNewContext } from 'node:vm';

const html = readFileSync(
	join(import.meta.dirname, '../../../design/Mostrador v4.dc.html'),
	'utf8'
);
const desde = html.indexOf('const fmt = ');
const hasta = html.indexOf('class Component extends DCLogic');
const codigo = html.slice(desde, hasta);
const ctx: Record<string, unknown> = { window: { MOSTRADOR_TEMPLATES: [], MOSTRADOR_RUBROS: [] } };
const NOMBRES = [
	'FEED',
	'INTEG',
	'CATS',
	'LPLANS',
	'CMP',
	'CHAT',
	'SERIES',
	'NOTIFS',
	'ORDERS',
	'OR',
	'AREAS',
	'SIMPLE',
	'KICONS'
];
runInNewContext(codigo + `\n;globalThis.__out = { ${NOMBRES.join(', ')} };`, ctx);
const out = ctx.__out as Record<string, unknown>;

// Logos que el handoff pide y no existen en Iconify (en el diseño se ven vacíos): reemplazos.
const SUSTITUTOS: Record<string, Record<string, string>> = {
	'logos:google-sheets': { icon: 'simple-icons:googlesheets', col: '#34A853' },
	'logos:microsoft-excel': { icon: 'simple-icons:microsoftexcel', col: '#217346' },
	// Sin logo en ningún set: sigla en recuadro, como ML, ARCA u OCA en el handoff.
	'simple-icons:klaviyo': { w: 'Klaviyo', abbr: 'K', bg: '#232426', fg: '#fff' }
};
const usados = new Set<string>();
function resolver(v: unknown): unknown {
	if (Array.isArray(v)) return v.map(resolver);
	if (v && typeof v === 'object') {
		const o = { ...(v as Record<string, unknown>) };
		if (typeof o.icon === 'string' && SUSTITUTOS[o.icon]) {
			const sust = SUSTITUTOS[o.icon];
			delete o.icon;
			Object.assign(o, sust);
		}
		for (const k of Object.keys(o)) o[k] = resolver(o[k]);
		return o;
	}
	// La mascota de Caja se sirve desde nuestro static (el README del handoff pide no depender de caja.iamnaw.dev).
	if (typeof v === 'string' && v.startsWith('https://caja.iamnaw.dev/mascota/'))
		return v.replace('https://caja.iamnaw.dev/mascota/', '/caja/');
	if (typeof v === 'string' && /^(ph|logos|simple-icons):[a-z0-9-]+$/.test(v)) {
		usados.add(v);
		return `@@${v}@@`;
	}
	return v;
}
const camel = (n: string) =>
	n.replace(/^simple-icons:/, 'si-').replace(/[:-]+([a-z0-9])/g, (_, c: string) => c.toUpperCase());
for (const n of NOMBRES) out[n] = resolver(out[n]);

const ts = (
	`// Generado por scripts/datos-landing.ts desde design/Mostrador v4.dc.html — no editar a mano.\n` +
	`// Los íconos son referencias a #lib/icons.ts: el bundle lleva sólo los que se usan.\n` +
	`import * as I from '#lib/icons.ts';\n\n` +
	NOMBRES.map((n) => `export const ${n} = ${JSON.stringify(out[n], null, '\t')} as const;`).join(
		'\n\n'
	) +
	'\n'
).replace(/"@@([^@]+)@@"/g, (_, n: string) => `I.${camel(n)}`);
writeFileSync(join(import.meta.dirname, '../src/lib/landing/datos.ts'), ts);
console.log(
	NOMBRES.map(
		(n) =>
			`${n}: ${Array.isArray(out[n]) ? (out[n] as unknown[]).length : Object.keys(out[n] as object).length}`
	).join(' · ')
);
