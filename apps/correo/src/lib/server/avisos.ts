/**
 * Avisos en el teléfono (Web Push): Stalwart empuja directo al navegador con su propia clave
 * VAPID. La suscripción es sólo para correos que llegan (EmailDelivery), no para cada cambio:
 * así no hay pushes sin aviso, que en iPhone hacen que el sistema corte la suscripción.
 */
import { idDe } from './bandeja.ts';
import { Jmap } from './jmap.ts';

const PUSH = 'urn:ietf:params:jmap:webpush-vapid';

export async function claveVapid(j: Jmap): Promise<string | null> {
	const s = (await j.leerSesion()) as unknown as {
		capabilities: Record<string, { applicationServerKey?: string }>;
	};
	return s.capabilities[PUSH]?.applicationServerKey ?? null;
}

export type Suscripcion = { endpoint: string; keys: { p256dh: string; auth: string } };

/** Registra el navegador; si ya estaba (mismo endpoint), lo deja como está. */
export async function suscribir(j: Jmap, s: Suscripcion, dispositivo: string) {
	const r = await j.llamar([['PushSubscription/get', { ids: null }, 'g']]);
	const ya = ((r.get('g')?.list ?? []) as { id: string; url: string }[]).find(
		(x) => x.url === s.endpoint
	);
	if (ya) return ya.id;
	const c = await j.llamar([
		[
			'PushSubscription/set',
			{
				create: {
					n: {
						deviceClientId: dispositivo.slice(0, 120),
						url: s.endpoint,
						keys: { p256dh: s.keys.p256dh, auth: s.keys.auth },
						types: ['EmailDelivery']
					}
				}
			},
			's'
		]
	]);
	const creada = (c.get('s')?.created as Record<string, { id: string }> | undefined)?.n;
	if (!creada) throw new Error(JSON.stringify(c.get('s')?.notCreated ?? 'no se pudo'));
	return creada.id;
}

export async function verificar(j: Jmap, id: string, codigo: string) {
	await j.llamar([
		['PushSubscription/set', { update: { [id]: { verificationCode: codigo } } }, 'v']
	]);
}

export async function desuscribir(j: Jmap, endpoint: string) {
	const r = await j.llamar([['PushSubscription/get', { ids: null }, 'g']]);
	const ids = ((r.get('g')?.list ?? []) as { id: string; url: string }[])
		.filter((x) => x.url === endpoint)
		.map((x) => x.id);
	if (ids.length) await j.llamar([['PushSubscription/set', { destroy: ids }, 'd']]);
}

export type Aviso = {
	id: string;
	de: string;
	asunto: string;
	vista: string;
	casilla: string;
	fecha: string;
};

/** Los correos sin leer que entraron en la Entrada de cualquier casilla después de `desde`. */
export async function nuevos(j: Jmap, desde: string): Promise<Aviso[]> {
	const cuentas = await j.cuentas();
	const out: Aviso[] = [];
	await Promise.all(
		cuentas.map(async (c) => {
			const r = await j.llamar([
				['Mailbox/query', { accountId: c.id, filter: { role: 'inbox' } }, 'b'],
				[
					'Email/query',
					{
						accountId: c.id,
						filter: {
							operator: 'AND',
							conditions: [{ after: desde }, { notKeyword: '$seen' }]
						},
						sort: [{ property: 'receivedAt', isAscending: false }],
						limit: 5
					},
					'q'
				],
				[
					'Email/get',
					{
						accountId: c.id,
						'#ids': { resultOf: 'q', name: 'Email/query', path: '/ids' },
						properties: ['from', 'subject', 'preview', 'receivedAt', 'mailboxIds']
					},
					'e'
				]
			]);
			const entrada = ((r.get('b')?.ids ?? []) as string[])[0];
			for (const e of (r.get('e')?.list ?? []) as {
				id: string;
				from: { name: string | null; email: string }[] | null;
				subject: string | null;
				preview: string;
				receivedAt: string;
				mailboxIds: Record<string, boolean>;
			}[]) {
				if (entrada && !e.mailboxIds[entrada]) continue;
				const de = e.from?.[0];
				out.push({
					id: idDe(c.id, e.id),
					de: de?.name || de?.email || 'Alguien',
					asunto: e.subject || '(sin asunto)',
					vista: e.preview.slice(0, 140),
					casilla: c.email,
					fecha: e.receivedAt
				});
			}
		})
	);
	return out.sort((a, b) => b.fecha.localeCompare(a.fecha)).slice(0, 5);
}
