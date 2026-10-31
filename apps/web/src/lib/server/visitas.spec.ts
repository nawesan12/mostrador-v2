import { describe, expect, it } from 'vitest';
import { esBot, esPrefetch, Medidor, paginaDe, POR_VISITANTE } from './visitas.ts';

const CHROME =
	'Mozilla/5.0 (Linux; Android 14; SM-A146M) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Mobile Safari/537.36';
const IPHONE =
	'Mozilla/5.0 (iPhone; CPU iPhone OS 18_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.0 Mobile/15E148 Safari/604.1';
const INSTAGRAM =
	'Mozilla/5.0 (iPhone; CPU iPhone OS 17_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 Instagram 340.0.0.22.109 (iPhone14,7; iOS 17_5; es_AR)';
const FACEBOOK_APP =
	'Mozilla/5.0 (Linux; Android 13; moto g32) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/128.0.6613.146 Mobile Safari/537.36 [FB_IAB/FB4A;FBAV/482.0.0.40.79;]';

describe('esBot', () => {
	it('deja pasar navegadores de personas, también los de las apps', () => {
		for (const ua of [CHROME, IPHONE, INSTAGRAM, FACEBOOK_APP]) expect(esBot(ua)).toBe(false);
	});

	it('descarta crawlers, vistas previas de links, Lighthouse y clientes de programa', () => {
		for (const ua of [
			'Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)',
			'Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm) Chrome/116.0.1938.76 Safari/537.36',
			'facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)',
			'meta-externalagent/1.1 (+https://developers.facebook.com/docs/sharing/webmasters/crawler)',
			'WhatsApp/2.23.20.0 A',
			'TelegramBot (like TwitterBot)',
			'Mozilla/5.0 (compatible; Discordbot/2.0; +https://discordapp.com)',
			'Slackbot-LinkExpanding 1.0 (+https://api.slack.com/robots)',
			'Mozilla/5.0 (Linux; Android 11; moto g power (2022)) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Mobile Safari/537.36 Chrome-Lighthouse',
			'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) HeadlessChrome/129.0.0.0 Safari/537.36',
			'Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; GPTBot/1.2; +https://openai.com/gptbot)',
			'Mozilla/5.0 (compatible; AhrefsBot/7.0; +http://ahrefs.com/robot/)',
			'curl/8.7.1',
			'python-requests/2.32.3',
			'Go-http-client/2.0',
			'node',
			'',
			null
		])
			expect(esBot(ua), String(ua)).toBe(true);
	});
});

describe('esPrefetch', () => {
	it('reconoce el prefetch y el prerender del navegador', () => {
		expect(esPrefetch(new Headers({ 'sec-purpose': 'prefetch' }))).toBe(true);
		expect(esPrefetch(new Headers({ 'sec-purpose': 'prefetch;prerender' }))).toBe(true);
		expect(esPrefetch(new Headers({ purpose: 'prefetch' }))).toBe(true);
		expect(esPrefetch(new Headers({ 'x-moz': 'prefetch' }))).toBe(true);
		expect(esPrefetch(new Headers({ accept: 'text/html' }))).toBe(false);
	});
});

describe('paginaDe', () => {
	it('mide inicio, catálogo, producto, carrito y checkout; nada más', () => {
		expect(paginaDe('/t/[tienda]')).toEqual({ pagina: 'inicio' });
		expect(paginaDe('/t/[tienda]/catalogo')).toEqual({ pagina: 'catalogo' });
		expect(paginaDe('/t/[tienda]/producto/[id]', 'mate-de-calabaza')).toEqual({
			pagina: 'producto',
			producto: 'mate-de-calabaza'
		});
		expect(paginaDe('/t/[tienda]/producto/[id]', '../admin')).toBeNull();
		expect(paginaDe('/t/[tienda]/carrito')).toEqual({ pagina: 'carrito' });
		expect(paginaDe('/t/[tienda]/checkout')).toEqual({ pagina: 'checkout' });
		expect(paginaDe('/t/[tienda]/pedido/[codigo]')).toBeNull();
		expect(paginaDe('/panel')).toBeNull();
		expect(paginaDe(null)).toBeNull();
	});
});

type Llamada = { url: string; init?: RequestInit };

function medidor(ahora: { t: number }, respuestas: Record<string, () => Response> = {}) {
	const llamadas: Llamada[] = [];
	const f = (async (url: string, init?: RequestInit) => {
		llamadas.push({ url, init });
		const r = Object.entries(respuestas).find(([k]) => url.endsWith(k));
		return r ? r[1]() : new Response(null, { status: 204 });
	}) as unknown as typeof fetch;
	const m = new Medidor({
		api: () => ({ url: 'http://api', clave: 'clave' }),
		f,
		ahora: () => ahora.t,
		cada: 60_000
	});
	const lotes = () =>
		llamadas
			.filter((l) => l.url.endsWith('/v1/interno/visitas'))
			.flatMap(
				(l) => (JSON.parse(String(l.init?.body)) as { visitas: Record<string, string>[] }).visitas
			);
	return { m, llamadas, lotes };
}

const vista = (o: Partial<Parameters<Medidor['anotar']>[0]> = {}) => ({
	tienda: 'mate',
	pagina: 'inicio' as const,
	ip: '181.1.2.3',
	agente: CHROME,
	sesion: null,
	...o
});

describe('Medidor', () => {
	it('el mismo visitante en el día da el mismo hash; otro día, otro (la sal rota y no se guarda)', () => {
		const ahora = { t: Date.parse('2026-10-30T15:00:00Z') };
		const { m } = medidor(ahora);
		const a = m.visitante('181.1.2.3', CHROME, 'mate');
		expect(a).toMatch(/^[0-9a-f]{32}$/);
		expect(m.visitante('181.1.2.3', CHROME, 'mate')).toBe(a);
		expect(m.visitante('181.1.2.3', CHROME, 'otra')).not.toBe(a);
		expect(m.visitante('181.1.2.4', CHROME, 'mate')).not.toBe(a);
		// 23:59 de Argentina sigue siendo el mismo día; a las 00:00 cambia.
		ahora.t = Date.parse('2026-10-31T02:59:00Z');
		expect(m.visitante('181.1.2.3', CHROME, 'mate')).toBe(a);
		ahora.t = Date.parse('2026-10-31T03:00:00Z');
		expect(m.visitante('181.1.2.3', CHROME, 'mate')).not.toBe(a);
	});

	it('manda por lotes, sin IP ni navegador, y con tope por visitante', async () => {
		const ahora = { t: Date.parse('2026-10-30T15:00:00Z') };
		const { m, lotes } = medidor(ahora);
		for (let i = 0; i < POR_VISITANTE + 30; i++) m.anotar(vista());
		m.anotar(vista({ pagina: 'producto', producto: 'mate-de-calabaza', ip: '181.9.9.9' }));
		await m.mandar();
		const vs = lotes();
		expect(vs).toHaveLength(POR_VISITANTE + 1);
		expect(JSON.stringify(vs)).not.toContain('181.');
		expect(JSON.stringify(vs)).not.toContain('Chrome');
		expect(vs.at(-1)).toMatchObject({
			tienda: 'mate',
			pagina: 'producto',
			producto: 'mate-de-calabaza'
		});
		// La hora siguiente vuelve a contar.
		ahora.t += 3_600_000;
		expect(m.anotar(vista())).toBe(true);
	});

	it('con sesión del panel manda la tienda de la sesión; si no se sabe, no cuenta', async () => {
		const ahora = { t: Date.parse('2026-10-30T15:00:00Z') };
		let caida = false;
		const { m, lotes } = medidor(ahora, {
			'/v1/auth/sesion': () =>
				caida
					? new Response(null, { status: 503 })
					: Response.json({ tiendaId: 'id-de-mate', rol: 'duenio' })
		});
		m.anotar(vista({ sesion: 'token-de-la-duena' }));
		m.anotar(vista({ ip: '181.5.5.5' }));
		await m.mandar();
		expect(lotes()).toEqual([
			expect.objectContaining({ duena: 'id-de-mate' }),
			expect.not.objectContaining({ duena: expect.anything() })
		]);
		caida = true;
		m.anotar(vista({ sesion: 'otro-token', ip: '181.6.6.6' }));
		await m.mandar();
		expect(lotes()).toHaveLength(2);
	});

	it('si la API no responde, se pierde el lote sin tirar', async () => {
		const m = new Medidor({
			api: () => ({ url: 'http://api', clave: 'clave' }),
			f: (async () => {
				throw new Error('caída');
			}) as unknown as typeof fetch,
			cada: 60_000
		});
		expect(m.anotar(vista())).toBe(true);
		await expect(m.mandar()).resolves.toBeUndefined();
	});

	it('sin la API real (demo) no anota nada', () => {
		const m = new Medidor({ api: () => null });
		expect(m.anotar(vista())).toBe(false);
	});
});
