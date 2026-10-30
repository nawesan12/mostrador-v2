import { createHash, randomBytes } from 'node:crypto';
import { expect, test, type APIRequestContext, type Page } from '@playwright/test';

/**
 * "Manejá Tu Local Web desde tu IA" de punta a punta, como lo haría Claude: registro dinámico,
 * consentimiento en /oauth/autorizar (sin demo: elige tienda y vuelve), canje del código con PKCE,
 * tools/list, una herramienta de lectura y una de acción, refresh rotativo y revocación.
 */
const VUELTA = 'https://claude.ai/api/mcp/auth_callback';
const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });
const b64url = (b: Buffer) => b.toString('base64url');

async function rpc(api: APIRequestContext, token: string, method: string, params = {}) {
	const r = await api.post('/api/mcp', {
		headers: { authorization: `Bearer ${token}`, 'content-type': 'application/json' },
		data: { jsonrpc: '2.0', id: 1, method, params }
	});
	return { status: r.status(), body: r.status() === 200 ? await r.json() : null };
}

test('una IA se conecta por OAuth, usa el servidor MCP y se desconecta', async ({
	page,
	request
}) => {
	// Metadatos: el recurso apunta al servidor de autorización, que exige PKCE S256.
	const recurso = await (await request.get('/.well-known/oauth-protected-resource/api/mcp')).json();
	expect(recurso.resource).toMatch(/\/api\/mcp$/);
	const meta = await (await request.get('/.well-known/oauth-authorization-server')).json();
	expect(meta.code_challenge_methods_supported).toEqual(['S256']);

	// Sin token: 401 con el puntero a los metadatos.
	const sin = await request.post('/api/mcp', { data: { jsonrpc: '2.0', id: 1, method: 'ping' } });
	expect(sin.status()).toBe(401);
	expect(sin.headers()['www-authenticate']).toContain('resource_metadata=');

	// 1. Registro dinámico.
	const reg = await request.post('/api/oauth/register', {
		data: { client_name: 'Claude', redirect_uris: [VUELTA] }
	});
	expect(reg.status()).toBe(201);
	const { client_id } = await reg.json();

	// 2. Consentimiento: sin demo, primero se elige una tienda y se vuelve.
	const verificador = b64url(randomBytes(32));
	const desafio = b64url(createHash('sha256').update(verificador).digest());
	const q = new URLSearchParams({
		response_type: 'code',
		client_id,
		redirect_uri: VUELTA,
		code_challenge: desafio,
		code_challenge_method: 'S256',
		state: 'estado-123',
		scope: 'lectura acciones'
	});
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	await page.goto(`/oauth/autorizar?${q}`);
	await expect(page).toHaveURL(/\/demo\?volver=/);
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/oauth\/autorizar\?/);
	await listo(page);
	await expect(page.getByRole('heading', { name: /Conectar Claude/ })).toBeVisible();
	expect(
		await page.evaluate(() => document.documentElement.scrollWidth - innerWidth)
	).toBeLessThanOrEqual(0);
	await expect(page.getByRole('checkbox', { name: /También puede hacer cambios/ })).toBeChecked();

	// Permitir lleva a la vuelta de la IA con el código y el state.
	const vuelta = page.waitForRequest((r) => r.url().startsWith(VUELTA));
	await page.route(`${VUELTA}**`, (r) => r.fulfill({ status: 200, body: 'ok' }));
	await page.getByRole('button', { name: 'Permitir' }).click();
	const destino = new URL((await vuelta).url());
	expect(destino.searchParams.get('state')).toBe('estado-123');
	const code = destino.searchParams.get('code')!;
	expect(code).toBeTruthy();

	// 3. Canje del código: con otro verificador falla y el código ya no sirve (un solo uso).
	const canje = (v: string) =>
		request.post('/api/oauth/token', {
			form: {
				grant_type: 'authorization_code',
				code,
				client_id,
				redirect_uri: VUELTA,
				code_verifier: v
			}
		});
	const malo = await canje(b64url(randomBytes(32)));
	expect(malo.status()).toBe(400);
	expect((await canje(verificador)).status()).toBe(400);

	// Otra vuelta de consentimiento (la demo ya existe: va directo) y ahora sí, el verificador bueno.
	await page.goto(`/oauth/autorizar?${q}`);
	await listo(page);
	const vuelta2 = page.waitForRequest((r) => r.url().startsWith(VUELTA));
	await page.getByRole('button', { name: 'Permitir' }).click();
	const code2 = new URL((await vuelta2).url()).searchParams.get('code')!;
	const tokens = await (
		await request.post('/api/oauth/token', {
			form: {
				grant_type: 'authorization_code',
				code: code2,
				client_id,
				redirect_uri: VUELTA,
				code_verifier: verificador
			}
		})
	).json();
	expect(tokens.access_token).toMatch(/^mostrador_mcp_/);
	expect(tokens.scope).toBe('lectura acciones');

	// 4. MCP: initialize, tools/list, una lectura y una acción.
	const init = await rpc(request, tokens.access_token, 'initialize', {
		protocolVersion: '2025-06-18'
	});
	expect(init.body.result.serverInfo.name).toBe('mostrador');
	const lista = await rpc(request, tokens.access_token, 'tools/list');
	const nombres = lista.body.result.tools.map((t: { name: string }) => t.name);
	expect(nombres).toEqual(
		expect.arrayContaining(['resumen_del_dia', 'listar_pedidos', 'ajustar_producto'])
	);

	const productos = await rpc(request, tokens.access_token, 'tools/call', {
		name: 'listar_productos',
		arguments: { limite: 1 }
	});
	const prod = productos.body.result.structuredContent.productos[0];
	expect(prod.id).toBeTruthy();
	expect(prod.precio.texto).toMatch(/^\$/);

	const ajuste = await rpc(request, tokens.access_token, 'tools/call', {
		name: 'ajustar_producto',
		arguments: { id: prod.id, stock: 77 }
	});
	expect(ajuste.body.result.isError).toBeFalsy();
	expect(ajuste.body.result.structuredContent.stock.ahora).toBe(77);

	// El cambio se ve en el panel, y la conexión aparece en Integraciones.
	await page.goto('/panel/integraciones');
	await listo(page);
	await expect(page.getByRole('button', { name: /Claude/ }).first()).toContainText('Conectado');
	await page
		.getByRole('button', { name: /Claude/ })
		.first()
		.click();
	const d = page.getByRole('dialog');
	await expect(d).toContainText('/api/mcp');
	await expect(d).toContainText('Conectores');
	await d.getByRole('button', { name: 'Listo' }).click();

	// 5. Refresh rotativo: el refresh viejo deja de servir.
	const nuevo = await (
		await request.post('/api/oauth/token', {
			form: { grant_type: 'refresh_token', refresh_token: tokens.refresh_token, client_id }
		})
	).json();
	expect(nuevo.access_token).toBeTruthy();
	expect((await rpc(request, tokens.access_token, 'ping')).status).toBe(401);
	expect((await rpc(request, nuevo.access_token, 'ping')).status).toBe(200);

	// 6. Revocar: el token deja de servir y la conexión sale del panel.
	const rev = await request.post('/api/oauth/revoke', { form: { token: nuevo.refresh_token } });
	expect(rev.status()).toBe(200);
	expect((await rpc(request, nuevo.access_token, 'ping')).status).toBe(401);
	await page.reload();
	await listo(page);
	await expect(page.locator('.conexiones')).toHaveCount(0);
});

test('desconectar desde el panel revoca la conexión', async ({ page, request }) => {
	const reg = await request.post('/api/oauth/register', {
		data: { client_name: 'ChatGPT', redirect_uris: [VUELTA] }
	});
	const { client_id } = await reg.json();
	const verificador = b64url(randomBytes(32));
	const q = new URLSearchParams({
		response_type: 'code',
		client_id,
		redirect_uri: VUELTA,
		code_challenge: b64url(createHash('sha256').update(verificador).digest()),
		code_challenge_method: 'S256',
		scope: 'lectura'
	});
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	await page.route(`${VUELTA}**`, (r) => r.fulfill({ status: 200, body: 'ok' }));
	await page.goto(`/oauth/autorizar?${q}`);
	await listo(page);
	// Sólo pidió lectura: no se ofrece hacer cambios.
	await expect(page.getByRole('checkbox')).toHaveCount(0);
	const vuelta = page.waitForRequest((r) => r.url().startsWith(VUELTA));
	await page.getByRole('button', { name: 'Permitir' }).click();
	const code = new URL((await vuelta).url()).searchParams.get('code')!;
	const tokens = await (
		await request.post('/api/oauth/token', {
			form: {
				grant_type: 'authorization_code',
				code,
				client_id,
				redirect_uri: VUELTA,
				code_verifier: verificador
			}
		})
	).json();

	// Sólo lectura: las acciones no aparecen y, si se llaman, se rechazan.
	const lista = await rpc(request, tokens.access_token, 'tools/list');
	const nombres = lista.body.result.tools.map((t: { name: string }) => t.name);
	expect(nombres).toContain('resumen_del_dia');
	expect(nombres).not.toContain('cambiar_estado_pedido');
	const resumen = await rpc(request, tokens.access_token, 'tools/call', {
		name: 'resumen_del_dia',
		arguments: {}
	});
	expect(resumen.body.result.structuredContent.ventas_hoy.texto).toMatch(/^\$/);
	const accion = await rpc(request, tokens.access_token, 'tools/call', {
		name: 'cambiar_estado_pedido',
		arguments: { numero: 1, estado: 'Cancelado' }
	});
	expect(accion.body.result.isError).toBe(true);

	await page.goto('/panel/integraciones');
	await listo(page);
	const fila = page.locator('.cx', { hasText: 'ChatGPT' });
	await expect(fila).toContainText('Sólo lectura');
	await fila.getByRole('button', { name: 'Desconectar' }).click();
	await expect(page.getByRole('status').filter({ hasText: 'ChatGPT desconectado' })).toHaveCount(1);
	await expect(page.locator('.cx')).toHaveCount(0);
	expect((await rpc(request, tokens.access_token, 'ping')).status).toBe(401);
});
