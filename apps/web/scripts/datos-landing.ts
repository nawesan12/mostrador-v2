/**
 * Genera src/lib/landing/datos.ts con los datos de la landing que están en el script del handoff
 * (design/Mostrador v4.dc.html): feed en vivo, integraciones, categorías, planes, comparativa con la
 * competencia y el chat de Hornero. Uso: bun scripts/datos-landing.ts
 */
import { readFileSync, writeFileSync } from 'node:fs';
import { conMarca } from './marca.ts';
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

// El muro del handoff trae 70 logos. Quedan las que tienen API y funcionan en Argentina; salen las
// que no se pueden conectar (pedido de Nahuel, 22/10): marcas de tarjeta y bancos (llegan por Mercado
// Pago, Payway o Getnet), Stripe y PayPal (no cobran en pesos acá), TikTok Shop y YouTube Shopping (no
// están en Argentina), Pinterest (sin catálogo para Argentina), Lemon y Cuenta DNI (sin API para
// comercios), Uber Flash (es para personas), Cruz del Sur y Via Cargo (sin API pública).
// Segunda vuelta (28/10): Naranja X y MODO (sin API para comercios), PedidosYa y Rappi (sin API
// abierta), FedEx y DHL (internacionales), Binance Pay (cripto), Messenger (va con Facebook), Excel
// (ya está Sheets), Gemini (su app no admite conectores), Andreani, OCA, Moova y PedidosYa Envíos
// (piden contrato o volumen), y Tango, HubSpot, Klaviyo, Notion, Slack y Gmail (licencia paga,
// auditoría paga o fuera del comercio chico).
const SIN_INTEGRACION = [
	'Naranja X',
	'MODO',
	'PedidosYa',
	'Rappi',
	'FedEx',
	'DHL',
	'Binance Pay',
	'Messenger',
	'Excel',
	'Gemini',
	'Andreani',
	'OCA',
	'Moova',
	'PedidosYa Envíos',
	'Tango',
	'HubSpot',
	'Klaviyo',
	'Notion',
	'Slack',
	'Gmail',
	'Visa',
	'Mastercard',
	'American Express',
	'Banco Galicia',
	'Santander',
	'BBVA',
	'Banco Nación',
	'Stripe',
	'PayPal',
	'TikTok Shop',
	'YouTube',
	'Pinterest',
	'Lemon',
	'Cuenta DNI',
	'Uber Flash',
	'Cruz del Sur',
	'Via Cargo',
	// 29/10: tampoco las que no vamos a conectar por ahora.
	'Ualá',
	'Getnet',
	'Enviopack',
	'Mailchimp',
	'Zapier',
	'Colppy',
	'Xubio'
];
out.INTEG = [
	...(out.INTEG as unknown[][]).filter((i) => !SIN_INTEGRACION.includes(i[0] as string)),
	[
		'Empretienda',
		'canales',
		{ w: 'Empretienda', abbr: 'ET', bg: '#2E2A6B', fg: '#fff' },
		'Importá tu tienda'
	]
];
// Instagram y Facebook Shopping no existen en Argentina desde 8/2023: quedan los mensajes. Las
// migraciones no son de un clic.
const TEXTOS: Record<string, string> = {
	Instagram: 'Mensajes en la bandeja',
	Facebook: 'Mensajes en la bandeja',
	Tiendanube: 'Importá tu tienda'
};
// Agrupadas por categoría, en el orden del handoff: así las migraciones quedan junto a los canales.
const ORDEN = [...new Set((out.INTEG as unknown[][]).map((i) => i[1] as string))];
out.INTEG = (out.INTEG as unknown[][])
	.map((i) => (TEXTOS[i[0] as string] ? [i[0], i[1], i[2], TEXTOS[i[0] as string]] : i))
	.sort((a, b) => ORDEN.indexOf(a[1] as string) - ORDEN.indexOf(b[1] as string));
// Andreani salió de las integraciones: el feed despacha por Correo Argentino.
const aCorreo = (v: unknown): unknown =>
	Array.isArray(v)
		? v.map(aCorreo)
		: v && typeof v === 'object'
			? Object.fromEntries(Object.entries(v).map(([k, x]) => [k, aCorreo(x)]))
			: v === 'andreani'
				? 'correo'
				: v;
out.FEED = aCorreo(out.FEED);

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
		// En WebP, que pesa menos (static/caja tiene las dos).
		return v.replace('https://caja.iamnaw.dev/mascota/', '/caja/').replace(/\.png$/, '.webp');
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
writeFileSync(join(import.meta.dirname, '../src/lib/landing/datos.ts'), conMarca(ts));
console.log(
	NOMBRES.map(
		(n) =>
			`${n}: ${Array.isArray(out[n]) ? (out[n] as unknown[]).length : Object.keys(out[n] as object).length}`
	).join(' · ')
);
