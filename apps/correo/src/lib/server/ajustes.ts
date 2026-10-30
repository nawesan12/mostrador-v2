/**
 * Ajustes (handoff 2e y 2p): firma por casilla (la identidad JMAP de cada cuenta), respuesta
 * automática (VacationResponse, también en las compartidas), reglas y reenvío (un script Sieve
 * de la casilla propia, armado desde una lista de reglas) y la clave.
 */
import { Jmap, ErrorJmap, base } from './jmap.ts';

const VAC = 'urn:ietf:params:jmap:vacationresponse';
const SIEVE = 'urn:ietf:params:jmap:sieve';
const SUB = 'urn:ietf:params:jmap:submission';

export type Firma = {
	cuenta: string;
	email: string;
	identidad: string;
	nombre: string;
	texto: string;
};
export type Vacaciones = {
	cuenta: string;
	email: string;
	activa: boolean;
	desde: string | null;
	hasta: string | null;
	asunto: string;
	texto: string;
};

export async function firmasYVacaciones(j: Jmap) {
	const cuentas = await j.cuentas();
	const firmas: Firma[] = [];
	const vacaciones: Vacaciones[] = [];
	await Promise.all(
		cuentas.map(async (c) => {
			const r = await j.llamar(
				[
					['Identity/get', { accountId: c.id }, 'i'],
					['VacationResponse/get', { accountId: c.id }, 'v']
				],
				[SUB, VAC]
			);
			const ids = (r.get('i')?.list ?? []) as {
				id: string;
				email: string;
				name: string;
				textSignature: string;
			}[];
			const id = ids.find((x) => x.email === c.email) ?? ids[0];
			if (id)
				firmas.push({
					cuenta: c.id,
					email: c.email,
					identidad: id.id,
					nombre: id.name ?? '',
					texto: id.textSignature ?? ''
				});
			const v = (
				(r.get('v')?.list ?? []) as {
					isEnabled: boolean;
					fromDate: string | null;
					toDate: string | null;
					subject: string | null;
					textBody: string | null;
				}[]
			)[0];
			if (v)
				vacaciones.push({
					cuenta: c.id,
					email: c.email,
					activa: v.isEnabled,
					desde: v.fromDate,
					hasta: v.toDate,
					asunto: v.subject ?? '',
					texto: v.textBody ?? ''
				});
		})
	);
	const orden = (a: { cuenta: string }, b: { cuenta: string }) =>
		cuentas.findIndex((c) => c.id === a.cuenta) - cuentas.findIndex((c) => c.id === b.cuenta);
	return { firmas: firmas.sort(orden), vacaciones: vacaciones.sort(orden) };
}

export async function guardarFirma(
	j: Jmap,
	f: { cuenta: string; identidad: string; nombre: string; texto: string }
) {
	const r = await j.llamar(
		[
			[
				'Identity/set',
				{
					accountId: f.cuenta,
					update: {
						[f.identidad]: { name: f.nombre.slice(0, 120), textSignature: f.texto.slice(0, 4000) }
					}
				},
				'u'
			]
		],
		[SUB]
	);
	if (r.get('u')?.notUpdated) throw new Error('No se pudo guardar la firma.');
}

export async function guardarVacaciones(j: Jmap, v: Omit<Vacaciones, 'email'>) {
	const r = await j.llamar(
		[
			[
				'VacationResponse/set',
				{
					accountId: v.cuenta,
					update: {
						singleton: {
							isEnabled: v.activa,
							fromDate: v.desde,
							toDate: v.hasta,
							subject: v.asunto || null,
							textBody: v.texto.slice(0, 4000) || null
						}
					}
				},
				'u'
			]
		],
		[VAC]
	);
	if (r.get('u')?.notUpdated) throw new Error('No se pudo guardar la respuesta automática.');
}

// ---- Reglas y reenvío: un script Sieve propio, con la lista de reglas guardada adentro ----

export type Regla = {
	si: 'de' | 'asunto';
	valor: string;
	accion: 'mover' | 'etiquetar' | 'reenviar';
	destino: string;
};
export type Filtros = { reglas: Regla[]; reenviarA: string; guardarCopia: boolean };

const NOMBRE = 'correo-reglas';
const MARCA = '# correo-reglas: ';
const comillas = (s: string) => `"${s.replace(/\\/g, '\\\\').replace(/"/g, '\\"')}"`;
const EMAIL = /^[^\s@<>,;"]+@[^\s@<>,;"]+\.[^\s@<>,;"]+$/;

/** El Sieve de las reglas. La lista va en un comentario para volver a leerla tal cual. */
export function sieveDe(f: Filtros): string {
	const l = [MARCA + JSON.stringify(f), 'require ["fileinto", "imap4flags", "copy"];', ''];
	for (const r of f.reglas) {
		const valores = r.valor
			.split(/,| o /)
			.map((x) => x.trim())
			.filter(Boolean);
		if (!valores.length || !r.destino.trim()) continue;
		const lista = `[${valores.map(comillas).join(', ')}]`;
		const prueba =
			r.si === 'de' ? `address :contains "from" ${lista}` : `header :contains "subject" ${lista}`;
		const accion =
			r.accion === 'mover'
				? `fileinto ${comillas(r.destino)};`
				: r.accion === 'etiquetar'
					? `addflag ${comillas(etiquetaAKeyword(r.destino))};`
					: `redirect :copy ${comillas(r.destino)};`;
		l.push(`if ${prueba} {`, `\t${accion}`, '}');
	}
	if (f.reenviarA) l.push(`redirect${f.guardarCopia ? ' :copy' : ''} ${comillas(f.reenviarA)};`);
	return l.join('\n') + '\n';
}

/** "Pedido abierto" → keyword "pedido_abierto" (las etiquetas de la app son keywords JMAP). */
export const etiquetaAKeyword = (t: string) =>
	t
		.normalize('NFD')
		.replace(/[̀-ͯ]/g, '')
		.toLowerCase()
		.replace(/[^a-z0-9]+/g, '_')
		.replace(/^_+|_+$/g, '')
		.slice(0, 60);

export function validar(f: Filtros): Filtros {
	const reglas = (f.reglas ?? []).slice(0, 40).map((r) => {
		if (!['de', 'asunto'].includes(r.si) || !['mover', 'etiquetar', 'reenviar'].includes(r.accion))
			throw new Error('Regla inválida.');
		const valor = String(r.valor ?? '').slice(0, 300);
		const destino = String(r.destino ?? '')
			.trim()
			.slice(0, 120);
		if (r.accion === 'reenviar' && !EMAIL.test(destino))
			throw new Error(`«${destino}» no parece un email.`);
		return { si: r.si, valor, accion: r.accion, destino };
	});
	const reenviarA = String(f.reenviarA ?? '').trim();
	if (reenviarA && !EMAIL.test(reenviarA)) throw new Error(`«${reenviarA}» no parece un email.`);
	return { reglas, reenviarA, guardarCopia: f.guardarCopia !== false };
}

async function principal(j: Jmap) {
	const c = (await j.cuentas()).find((x) => x.personal);
	if (!c) throw new Error('No encontramos tu casilla.');
	return c;
}

export async function leerFiltros(j: Jmap): Promise<Filtros> {
	const c = await principal(j);
	const r = await j.llamar([['SieveScript/get', { accountId: c.id }, 's']], [SIEVE]);
	const s = ((r.get('s')?.list ?? []) as { name: string; blobId: string }[]).find(
		(x) => x.name === NOMBRE
	);
	const vacio: Filtros = { reglas: [], reenviarA: '', guardarCopia: true };
	if (!s) return vacio;
	const texto = await (
		await j.descargar(c.id, s.blobId, 'reglas.sieve', 'application/sieve')
	).text();
	const linea = texto.split('\n').find((x) => x.startsWith(MARCA));
	try {
		return linea ? validar(JSON.parse(linea.slice(MARCA.length))) : vacio;
	} catch {
		return vacio;
	}
}

export async function guardarFiltros(j: Jmap, f: Filtros) {
	const c = await principal(j);
	const blob = await j.subir(c.id, new Blob([sieveDe(f)], { type: 'application/sieve' }));
	const r = await j.llamar(
		[['SieveScript/get', { accountId: c.id, properties: ['name'] }, 's']],
		[SIEVE]
	);
	const ya = ((r.get('s')?.list ?? []) as { id: string; name: string }[]).find(
		(x) => x.name === NOMBRE
	);
	const res = await j.llamar(
		[
			[
				'SieveScript/set',
				ya
					? {
							accountId: c.id,
							update: { [ya.id]: { blobId: blob.blobId } },
							onSuccessActivateScript: ya.id
						}
					: {
							accountId: c.id,
							create: { n: { name: NOMBRE, blobId: blob.blobId } },
							onSuccessActivateScript: '#n'
						},
				'u'
			]
		],
		[SIEVE]
	);
	const error = (res.get('u')?.notCreated ?? res.get('u')?.notUpdated) as
		Record<string, { description?: string }> | undefined;
	if (error)
		throw new Error(
			`No se pudieron guardar las reglas: ${Object.values(error)[0]?.description ?? ''}`
		);
}

// ---- Clave: Stalwart no deja que cada uno la cambie por JMAP; lo hace el administrador ----

export async function cambiarClave(email: string, actual: string, nueva: string) {
	if (nueva.length < 10) throw new Error('La clave nueva tiene que tener al menos 10 caracteres.');
	if (nueva === actual) throw new Error('La clave nueva es igual a la actual.');
	await new Jmap({ email, clave: actual }).cuentas().catch((e) => {
		throw e instanceof ErrorJmap && e.estado === 401
			? new Error('La clave actual no es correcta.')
			: e;
	});
	const admin = process.env.CORREO_ADMIN;
	if (!admin) throw new Error('El cambio de clave no está habilitado en este servidor.');
	const auth = 'Basic ' + Buffer.from(admin).toString('base64');
	const pedir = async (calls: unknown[]) => {
		const r = await fetch(base() + '/jmap/', {
			method: 'POST',
			headers: { authorization: auth, 'content-type': 'application/json' },
			body: JSON.stringify({
				using: ['urn:ietf:params:jmap:core', 'urn:stalwart:jmap'],
				methodCalls: calls
			})
		});
		if (!r.ok) throw new Error(`Administración ${r.status}`);
		return ((await r.json()) as { methodResponses: [string, Record<string, unknown>, string][] })
			.methodResponses;
	};
	const [local, dominio] = email.split('@');
	const q = await pedir([['x:Account/query', { accountId: 'b', filter: { text: local } }, 'q']]);
	const ids = (q[0][1].ids ?? []) as string[];
	const g = await pedir([
		['x:Account/get', { accountId: 'b', ids, properties: ['emailAddress', '@type'] }, 'g']
	]);
	const cuenta = (
		(g[0][1].list ?? []) as { id: string; emailAddress: string; '@type': string }[]
	).find((x) => x.emailAddress === `${local}@${dominio}` && x['@type'] === 'User');
	if (!cuenta) throw new Error('No encontramos tu cuenta.');
	const u = await pedir([
		[
			'x:Account/set',
			{
				accountId: 'b',
				update: { [cuenta.id]: { credentials: { '0': { '@type': 'Password', secret: nueva } } } }
			},
			'u'
		]
	]);
	const hechos = u[0][1].updated as Record<string, unknown> | undefined;
	if (!hechos || !(cuenta.id in hechos)) throw new Error('No se pudo cambiar la clave.');
}
