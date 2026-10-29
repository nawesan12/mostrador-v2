/**
 * Un Tiendanube falso para las pruebas (playwright.config.ts lo levanta en 4179): el canje del
 * código y la API de la tienda 4567 con el caso compartido con Go. El preview lo usa por
 * TIENDANUBE_LOCAL; la autorización (www.tiendanube.com) la intercepta el navegador con page.route.
 */
import { createServer, type ServerResponse } from 'node:http';
import casos from '../../../packages/contrato/casos/importar/tiendanube.json' with { type: 'json' };

const API = '/2025-03/4567';

function responder(res: ServerResponse, estado: number, cuerpo: unknown) {
	res.writeHead(estado, { 'content-type': 'application/json', 'x-main-language': 'es' });
	res.end(typeof cuerpo === 'string' ? cuerpo : JSON.stringify(cuerpo));
}

createServer((req, res) => {
	const u = new URL(req.url ?? '/', 'http://127.0.0.1:4179');
	if (u.pathname === '/salud') return responder(res, 200, 'ok');
	if (u.pathname === '/apps/authorize/token' && req.method === 'POST') {
		let cuerpo = '';
		req.on('data', (c) => (cuerpo += c));
		req.on('end', () => {
			let b: Record<string, string> = {};
			try {
				b = JSON.parse(cuerpo);
			} catch {
				// Queda vacío: invalid_grant.
			}
			if (b.code === 'codigo-e2e' && b.client_secret === 'secreto-e2e')
				responder(res, 200, { access_token: 'tok-e2e', scope: 'read_products', user_id: 4567 });
			else responder(res, 400, { error: 'invalid_grant' });
		});
		return;
	}
	if (req.headers.authorization !== 'Bearer tok-e2e') return responder(res, 401, '');
	if (u.searchParams.get('page') !== '1') return responder(res, 404, 'Last page is 0');
	if (u.pathname === `${API}/products`) return responder(res, 200, casos.productos);
	if (u.pathname === `${API}/customers`) return responder(res, 200, casos.clientes);
	if (u.pathname === `${API}/orders`) return responder(res, 200, casos.pedidos);
	responder(res, 404, '');
}).listen(4179, '127.0.0.1');
