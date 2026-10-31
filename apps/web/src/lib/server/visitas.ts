/**
 * Visitas de la tienda online, medidas por Tu Local Web sin cookies ni datos personales
 * (decisiones §36). Sin scripts de terceros: cuenta este servidor.
 *
 * - Una vista por página: la carga completa la cuenta `hooks.server.ts` al responder; la navegación
 *   dentro de la tienda (que no pide HTML) la avisa el layout de /t/[tienda] a /t/<slug>/vista.
 * - Visitante del día: sha256(sal del día + IP + navegador + tienda), recortado a 16 bytes. La sal
 *   es al azar, vive sólo en la memoria de este proceso y cambia cada día de Argentina: no se guarda
 *   en ningún lado, así que nadie (ni nosotros) puede volver del hash a la persona, ni unir días.
 * - No cuentan: bots y vistas previas (WhatsApp, Facebook, Telegram…), Lighthouse, prefetch y
 *   prerender, y la dueña o su equipo con la sesión del panel en su propia tienda.
 * - Tope: 120 vistas por visitante y 600 por IP por hora; lo que pase, no suma.
 * - En segundo plano: se juntan y se mandan por lotes a la API Go cada 5 s. Si la API no responde,
 *   el lote se pierde (son estadísticas): la tienda nunca espera ni se cae por esto.
 */
import { createHash, randomBytes } from 'node:crypto';

export type PaginaMedida = 'inicio' | 'catalogo' | 'producto' | 'carrito' | 'checkout';

/** Crawlers, vistas previas de links, medidores y clientes HTTP de programa. */
const BOTS = new RegExp(
	[
		'bot\\b',
		'bot/',
		'crawl',
		'spider',
		'slurp',
		'facebookexternalhit',
		'facebookcatalog',
		'meta-externalagent',
		'whatsapp',
		'telegram',
		'twitterbot',
		'linkedin',
		'pinterest',
		'discord',
		'slack',
		'skypeuripreview',
		'embedly',
		'preview',
		'vkshare',
		'redditbot',
		'applebot',
		'lighthouse',
		'pagespeed',
		'gtmetrix',
		'headless',
		'phantomjs',
		'puppeteer',
		'playwright',
		'selenium',
		'curl/',
		'wget',
		'python',
		'go-http-client',
		'axios',
		'node-fetch',
		'undici',
		'okhttp',
		'java/',
		'libwww',
		'httpclient',
		'postman',
		'insomnia',
		'uptime',
		'pingdom',
		'statuscake',
		'monitor',
		'ahrefs',
		'semrush',
		'yandex',
		'baidu',
		'bytespider',
		'petalbot',
		'gptbot',
		'chatgpt',
		'claudebot',
		'anthropic',
		'perplexity',
		'ccbot',
		'google-inspectiontool',
		'googleother',
		'feedfetcher',
		'mediapartners',
		'adsbot',
		'bingpreview'
	].join('|'),
	'i'
);

/** Un navegador de persona: con user agent de navegador y sin marcas de bot. */
export function esBot(agente: string | null | undefined): boolean {
	const ua = (agente ?? '').trim();
	if (ua.length < 20 || !/^Mozilla\//.test(ua)) return true;
	return BOTS.test(ua);
}

/** Prefetch o prerender del navegador (o de un proxy): todavía no es una persona mirando. */
export function esPrefetch(h: Headers): boolean {
	const p = [h.get('purpose'), h.get('sec-purpose'), h.get('x-purpose'), h.get('x-moz')]
		.filter(Boolean)
		.join(' ')
		.toLowerCase();
	return /prefetch|prerender|preview/.test(p);
}

/** La página medida de una ruta de /t/[tienda] (null: no se mide, como el pedido o los turnos). */
export function paginaDe(
	ruta: string | null | undefined,
	producto?: string
): { pagina: PaginaMedida; producto?: string } | null {
	switch (ruta) {
		case '/t/[tienda]':
			return { pagina: 'inicio' };
		case '/t/[tienda]/catalogo':
			return { pagina: 'catalogo' };
		case '/t/[tienda]/carrito':
			return { pagina: 'carrito' };
		case '/t/[tienda]/checkout':
			return { pagina: 'checkout' };
		case '/t/[tienda]/producto/[id]':
			return producto && /^[a-z0-9-]{1,60}$/.test(producto)
				? { pagina: 'producto', producto }
				: null;
	}
	return null;
}

const HORA = 3_600_000;
/** Día de Argentina (UTC−3, sin horario de verano). */
const diaAr = (t: number) => new Date(t - 3 * HORA).toISOString().slice(0, 10);

export type Vista = {
	tienda: string;
	pagina: PaginaMedida;
	producto?: string;
	visitante: string;
	duena?: string;
};

type Pendiente = Omit<Vista, 'duena'> & { sesion: string | null };

export type Opciones = {
	api: () => { url: string; clave: string } | null;
	f?: typeof fetch;
	ahora?: () => number;
	/** Cada cuánto se manda el lote (ms). */
	cada?: number;
};

export const POR_VISITANTE = 120;
export const POR_IP = 600;
const MAX_COLA = 10_000;
const MAX_LOTE = 1_000;
const MAX_CONTADORES = 200_000;

export class Medidor {
	private sal = { dia: '', valor: Buffer.alloc(0) };
	private ventana = -1;
	private porVisitante = new Map<string, number>();
	private porIp = new Map<string, number>();
	private cola: Pendiente[] = [];
	private timer: ReturnType<typeof setTimeout> | null = null;
	private mandando: Promise<void> | null = null;
	/** Sesión del panel (hash del token) → tienda de esa sesión ('' si no es válida). */
	private sesiones = new Map<string, { tienda: string; hasta: number }>();

	constructor(private o: Opciones) {}

	private ahora() {
		return this.o.ahora?.() ?? Date.now();
	}

	/** El visitante del día: no se puede volver a la persona sin la sal, que no se guarda. */
	visitante(ip: string, agente: string, tienda: string): string {
		const hoy = diaAr(this.ahora());
		if (this.sal.dia !== hoy) this.sal = { dia: hoy, valor: randomBytes(32) };
		return createHash('sha256')
			.update(this.sal.valor)
			.update(`\0${ip}\0${agente}\0${tienda}`)
			.digest()
			.subarray(0, 16)
			.toString('hex');
	}

	/** Suma un intento y dice si sigue dentro del tope de la hora. */
	private permitido(visitante: string, ip: string): boolean {
		const v = Math.floor(this.ahora() / HORA);
		if (v !== this.ventana) {
			this.ventana = v;
			this.porVisitante.clear();
			this.porIp.clear();
		}
		const ipH = createHash('sha256').update(this.sal.valor).update(ip).digest('hex').slice(0, 16);
		const a = (this.porVisitante.get(visitante) ?? 0) + 1;
		const b = (this.porIp.get(ipH) ?? 0) + 1;
		// Con los mapas llenos (un ataque), no se cuentan visitantes nuevos hasta la próxima hora.
		if (
			(a === 1 && this.porVisitante.size >= MAX_CONTADORES) ||
			(b === 1 && this.porIp.size >= MAX_CONTADORES)
		)
			return false;
		this.porVisitante.set(visitante, a);
		this.porIp.set(ipH, b);
		return a <= POR_VISITANTE && b <= POR_IP;
	}

	/**
	 * Anota una vista (ya filtrada: página medida, persona, sin prefetch). Nunca tira ni espera.
	 * `sesion`: el token del panel de quien mira, si tiene; se usa sólo para saber si es la dueña.
	 */
	anotar(v: {
		tienda: string;
		pagina: PaginaMedida;
		producto?: string;
		ip: string;
		agente: string;
		sesion: string | null;
	}): boolean {
		try {
			if (!this.o.api()) return false;
			const visitante = this.visitante(v.ip, v.agente, v.tienda);
			if (!this.permitido(visitante, v.ip) || this.cola.length >= MAX_COLA) return false;
			this.cola.push({
				tienda: v.tienda,
				pagina: v.pagina,
				...(v.producto ? { producto: v.producto } : {}),
				visitante,
				sesion: v.sesion
			});
			this.programar();
			return true;
		} catch {
			return false;
		}
	}

	private programar() {
		if (this.timer) return;
		this.timer = setTimeout(() => {
			this.timer = null;
			void this.mandar();
		}, this.o.cada ?? 5_000);
		// No mantiene vivo el proceso (ni los tests).
		(this.timer as { unref?: () => void }).unref?.();
	}

	/** Manda lo que haya en la cola (de a lotes). Los errores se tragan: son estadísticas. */
	async mandar(): Promise<void> {
		if (this.mandando) return this.mandando;
		this.mandando = (async () => {
			while (this.cola.length) {
				const lote = this.cola.splice(0, MAX_LOTE);
				await this.mandarLote(lote).catch(() => {});
			}
		})().finally(() => (this.mandando = null));
		return this.mandando;
	}

	private async tiendaDeSesion(token: string, api: { url: string; clave: string }) {
		const clave = createHash('sha256').update(token).digest('hex');
		const ya = this.sesiones.get(clave);
		if (ya && ya.hasta > this.ahora()) return ya.tienda;
		const f = this.o.f ?? fetch;
		const r = await f(`${api.url}/v1/auth/sesion`, {
			headers: { authorization: `Bearer ${token}`, 'x-mostrador-bff': api.clave },
			signal: AbortSignal.timeout(2_000)
		});
		let tienda = '';
		if (r.ok) tienda = String(((await r.json()) as { tiendaId?: string }).tiendaId ?? '');
		else if (r.status !== 401) throw new Error(`sesión ${r.status}`);
		if (this.sesiones.size > 5_000) this.sesiones.clear();
		this.sesiones.set(clave, { tienda, hasta: this.ahora() + 10 * 60_000 });
		return tienda;
	}

	private async mandarLote(lote: Pendiente[]) {
		const api = this.o.api();
		if (!api) return;
		// La tienda de cada sesión del panel: si no se sabe (la API no respondió), la vista no
		// cuenta, por las dudas de que sea la dueña.
		const tokens = [...new Set(lote.map((v) => v.sesion).filter((t): t is string => !!t))];
		const duenas = new Map<string, string | null>();
		await Promise.all(
			tokens.map((t) =>
				this.tiendaDeSesion(t, api).then(
					(x) => duenas.set(t, x),
					() => duenas.set(t, null)
				)
			)
		);
		const visitas: Vista[] = [];
		for (const { sesion, ...v } of lote) {
			if (!sesion) visitas.push(v);
			else {
				const d = duenas.get(sesion);
				if (d === null || d === undefined) continue;
				visitas.push(d ? { ...v, duena: d } : v);
			}
		}
		if (!visitas.length) return;
		const f = this.o.f ?? fetch;
		await f(`${api.url}/v1/interno/visitas`, {
			method: 'POST',
			headers: { 'content-type': 'application/json', 'x-mostrador-bff': api.clave },
			body: JSON.stringify({ visitas }),
			signal: AbortSignal.timeout(3_000)
		});
	}
}
