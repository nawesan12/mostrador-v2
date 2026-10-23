import { env } from 'cloudflare:workers';
import { usaSandbox } from '#lib/server/datos/index.ts';
import {
	herramientasPara,
	INSTRUCTIONS,
	llamar,
	PROTOCOL_VERSIONS,
	SERVER_INFO,
	type Contexto
} from '#lib/server/mcp/herramientas.ts';
import { conexionApi } from '#lib/server/mcp/api.ts';
import { CORS, json, preflight } from '#lib/server/mcp/http.ts';
import { autenticar, LIMITE_POR_MINUTO } from '#lib/server/mcp/oauth.ts';
import { DatosApi } from '#lib/server/datos/api.ts';
import { DatosSandbox, dbSandbox, leer } from '#lib/server/sandbox/sandbox.ts';
import type { RequestHandler } from './$types';

/**
 * Servidor MCP (Streamable HTTP, sin sesión), como el de Caja: cada POST trae un mensaje JSON-RPC
 * 2.0 (o un lote de hasta 10) y respondemos con JSON. Autenticación: `Bearer mostrador_mcp_…` del
 * OAuth de /oauth/autorizar. En la demo, los datos son los del sandbox de la conexión.
 */

type Mensaje = {
	jsonrpc?: string;
	id?: string | number | null;
	method?: string;
	params?: Record<string, unknown>;
};

const errorRpc = (id: Mensaje['id'], code: number, message: string) => ({
	jsonrpc: '2.0',
	id: id ?? null,
	error: { code, message }
});

const sinToken = (url: URL) =>
	json(
		errorRpc(
			null,
			-32001,
			'Token inválido, vencido o revocado. Volvé a conectar Mostrador desde tu IA.'
		),
		401,
		{
			'WWW-Authenticate': `Bearer realm="mostrador", error="invalid_token", resource_metadata="${url.origin}/.well-known/oauth-protected-resource"`
		}
	);

/** La conexión del pedido: en la demo, el sandbox; con la API real, el token va a Go tal cual. */
async function conexion(request: Request, url: URL) {
	const auth = request.headers.get('authorization');
	if (!usaSandbox()) {
		const token = auth?.startsWith('Bearer ') ? auth.slice(7).trim() : '';
		const cx = token && token.length <= 200 ? await conexionApi(token) : null;
		if (!cx) return sinToken(url);
		const c: Contexto = {
			datos: new DatosApi(String(env.API_URL), token),
			permisos: cx.permisos,
			permisosPanel: cx.permisosPanel,
			sitio: url.origin
		};
		// El límite por minuto lo pone la API (Guardia del paquete ia).
		return { c, nombre: 'tu tienda' };
	}
	const db = dbSandbox(env.SANDBOX);
	const cx = await autenticar(db, auth);
	const sb = cx && (await leer(db, cx.sandboxId));
	if (!cx || !sb) return sinToken(url);
	if (cx.pedidos > LIMITE_POR_MINUTO)
		return json(
			errorRpc(
				null,
				-32000,
				`Pasaste el límite de ${LIMITE_POR_MINUTO} pedidos por minuto. Esperá un momento.`
			),
			429,
			{ 'Retry-After': '60' }
		);
	const c: Contexto = { datos: new DatosSandbox(db, sb), permisos: cx.permisos, sitio: url.origin };
	return { c, nombre: cx.nombre };
}

export const POST: RequestHandler = async ({ request, url }) => {
	const cx = await conexion(request, url);
	if (cx instanceof Response) return cx;
	const { c } = cx;

	let body: unknown;
	try {
		body = await request.json();
	} catch {
		return json(errorRpc(null, -32700, 'JSON inválido'), 400);
	}
	const lote = Array.isArray(body);
	const mensajes = (lote ? body : [body]) as Mensaje[];
	if (!mensajes.length) return json(errorRpc(null, -32600, 'Lote vacío'), 400);
	if (mensajes.length > 10) return json(errorRpc(null, -32600, 'Máximo 10 mensajes por lote'), 400);

	const respuestas = [];
	for (const m of mensajes) {
		const r = await atender(c, cx.nombre, m);
		if (r) respuestas.push(r);
	}
	// Sólo notificaciones: 202 sin cuerpo.
	if (!respuestas.length) return new Response(null, { status: 202, headers: CORS });
	return json(lote ? respuestas : respuestas[0]);
};

async function atender(c: Contexto, nombre: string, m: Mensaje) {
	if (!m || typeof m !== 'object' || m.jsonrpc !== '2.0')
		return errorRpc(m?.id, -32600, 'Mensaje JSON-RPC inválido');
	// Respuestas del cliente y notificaciones: nada que responder.
	if (!m.method || m.id === undefined || m.id === null) return null;
	const ok = (result: unknown) => ({ jsonrpc: '2.0', id: m.id, result });
	const p = m.params ?? {};
	switch (m.method) {
		case 'initialize': {
			const pedida = String(p.protocolVersion ?? '');
			return ok({
				protocolVersion: PROTOCOL_VERSIONS.includes(pedida) ? pedida : PROTOCOL_VERSIONS[0],
				capabilities: { tools: { listChanged: false } },
				serverInfo: SERVER_INFO,
				instructions: `${INSTRUCTIONS} Conectado como: ${nombre}. Permisos: ${
					c.permisos.includes('acciones') ? 'lectura y acciones' : 'sólo lectura'
				}.`
			});
		}
		case 'ping':
			return ok({});
		case 'tools/list':
			return ok({ tools: herramientasPara(c.permisos) });
		case 'tools/call': {
			const args =
				p.arguments && typeof p.arguments === 'object'
					? (p.arguments as Record<string, unknown>)
					: {};
			return ok(await llamar(c, String(p.name ?? ''), args));
		}
		case 'resources/list':
			return ok({ resources: [] });
		case 'resources/templates/list':
			return ok({ resourceTemplates: [] });
		case 'prompts/list':
			return ok({ prompts: [] });
		default:
			return errorRpc(m.id, -32601, `Método no soportado: ${m.method}`);
	}
}

/** Sin streams SSE (servidor sin estado). */
export const GET: RequestHandler = () =>
	new Response('Este servidor MCP no ofrece stream SSE: usá POST.', {
		status: 405,
		headers: { Allow: 'POST', ...CORS }
	});

export const DELETE: RequestHandler = () =>
	new Response(null, { status: 405, headers: { Allow: 'POST', ...CORS } });

export const OPTIONS = preflight;
