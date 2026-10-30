/**
 * Mandar un correo desde una casilla (la propia o una compartida): se crea en los Borradores de
 * esa cuenta, se entrega con su identidad y, entregado, pasa a sus Enviados. Respondiendo, lleva
 * In-Reply-To y References, y el original queda marcado como respondido ($answered).
 */
import { Jmap, SUBMISSION } from './jmap.ts';
import { partir } from './bandeja.ts';

export type Nuevo = {
	cuenta: string;
	para: string[];
	cc: string[];
	asunto: string;
	texto: string;
	/** Los blobs ya subidos (/subir). */
	adjuntos: { blobId: string; nombre: string; tipo: string }[];
	/** El correo al que responde (id de la bandeja). */
	responde?: string | null;
	/** Programado: ISO; Stalwart lo retiene hasta esa hora (FUTURERELEASE). */
	programado?: string | null;
};

const EMAIL = /^[^\s@<>,;]+@[^\s@<>,;]+\.[^\s@<>,;]+$/;

export function direcciones(s: string[]): { email: string }[] {
	return s
		.flatMap((x) => x.split(/[,;]/))
		.map((x) => x.trim().replace(/^.*<([^>]+)>$/, '$1'))
		.filter(Boolean)
		.map((email) => {
			if (!EMAIL.test(email)) throw new Error(`«${email}» no parece un email.`);
			return { email };
		});
}

export async function enviar(j: Jmap, n: Nuevo) {
	const para = direcciones(n.para);
	const cc = direcciones(n.cc);
	if (!para.length) throw new Error('Falta a quién mandarlo.');
	const r = await j.llamar(
		[
			['Identity/get', { accountId: n.cuenta }, 'i'],
			['Mailbox/get', { accountId: n.cuenta, properties: ['role'] }, 'b']
		],
		[SUBMISSION]
	);
	const cuenta = (await j.cuentas()).find((c) => c.id === n.cuenta);
	if (!cuenta) throw new Error('No tenés esa casilla.');
	const ids = (r.get('i')?.list ?? []) as { id: string; email: string; name: string }[];
	const identidad = ids.find((x) => x.email === cuenta.email) ?? ids[0];
	if (!identidad) throw new Error('Esa casilla no puede mandar correos.');
	const bs = (r.get('b')?.list ?? []) as { id: string; role: string | null }[];
	const borradores = bs.find((b) => b.role === 'drafts')?.id;
	const enviados = bs.find((b) => b.role === 'sent')?.id;
	if (!borradores || !enviados) throw new Error('A esa casilla le faltan Borradores o Enviados.');

	let enRespuesta: { id: string; messageId: string | null; refs: string[] } | null = null;
	if (n.responde) {
		const [c, e] = partir(n.responde);
		const o = await j.llamar([
			['Email/get', { accountId: c, ids: [e], properties: ['messageId', 'references'] }, 'o']
		]);
		const m = (
			(o.get('o')?.list ?? []) as { messageId: string[] | null; references: string[] | null }[]
		)[0];
		if (m)
			enRespuesta = {
				id: n.responde,
				messageId: m.messageId?.[0] ?? null,
				refs: m.references ?? []
			};
	}

	const email: Record<string, unknown> = {
		mailboxIds: { [borradores]: true },
		keywords: { $draft: true, $seen: true },
		from: [{ name: identidad.name || null, email: identidad.email }],
		to: para,
		cc,
		subject: n.asunto,
		bodyValues: { t: { value: n.texto } },
		textBody: [{ partId: 't', type: 'text/plain' }],
		attachments: n.adjuntos.map((a) => ({
			blobId: a.blobId,
			type: a.tipo,
			name: a.nombre,
			disposition: 'attachment'
		}))
	};
	if (enRespuesta?.messageId) {
		email.inReplyTo = [enRespuesta.messageId];
		email.references = [...enRespuesta.refs, enRespuesta.messageId];
	}
	const envio: Record<string, unknown> = { identityId: identidad.id, emailId: '#c' };
	if (n.programado) {
		const hasta = new Date(n.programado);
		if (Number.isNaN(+hasta) || +hasta < Date.now())
			throw new Error('Elegí una hora que no haya pasado.');
		envio.envelope = {
			mailFrom: {
				email: identidad.email,
				parameters: { HOLDUNTIL: hasta.toISOString().replace(/\.\d+Z$/, 'Z') }
			},
			rcptTo: [...para, ...cc].map((d) => ({ email: d.email }))
		};
	}
	const res = await j.llamar(
		[
			['Email/set', { accountId: n.cuenta, create: { c: email } }, 'e'],
			[
				'EmailSubmission/set',
				{
					accountId: n.cuenta,
					create: { s: envio },
					onSuccessUpdateEmail: {
						'#s': {
							[`mailboxIds/${borradores}`]: null,
							[`mailboxIds/${enviados}`]: true,
							'keywords/$draft': null
						}
					}
				},
				's'
			]
		],
		[SUBMISSION]
	);
	const noCreado = (res.get('e')?.notCreated as Record<string, { description?: string }>)?.c;
	const noEnviado = (
		res.get('s')?.notCreated as Record<string, { description?: string; type?: string }>
	)?.s;
	if (noCreado || noEnviado)
		throw new Error(
			`No se pudo mandar: ${(noCreado ?? noEnviado)?.description ?? (noEnviado?.type || 'error')}`
		);
	if (enRespuesta) {
		const [c, e] = partir(enRespuesta.id);
		await j.llamar([
			['Email/set', { accountId: c, update: { [e]: { 'keywords/$answered': true } } }, 'a']
		]);
	}
}
