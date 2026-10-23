/**
 * OAuth 2.1 mínimo para clientes MCP (MCP Authorization spec), como el de Caja: registro dinámico de
 * clientes, código de autorización con PKCE S256 obligatorio (un solo uso, 5 min), access token
 * corto y refresh token que rota en cada uso. Todo se guarda hasheado (sha256), nunca en claro.
 *
 * En la demo (DATA_SOURCE=sandbox) la conexión queda atada al sandbox del visitante: vence con él y
 * se borra con él. Con la API real (DATA_SOURCE=api) esto lo va a atender Go; mientras tanto los
 * endpoints responden 501 (`oauthDisponible`).
 */
import { and, eq, gt, lt, sql } from 'drizzle-orm';
import { b64url, randomToken, sha256 } from '../crypto.ts';
import type { DbSandbox } from '../sandbox/sandbox.ts';
import { oauthClientes, oauthCodigos, oauthConexiones, sandboxes } from '../sandbox/schema.ts';

/** Lo que una IA puede hacer: `lectura` siempre; `acciones` si la persona lo permite. */
export const PERMISOS = ['lectura', 'acciones'] as const;
export type Permiso = (typeof PERMISOS)[number];

export const ACCESO_MS = 60 * 60 * 1000;
export const CODIGO_MS = 5 * 60 * 1000;
/** Pedidos al servidor MCP por conexión y por minuto. */
export const LIMITE_POR_MINUTO = 60;

export const PREFIJO_ACCESO = 'mostrador_mcp_';
export const PREFIJO_REFRESH = 'mostrador_rt_';

export const NO_DISPONIBLE =
	'Conectar una IA a tu cuenta real de Mostrador todavía no está disponible. Por ahora funciona con la demo.';

export const metadatosServidor = (base: string) => ({
	issuer: base,
	authorization_endpoint: `${base}/oauth/autorizar`,
	token_endpoint: `${base}/api/oauth/token`,
	registration_endpoint: `${base}/api/oauth/register`,
	revocation_endpoint: `${base}/api/oauth/revoke`,
	response_types_supported: ['code'],
	grant_types_supported: ['authorization_code', 'refresh_token'],
	code_challenge_methods_supported: ['S256'],
	token_endpoint_auth_methods_supported: ['none'],
	scopes_supported: [...PERMISOS]
});

export const metadatosRecurso = (base: string) => ({
	resource: `${base}/api/mcp`,
	authorization_servers: [base],
	scopes_supported: [...PERMISOS],
	bearer_methods_supported: ['header']
});

/** Los permisos que se piden: `acciones` sólo si se pide (o si no se pide nada en particular). Pura. */
export function permisosPedidos(scope: string | null): Permiso[] {
	const pedidos = (scope ?? '').split(/[\s+]+/).filter(Boolean);
	return !pedidos.length || pedidos.includes('acciones') ? ['lectura', 'acciones'] : ['lectura'];
}

/** https, o localhost para clientes de escritorio (RFC 8252). Pura. */
export function vueltaValida(u: string) {
	try {
		const url = new URL(u);
		return (
			url.protocol === 'https:' ||
			(url.protocol === 'http:' && ['localhost', '127.0.0.1', '[::1]'].includes(url.hostname))
		);
	} catch {
		return false;
	}
}

/** S256: base64url(sha256(verificador)) tiene que dar el desafío. */
export async function pkceValido(verificador: string, desafio: string) {
	if (!/^[A-Za-z0-9._~-]{43,128}$/.test(verificador)) return false;
	const h = b64url(await crypto.subtle.digest('SHA-256', new TextEncoder().encode(verificador)));
	return h === desafio;
}

export async function registrarCliente(
	db: DbSandbox,
	body: { client_name?: unknown; redirect_uris?: unknown },
	ahora = Date.now()
) {
	const uris = (Array.isArray(body.redirect_uris) ? body.redirect_uris : [])
		.filter((u): u is string => typeof u === 'string' && u.length <= 500 && vueltaValida(u))
		.slice(0, 10);
	if (!uris.length) return null;
	const id = `mcp_${randomToken(16)}`;
	const nombre =
		(typeof body.client_name === 'string' ? body.client_name.trim() : '').slice(0, 60) ||
		'Otra IA (MCP)';
	await db
		.insert(oauthClientes)
		.values({ id, nombre, redirecciones: JSON.stringify(uris), creadoEn: ahora });
	return {
		client_id: id,
		client_name: nombre,
		redirect_uris: uris,
		token_endpoint_auth_method: 'none',
		grant_types: ['authorization_code', 'refresh_token'],
		response_types: ['code']
	};
}

export async function cliente(db: DbSandbox, id: string) {
	if (!id || id.length > 64) return null;
	const [c] = await db.select().from(oauthClientes).where(eq(oauthClientes.id, id)).limit(1);
	return c ? { ...c, redirecciones: JSON.parse(c.redirecciones) as string[] } : null;
}

export async function emitirCodigo(
	db: DbSandbox,
	p: {
		clienteId: string;
		sandboxId: string;
		redireccion: string;
		desafio: string;
		permisos: Permiso[];
	},
	ahora = Date.now()
) {
	const codigo = randomToken(32);
	await db.insert(oauthCodigos).values({
		hash: await sha256(codigo),
		clienteId: p.clienteId,
		sandboxId: p.sandboxId,
		redireccion: p.redireccion,
		desafio: p.desafio,
		permisos: p.permisos.join(' '),
		venceEn: ahora + CODIGO_MS
	});
	return codigo;
}

type Tokens = {
	access_token: string;
	token_type: 'Bearer';
	expires_in: number;
	refresh_token: string;
	scope: string;
};

async function parNuevo(venceEn: number, ahora: number) {
	const acceso = PREFIJO_ACCESO + randomToken(32);
	const refresh = PREFIJO_REFRESH + randomToken(32);
	const accesoVence = Math.min(ahora + ACCESO_MS, venceEn);
	return {
		acceso,
		refresh,
		accesoHash: await sha256(acceso),
		refreshHash: await sha256(refresh),
		accesoVence
	};
}

const respuesta = (
	par: { acceso: string; refresh: string; accesoVence: number },
	permisos: string,
	ahora: number
): Tokens => ({
	access_token: par.acceso,
	token_type: 'Bearer',
	expires_in: Math.max(1, Math.round((par.accesoVence - ahora) / 1000)),
	refresh_token: par.refresh,
	scope: permisos
});

/**
 * Canjea el código (se borra al leerlo: un solo uso) por un par de tokens. El cliente, la vuelta y el
 * verificador PKCE tienen que coincidir; el sandbox tiene que seguir vivo.
 */
export async function canjearCodigo(
	db: DbSandbox,
	p: { codigo: string; clienteId: string; redireccion: string; verificador: string },
	ahora = Date.now()
): Promise<Tokens | null> {
	if (!p.codigo) return null;
	const [fila] = await db
		.delete(oauthCodigos)
		.where(eq(oauthCodigos.hash, await sha256(p.codigo)))
		.returning();
	if (!fila || fila.venceEn < ahora) return null;
	if (fila.clienteId !== p.clienteId || fila.redireccion !== p.redireccion) return null;
	if (!(await pkceValido(p.verificador, fila.desafio))) return null;
	const [c, sb] = await Promise.all([
		cliente(db, fila.clienteId),
		db
			.select({ venceEn: sandboxes.venceEn })
			.from(sandboxes)
			.where(eq(sandboxes.id, fila.sandboxId))
			.limit(1)
			.then((f) => f[0])
	]);
	if (!c || !sb || sb.venceEn <= ahora) return null;
	const par = await parNuevo(sb.venceEn, ahora);
	await db.insert(oauthConexiones).values({
		id: randomToken(12),
		sandboxId: fila.sandboxId,
		clienteId: c.id,
		nombre: c.nombre,
		permisos: fila.permisos,
		accesoHash: par.accesoHash,
		accesoVence: par.accesoVence,
		refreshHash: par.refreshHash,
		venceEn: sb.venceEn,
		creadoEn: ahora
	});
	return respuesta(par, fila.permisos, ahora);
}

/**
 * Rota el refresh token: el viejo deja de servir y el access token anterior también. Si llega un
 * refresh token ya rotado (alguien lo copió), se corta la conexión entera (OAuth 2.1 §4.3.1).
 */
export async function renovar(
	db: DbSandbox,
	p: { refresh: string; clienteId: string },
	ahora = Date.now()
): Promise<Tokens | null> {
	if (!p.refresh) return null;
	const hash = await sha256(p.refresh);
	const [fila] = await db
		.select({
			id: oauthConexiones.id,
			clienteId: oauthConexiones.clienteId,
			venceEn: oauthConexiones.venceEn,
			permisos: oauthConexiones.permisos
		})
		.from(oauthConexiones)
		.where(eq(oauthConexiones.refreshHash, hash))
		.limit(1);
	if (!fila) {
		await db.delete(oauthConexiones).where(eq(oauthConexiones.refreshAnterior, hash));
		return null;
	}
	if (fila.clienteId !== p.clienteId || fila.venceEn <= ahora) return null;
	const par = await parNuevo(fila.venceEn, ahora);
	// Condicionado al hash viejo: dos renovaciones a la vez no sacan dos pares válidos.
	const [ok] = await db
		.update(oauthConexiones)
		.set({
			accesoHash: par.accesoHash,
			accesoVence: par.accesoVence,
			refreshHash: par.refreshHash,
			refreshAnterior: hash
		})
		.where(and(eq(oauthConexiones.id, fila.id), eq(oauthConexiones.refreshHash, hash)))
		.returning({ id: oauthConexiones.id });
	if (!ok) return null;
	return respuesta(par, fila.permisos, ahora);
}

/** Revocación (RFC 7009): sirve el access token o el refresh token y corta la conexión. */
export async function revocar(db: DbSandbox, token: string) {
	if (!token || token.length > 200) return;
	const hash = await sha256(token);
	await db
		.delete(oauthConexiones)
		.where(
			sql`${oauthConexiones.accesoHash} = ${hash} or ${oauthConexiones.refreshHash} = ${hash}`
		);
}

export type Conexion = {
	id: string;
	sandboxId: string;
	nombre: string;
	permisos: Permiso[];
	/** Pedidos en el minuto en curso (para el límite). */
	pedidos: number;
};

/**
 * Valida `Authorization: Bearer mostrador_mcp_…` y suma el pedido al límite por minuto, en una
 * sola escritura. null si no existe, venció o se revocó.
 */
export async function autenticar(
	db: DbSandbox,
	header: string | null,
	ahora = Date.now()
): Promise<Conexion | null> {
	const token = (header ?? '').replace(/^Bearer\s+/i, '').trim();
	if (!token.startsWith(PREFIJO_ACCESO) || token.length > 200) return null;
	const [fila] = await db
		.update(oauthConexiones)
		.set({
			pedidos: sql`case when ${oauthConexiones.ventanaHasta} < ${ahora} then 1 else ${oauthConexiones.pedidos} + 1 end`,
			ventanaHasta: sql`case when ${oauthConexiones.ventanaHasta} < ${ahora} then ${ahora + 60_000} else ${oauthConexiones.ventanaHasta} end`,
			usadoEn: ahora
		})
		.where(
			and(
				eq(oauthConexiones.accesoHash, await sha256(token)),
				gt(oauthConexiones.accesoVence, ahora),
				gt(oauthConexiones.venceEn, ahora)
			)
		)
		.returning({
			id: oauthConexiones.id,
			sandboxId: oauthConexiones.sandboxId,
			nombre: oauthConexiones.nombre,
			permisos: oauthConexiones.permisos,
			pedidos: oauthConexiones.pedidos
		});
	if (!fila) return null;
	return { ...fila, permisos: fila.permisos.split(' ') as Permiso[] };
}

/** Las conexiones activas de un sandbox (Integraciones › Manejá Mostrador desde tu IA). */
export async function conexionesDe(db: DbSandbox, sandboxId: string, ahora = Date.now()) {
	return db
		.select({
			id: oauthConexiones.id,
			nombre: oauthConexiones.nombre,
			permisos: oauthConexiones.permisos,
			creadoEn: oauthConexiones.creadoEn,
			usadoEn: oauthConexiones.usadoEn
		})
		.from(oauthConexiones)
		.where(and(eq(oauthConexiones.sandboxId, sandboxId), gt(oauthConexiones.venceEn, ahora)))
		.orderBy(oauthConexiones.creadoEn);
}

/** Desconectar desde el panel: sólo una conexión del propio sandbox. */
export async function desconectar(db: DbSandbox, sandboxId: string, id: string) {
	const borradas = await db
		.delete(oauthConexiones)
		.where(and(eq(oauthConexiones.id, id), eq(oauthConexiones.sandboxId, sandboxId)))
		.returning({ id: oauthConexiones.id });
	return borradas.length > 0;
}

/**
 * Limpieza del cron: códigos vencidos y clientes registrados hace más de 30 días sin ninguna
 * conexión (las conexiones se van solas con su sandbox).
 */
export async function purgarOauth(db: DbSandbox, ahora = Date.now()) {
	await db.batch([
		db.delete(oauthCodigos).where(lt(oauthCodigos.venceEn, ahora)),
		db
			.delete(oauthClientes)
			.where(
				and(
					lt(oauthClientes.creadoEn, ahora - 30 * 86_400_000),
					sql`not exists (select 1 from oauth_conexiones c where c.cliente_id = ${oauthClientes.id})`
				)
			)
	]);
}
