/**
 * El OAuth y el MCP con la API real (DATA_SOURCE=api): las tablas viven en Go (paquete ia) y el token
 * de la IA entra a la API como una sesión más, limitada a las rutas de las herramientas. Este módulo
 * traduce entre los nombres de OAuth (RFC 6749/7591) y los de la API.
 */
import { env } from 'cloudflare:workers';
import type { Permiso } from './oauth.ts';

type Resp<T> = { ok: true; v: T } | { ok: false; estado: number };

/**
 * Un error que no es culpa del pedido (Go caído, 5xx o el límite de pedidos): la IA tiene que
 * reintentar más tarde, no tirar sus tokens ni volver a pedir permiso.
 */
export type Fallo = { fallo: number };
const esFallo = (estado: number) => estado === 429 || estado >= 500;

async function go<T>(
	ruta: string,
	metodo: 'GET' | 'POST' | 'PUT' | 'DELETE',
	cuerpo?: unknown,
	token?: string | null,
	/** IP de quien llama: con la clave del BFF, Go aplica sus límites a esa IP y no a la del Worker. */
	ip?: string
): Promise<Resp<T>> {
	try {
		const res = await fetch(String(env.API_URL) + ruta, {
			method: metodo,
			headers: {
				accept: 'application/json',
				...(cuerpo === undefined ? {} : { 'content-type': 'application/json' }),
				...(token ? { authorization: `Bearer ${token}` } : {}),
				...(ip ? { 'x-mostrador-bff': String(env.CLAVE_BFF ?? ''), 'x-cliente-ip': ip } : {})
			},
			body: cuerpo === undefined ? undefined : JSON.stringify(cuerpo),
			// Go que acepta la conexión y no contesta también es un fallo, no una espera sin fin.
			signal: AbortSignal.timeout(10_000)
		});
		if (!res.ok) return { ok: false, estado: res.status };
		if (res.status === 204) return { ok: true, v: undefined as T };
		return { ok: true, v: (await res.json()) as T };
	} catch {
		// Red, plazo vencido o una respuesta cortada: para la IA, "probá más tarde".
		return { ok: false, estado: 503 };
	}
}

type ClienteGo = { id: string; nombre: string; redirecciones: string[] };
type TokensGo = { acceso: string; refresco: string; venceEn: number; acciones: boolean };

/** Registro dinámico: devuelve la respuesta de RFC 7591, o el estado de error de la API. */
export async function registrarApi(
	body: { client_name?: unknown; redirect_uris?: unknown },
	ip: string
) {
	const r = await go<ClienteGo>(
		'/v1/ia/clientes',
		'POST',
		{
			nombre: typeof body.client_name === 'string' ? body.client_name : '',
			redirecciones: Array.isArray(body.redirect_uris) ? body.redirect_uris : []
		},
		null,
		ip
	);
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
	if (r.ok) return r.v;
	return esFallo(r.estado) ? ({ fallo: r.estado } satisfies Fallo) : null;
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
	if (r.ok) return comoOAuth(r.v);
	return esFallo(r.estado) ? ({ fallo: r.estado } satisfies Fallo) : null;
}

/** false si Go no pudo revocar (caído o límite): el token puede seguir vivo. */
export async function revocarApi(token: string) {
	if (!token) return true;
	const r = await go('/v1/ia/revocar', 'POST', { token });
	return r.ok || !esFallo(r.estado);
}

/** La conexión del token del MCP: qué concedió la persona y los interruptores del panel. */
export async function conexionApi(token: string) {
	const r = await go<{ conexion: string; acciones: boolean; permisos: Record<string, boolean> }>(
		'/v1/ia/yo',
		'GET',
		undefined,
		token
	);
	if (!r.ok) return esFallo(r.estado) ? ({ fallo: r.estado } satisfies Fallo) : null;
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
	if (r.ok) return r.v;
	return esFallo(r.estado) ? ({ fallo: r.estado } satisfies Fallo) : null;
}

export async function desconectarApi(sesion: string | null, id: string) {
	const r = await go(
		`/v1/panel/ia/conexiones/${encodeURIComponent(id)}`,
		'DELETE',
		undefined,
		sesion
	);
	/** 0 si se desconectó; si no, el estado de Go. */
	return r.ok ? 0 : r.estado;
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
