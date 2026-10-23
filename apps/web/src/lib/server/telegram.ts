/**
 * Hornero en Telegram (@mostradorhornero_bot): la dueña vincula su tienda con un código de un solo
 * uso (Integraciones › Telegram › Vincular), le pregunta en lenguaje natural y recibe un aviso por
 * cada venta. En la demo vive sobre el sandbox; con la API real, el webhook lo atiende Go.
 *
 * Seguridad: Telegram manda el secreto del webhook en X-Telegram-Bot-Api-Secret-Token (se compara en
 * tiempo constante) y el chat sólo ve la tienda a la que se vinculó.
 */
import { env } from 'cloudflare:workers';
import { sistema, type DatosHornero } from './hornero/contexto.ts';
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

/** Un código de un solo uso para vincular el chat con esta tienda; devuelve el link al bot. */
export async function linkVinculo(db: DbSandbox, sandboxId: string) {
	const cod = codigo();
	await guardarCambio(db, sandboxId, 'tg-codigo', cod, { vence: Date.now() + VIGENCIA });
	return `https://t.me/${BOT}?start=${cod}`;
}

export async function vinculoDe(db: DbSandbox, sandboxId: string) {
	const c = await cambiosDe(db, sandboxId, 'cuenta');
	return (c.get('telegram') as Vinculo | null | undefined) ?? null;
}

export async function desvincular(db: DbSandbox, sandboxId: string) {
	const v = await vinculoDe(db, sandboxId);
	if (!v) return;
	await borrarCambio(db, sandboxId, 'tg-chat', String(v.chat));
	await guardarCambio(db, sandboxId, 'cuenta', 'telegram', null);
	await enviar(
		v.chat,
		'Listo, desvinculé este chat de la tienda. Para volver, tocá Vincular en Integraciones.'
	);
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
	'Soy Hornero, de Mostrador. Preguntame lo que quieras de tu tienda: "¿qué se vendió hoy?", "¿qué tengo para preparar?", "¿qué me falta?". Y te aviso cada venta.';

/** Procesa un mensaje del bot. Devuelve la respuesta que se mandó (para los tests). */
export async function procesar(u: Update): Promise<string | null> {
	const m = u.message;
	if (!m?.text || m.chat.type !== 'private') return null;
	const db = dbSandbox(env.SANDBOX);
	const chat = m.chat.id;
	const texto = m.text.trim().slice(0, 1000);

	const responder = async (t: string) => {
		await enviar(chat, t);
		return t;
	};

	if (texto.startsWith('/start')) {
		const cod = texto.split(/\s+/)[1]?.toUpperCase() ?? '';
		if (!/^[2-9A-Z]{8}$/.test(cod))
			return responder(
				`¡Hola! ${AYUDA}\n\nPara empezar, vinculá tu tienda desde el panel: Integraciones › Telegram › Vincular.`
			);
		const r = await buscarCambio<{ vence: number }>(db, 'tg-codigo', cod);
		if (!r || r.datos.vence < Date.now())
			return responder(
				'Ese link venció o ya se usó. Generá uno nuevo desde Integraciones › Telegram.'
			);
		await borrarCambio(db, r.sb.id, 'tg-codigo', cod);
		// Un chat por tienda: el anterior queda desvinculado.
		const antes = await vinculoDe(db, r.sb.id);
		if (antes && antes.chat !== chat)
			await borrarCambio(db, r.sb.id, 'tg-chat', String(antes.chat));
		const nombre = m.from?.first_name?.slice(0, 60) || 'Telegram';
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
	if (!v)
		return responder(
			'Todavía no sé de qué tienda sos. Vinculá este chat desde el panel: Integraciones › Telegram › Vincular.'
		);
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
	const d: DatosHornero = { tienda, hola: hoy.hola, hoy, semana, contexto };
	// Los mismos topes que el Hornero del panel (cupo.ts): por tienda y para toda la demo.
	const usos = await sumarUsoHornero(db, v.sb.id, diaAr());
	if (usos > POR_VISITANTE)
		return responder('Por hoy ya charlamos un montón. Mañana sigo respondiendo.');
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
