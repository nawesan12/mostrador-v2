/** Piezas comunes de los endpoints de OAuth y MCP (los llaman servidores ajenos: CORS abierto). */
import { NO_DISPONIBLE } from './oauth.ts';

export const CORS = {
	'Access-Control-Allow-Origin': '*',
	'Access-Control-Allow-Methods': 'GET, POST, DELETE, OPTIONS',
	'Access-Control-Allow-Headers':
		'Authorization, Content-Type, Accept, Mcp-Session-Id, MCP-Protocol-Version, Last-Event-ID',
	'Access-Control-Expose-Headers': 'Mcp-Session-Id, WWW-Authenticate'
};

export const json = (body: unknown, status = 200, headers: Record<string, string> = {}) =>
	new Response(JSON.stringify(body), {
		status,
		headers: {
			'Content-Type': 'application/json',
			'Cache-Control': 'no-store',
			...CORS,
			...headers
		}
	});

/** DATA_SOURCE=api: la versión real va a vivir en la API Go. */
export const noDisponible = () =>
	json({ error: 'not_implemented', error_description: NO_DISPONIBLE }, 501);

export const preflight = () => new Response(null, { status: 204, headers: CORS });

/** Parámetros de un POST de OAuth: formulario (RFC 6749) o JSON. */
export async function parametros(request: Request) {
	const ct = request.headers.get('content-type') ?? '';
	if (ct.includes('application/json')) {
		const b = (await request.json().catch(() => ({}))) as Record<string, unknown>;
		return new URLSearchParams(
			Object.entries(b && typeof b === 'object' ? b : {}).map(([k, v]) => [k, String(v)])
		);
	}
	return new URLSearchParams(await request.text());
}
