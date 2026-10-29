/**
 * Tincho en Telegram (@mostradorhornero_bot): la dueña vincula su tienda con un código de un solo
 * uso (Integraciones › Telegram › Vincular), le pregunta en lenguaje natural y recibe un aviso por
 * cada venta. En la demo vive sobre el sandbox; con la API real (DATA_SOURCE=api), el vínculo y el
 * contexto de la tienda salen de Go (`TelegramApi`) y el aviso de cada venta lo manda el worker de Go.
 * El webhook es uno solo y las respuestas con IA salen siempre del Worker, con los mismos topes.
 *
 * Seguridad: Telegram manda el secreto del webhook en X-Telegram-Bot-Api-Secret-Token (se compara en
 * tiempo constante) y el chat sólo ve la tienda a la que se vinculó.
 */
import { env } from 'cloudflare:workers';
import { sistema, type DatosHornero } from './hornero/contexto.ts';
import { usaSandbox } from './datos/index.ts';
import { cupoDemo, diaAr, POR_VISITANTE } from './hornero/cupo.ts';
import { responderConIa, type Mensaje } from './hornero/modelo.ts';
import { responderSimulado } from './hornero/simulado.ts';
import {
	borrarCambio,
	buscarCambio,
	cambiosDe,
	DatosSandbox,
	dbSandbox,
	guardarCambio,
	sumarUsoHornero,
	type DbSandbox
} from './sandbox/sandbox.ts';
import { TelegramApi } from './telegram-api.ts';

export const BOT = 'mostradorhornero_bot';
/** Cuánto dura el código de vinculación. */
const VIGENCIA = 10 * 60_000;

type Vinculo = { chat: number; nombre: string; en: number };

export const habilitado = () => !!env.TELEGRAM_BOT_TOKEN && !!env.TELEGRAM_WEBHOOK_SECRET;

/** Compara en tiempo constante. */
export function secretoValido(recibido: string | null) {
	const esperado = String(env.TELEGRAM_WEBHOOK_SECRET ?? '');
	if (!esperado || !recibido || recibido.length !== esperado.length) return false;
	let d = 0;
	for (let i = 0; i < esperado.length; i++) d |= esperado.charCodeAt(i) ^ recibido.charCodeAt(i);
	return d === 0;
}

export async function enviar(chat: number, texto: string) {
	// En las pruebas el token es de mentira: no se sale a Telegram.
	if (env.TELEGRAM_BOT_TOKEN === 'prueba') return true;
	const r = await fetch(`https://api.telegram.org/bot${env.TELEGRAM_BOT_TOKEN}/sendMessage`, {
		method: 'POST',
		headers: { 'content-type': 'application/json' },
		body: JSON.stringify({ chat_id: chat, text: texto.slice(0, 4000) })
	}).catch(() => null);
	return !!r?.ok;
}

const ALFABETO = '23456789ABCDEFGHJKLMNPQRSTUVWXYZ';
const codigo = () =>
	[...crypto.getRandomValues(new Uint8Array(8))].map((b) => ALFABETO[b % ALFABETO.length]).join('');

const linkDe = (cod: string) => `https://t.me/${BOT}?start=${cod}`;

/** Un código de un solo uso para vincular el chat con esta tienda; devuelve el link al bot. */
export async function linkVinculo(db: DbSandbox, sandboxId: string) {
	const cod = codigo();
	await guardarCambio(db, sandboxId, 'tg-codigo', cod, { vence: Date.now() + VIGENCIA });
	return linkDe(cod);
}

/** Go, desde el webhook o el panel. La IP no aplica: quien llama es Telegram o la sesión. */
const api = () =>
	new TelegramApi(String(env.API_URL), String(env.CLAVE_BFF ?? ''), { ip: '', agente: 'telegram' });

/** Con la API: el código lo genera Go (dueña o admin) y se guarda sólo su hash. */
export async function linkVinculoApi(sesion: string) {
	return linkDe((await api().codigo(sesion)).codigo);
}

/** Con la API: el chat vinculado a la tienda de la sesión (null si no hay). */
export async function vinculoApi(sesion: string) {
	const e = await api().estado(sesion);
	return e.vinculado ? { nombre: e.nombre ?? 'Telegram' } : null;
}

/** Con la API: desvincula desde el panel y le avisa al chat que estaba. */
export async function desvincularApi(sesion: string) {
	const chat = await api().desvincular(sesion);
	if (chat) await enviar(chat, DESVINCULADO);
}

export async function vinculoDe(db: DbSandbox, sandboxId: string) {
	const c = await cambiosDe(db, sandboxId, 'cuenta');
	return (c.get('telegram') as Vinculo | null | undefined) ?? null;
}

const DESVINCULADO =
	'Listo, desvinculé este chat de la tienda. Para volver, tocá Vincular en Integraciones.';

export async function desvincular(db: DbSandbox, sandboxId: string) {
	const v = await vinculoDe(db, sandboxId);
	if (!v) return;
	await borrarCambio(db, sandboxId, 'tg-chat', String(v.chat));
	await guardarCambio(db, sandboxId, 'cuenta', 'telegram', null);
	await enviar(v.chat, DESVINCULADO);
}

/** Aviso de venta al chat vinculado (si hay). */
export async function avisarVenta(db: DbSandbox, sandboxId: string, texto: string) {
	if (!habilitado()) return;
	const v = await vinculoDe(db, sandboxId);
	if (v) await enviar(v.chat, texto);
}

type Update = {
	message?: {
		chat: { id: number; type: string };
		from?: { first_name?: string };
		text?: string;
	};
};

const AYUDA =
	'Soy Tincho, de Tu Local Web. Preguntame lo que quieras de tu tienda: "¿qué se vendió hoy?", "¿qué tengo para preparar?", "¿qué me falta?". Y te aviso cada venta.';

const SIN_VINCULO =
	'Todavía no sé de qué tienda sos. Vinculá este chat desde el panel: Integraciones › Telegram › Vincular.';
const VENCIDO = 'Ese link venció o ya se usó. Generá uno nuevo desde Integraciones › Telegram.';
const BASTA = 'Por hoy ya charlamos un montón. Mañana sigo respondiendo.';

/** Procesa un mensaje del bot. Devuelve la respuesta que se mandó (para los tests). */
export async function procesar(u: Update): Promise<string | null> {
	const m = u.message;
	if (!m?.text || m.chat.type !== 'private') return null;
	const chat = m.chat.id;
	const texto = m.text.trim().slice(0, 1000);
	const nombre = m.from?.first_name?.slice(0, 60) || 'Telegram';

	const responder = async (t: string) => {
		await enviar(chat, t);
		return t;
	};

	let cod: string | null = null;
	if (texto.startsWith('/start')) {
		cod = texto.split(/\s+/)[1]?.toUpperCase() ?? '';
		if (!/^[2-9A-Z]{8}$/.test(cod))
			return responder(
				`¡Hola! ${AYUDA}\n\nPara empezar, vinculá tu tienda desde el panel: Integraciones › Telegram › Vincular.`
			);
	}
	const c = { chat, texto, nombre, cod, responder };
	return usaSandbox() ? procesarSandbox(c) : procesarApi(c);
}

type Entrante = {
	chat: number;
	texto: string;
	nombre: string;
	/** El código del /start (ya validado), o null si no es un /start. */
	cod: string | null;
	responder: (t: string) => Promise<string>;
};

async function procesarSandbox({ chat, texto, nombre, cod, responder }: Entrante) {
	const db = dbSandbox(env.SANDBOX);
	if (cod) {
		const r = await buscarCambio<{ vence: number }>(db, 'tg-codigo', cod);
		if (!r || r.datos.vence < Date.now()) return responder(VENCIDO);
		await borrarCambio(db, r.sb.id, 'tg-codigo', cod);
		// Un chat por tienda: el anterior queda desvinculado.
		const antes = await vinculoDe(db, r.sb.id);
		if (antes && antes.chat !== chat)
			await borrarCambio(db, r.sb.id, 'tg-chat', String(antes.chat));
		await guardarCambio(db, r.sb.id, 'tg-chat', String(chat), { en: Date.now() });
		await guardarCambio(db, r.sb.id, 'cuenta', 'telegram', {
			chat,
			nombre,
			en: Date.now()
		} satisfies Vinculo);
		const tienda = await new DatosSandbox(db, r.sb).tienda();
		return responder(`¡Listo, ${nombre}! Quedé vinculado a ${tienda.nombre}. ${AYUDA}`);
	}

	const v = await buscarCambio<unknown>(db, 'tg-chat', String(chat));
	if (!v) return responder(SIN_VINCULO);
	if (texto === '/desvincular') {
		await desvincular(db, v.sb.id);
		return 'desvinculado';
	}
	if (texto === '/ayuda' || texto === '/help') return responder(AYUDA);

	const datos = new DatosSandbox(db, v.sb);
	const [tienda, hoy, semana, contexto] = await Promise.all([
		datos.tienda(),
		datos.tablero('hoy'),
		datos.tablero('7d'),
		datos.contextoHornero()
	]);
	// Los mismos topes que el Tincho del panel (cupo.ts): por tienda y para toda la demo.
	const usos = await sumarUsoHornero(db, v.sb.id, diaAr());
	return responderHornero(
		texto,
		{ tienda, hola: hoy.hola, hoy, semana, contexto },
		usos,
		responder
	);
}

async function procesarApi({ chat, texto, nombre, cod, responder }: Entrante) {
	const go = api();
	if (cod) {
		const r = await go.vincular(cod, chat, nombre);
		if (r.ok) return responder(`¡Listo, ${nombre}! Quedé vinculado a ${r.tienda}. ${AYUDA}`);
		if (r.motivo === 'ocupado')
			return responder(
				'Este chat ya está vinculado a otra tienda. Mandá /desvincular y después tocá el link de nuevo.'
			);
		if (r.motivo === 'demasiados')
			return responder('Probaste muchas veces seguidas. Esperá un rato y generá un link nuevo.');
		return responder(VENCIDO);
	}
	if (texto === '/desvincular') {
		if (!(await go.desvincularChat(chat))) return responder(SIN_VINCULO);
		await enviar(chat, DESVINCULADO);
		return 'desvinculado';
	}
	if (texto === '/ayuda' || texto === '/help') return responder(AYUDA);
	// Go cuenta el mensaje en el día de la tienda y devuelve lo que sabe Tincho (como el panel).
	const c = await go.contexto(chat);
	if (!c) return responder(SIN_VINCULO);
	const { tienda, hoy, semana, contexto, usos } = c;
	return responderHornero(
		texto,
		{ tienda, hola: hoy.hola, hoy, semana, contexto },
		usos,
		responder
	);
}

/** La respuesta de Tincho: Workers AI con los topes de cupo.ts, o la simulada. */
async function responderHornero(
	texto: string,
	d: DatosHornero,
	usos: number,
	responder: (t: string) => Promise<string>
) {
	if (usos > POR_VISITANTE) return responder(BASTA);
	let conModelo = env.HORNERO === 'workers-ai' && !!env.AI;
	if (conModelo && !(await cupoDemo())) conModelo = false;
	if (!conModelo || !env.AI) return responder(responderSimulado(texto, d));
	const conversacion: Mensaje[] = [
		{
			role: 'system',
			content: sistema(d) + '\nRespondés por Telegram: texto plano, corto, sin tablas ni markdown.'
		},
		{ role: 'user', content: texto }
	];
	try {
		const flujo = await responderConIa(env.AI, conversacion);
		const t = await new Response(flujo).text();
		return responder(t.trim() || responderSimulado(texto, d));
	} catch {
		return responder(responderSimulado(texto, d));
	}
}
