/**
 * El OAuth y el MCP con la API real (DATA_SOURCE=api): las tablas viven en Go (paquete ia) y el token
 * de la IA entra a la API como una sesión más, limitada a las rutas de las herramientas. Este módulo
 * traduce entre los nombres de OAuth (RFC 6749/7591) y los de la API.
 */
import { env } from 'cloudflare:workers';
import type { Permiso } from './oauth.ts';

type Resp<T> = { ok: true; v: T } | { ok: false; estado: number };

async function go<T>(
	ruta: string,
	metodo: 'GET' | 'POST' | 'PUT' | 'DELETE',
	cuerpo?: unknown,
	token?: string | null
): Promise<Resp<T>> {
	const res = await fetch(String(env.API_URL) + ruta, {
		method: metodo,
		headers: {
			accept: 'application/json',
			...(cuerpo === undefined ? {} : { 'content-type': 'application/json' }),
			...(token ? { authorization: `Bearer ${token}` } : {})
		},
		body: cuerpo === undefined ? undefined : JSON.stringify(cuerpo)
	});
	if (!res.ok) return { ok: false, estado: res.status };
	if (res.status === 204) return { ok: true, v: undefined as T };
	return { ok: true, v: (await res.json()) as T };
}

type ClienteGo = { id: string; nombre: string; redirecciones: string[] };
type TokensGo = { acceso: string; refresco: string; venceEn: number; acciones: boolean };

/** Registro dinámico: devuelve la respuesta de RFC 7591, o el estado de error de la API. */
export async function registrarApi(body: { client_name?: unknown; redirect_uris?: unknown }) {
	const r = await go<ClienteGo>('/v1/ia/clientes', 'POST', {
		nombre: typeof body.client_name === 'string' ? body.client_name : '',
		redirecciones: Array.isArray(body.redirect_uris) ? body.redirect_uris : []
	});
	if (!r.ok) return r.estado;
	return {
		client_id: r.v.id,
		client_name: r.v.nombre,
		redirect_uris: r.v.redirecciones,
		token_endpoint_auth_method: 'none',
		grant_types: ['authorization_code', 'refresh_token'],
		response_types: ['code']
	};
}

export async function clienteApi(id: string) {
	if (!id || id.length > 64) return null;
	const r = await go<ClienteGo>(`/v1/ia/clientes/${encodeURIComponent(id)}`, 'GET');
	return r.ok ? r.v : null;
}

/** El "Permitir" de la pantalla de consentimiento, con la sesión de la persona. */
export async function autorizarApi(
	sesion: string,
	p: { cliente: string; redireccion: string; desafio: string; acciones: boolean }
) {
	const r = await go<{ codigo: string }>('/v1/ia/autorizar', 'POST', p, sesion);
	return r.ok ? r.v.codigo : r.estado;
}

const comoOAuth = (t: TokensGo) => ({
	access_token: t.acceso,
	token_type: 'Bearer' as const,
	expires_in: t.venceEn,
	refresh_token: t.refresco,
	scope: (t.acciones ? ['lectura', 'acciones'] : ['lectura']).join(' ')
});

export async function tokenApi(p: URLSearchParams) {
	const tipo = p.get('grant_type');
	if (tipo !== 'authorization_code' && tipo !== 'refresh_token') return undefined;
	const r = await go<TokensGo>('/v1/ia/token', 'POST', {
		tipo,
		codigo: p.get('code') ?? '',
		verificador: p.get('code_verifier') ?? '',
		cliente: p.get('client_id') ?? '',
		redireccion: p.get('redirect_uri') ?? '',
		refresco: p.get('refresh_token') ?? ''
	});
	return r.ok ? comoOAuth(r.v) : null;
}

export async function revocarApi(token: string) {
	if (token) await go('/v1/ia/revocar', 'POST', { token });
}

/** La conexión del token del MCP: qué concedió la persona y los interruptores del panel. */
export async function conexionApi(token: string) {
	const r = await go<{ conexion: string; acciones: boolean; permisos: Record<string, boolean> }>(
		'/v1/ia/yo',
		'GET',
		undefined,
		token
	);
	if (!r.ok) return null;
	const permisos: Permiso[] = r.v.acciones ? ['lectura', 'acciones'] : ['lectura'];
	return { permisos, permisosPanel: r.v.permisos };
}

export type ConexionPanel = {
	id: string;
	cliente: string;
	persona: string;
	acciones: boolean;
	creadaEn: string;
	usadaEn: string;
};

export async function panelApi(sesion: string | null) {
	const r = await go<{ permisos: Record<string, boolean>; conexiones: ConexionPanel[] }>(
		'/v1/panel/ia',
		'GET',
		undefined,
		sesion
	);
	return r.ok ? r.v : null;
}

export async function desconectarApi(sesion: string | null, id: string) {
	const r = await go(
		`/v1/panel/ia/conexiones/${encodeURIComponent(id)}`,
		'DELETE',
		undefined,
		sesion
	);
	return r.ok;
}

export async function permisoApi(sesion: string | null, permiso: string, valor: boolean) {
	const r = await go(
		`/v1/panel/ia/permisos/${encodeURIComponent(permiso)}`,
		'PUT',
		{ valor },
		sesion
	);
	return r.ok;
}
