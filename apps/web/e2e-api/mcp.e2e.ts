import { createHash, randomBytes } from 'node:crypto';
import { expect, test, type APIRequestContext, type Page } from '@playwright/test';

/**
 * El MCP con la API real: la IA se registra, la dueña entra con su cuenta y autoriza, el token va a
 * Go, las herramientas leen y escriben con los límites de la Guardia, y al revocar se corta.
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

test('una IA se conecta a una tienda real por OAuth y usa el MCP contra la API', async ({
	page,
	request
}) => {
	// La dueña: cuenta nueva (queda con sesión).
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	await page.goto('/ingresar?m=registro');
	await listo(page);
	await page.getByRole('textbox', { name: 'Tu nombre' }).fill('Lucía');
	await page.getByRole('textbox', { name: 'Email' }).fill(`lucia+ia${Date.now()}@laesquina.com.ar`);
	await page.getByLabel('Contraseña', { exact: true }).fill('Abcdefghijk1');
	await page.locator('label.terminos .caja-check').click();
	await page.getByRole('button', { name: 'Crear mi tienda' }).click();
	await expect(page).toHaveURL(/\/onboarding/);

	// Registro dinámico (va a Go).
	const reg = await request.post('/api/oauth/register', {
		data: { client_name: 'Claude', redirect_uris: [VUELTA] }
	});
	expect(reg.status()).toBe(201);
	const { client_id } = await reg.json();

	// Consentimiento con la sesión real; pide acciones y la dueña las deja.
	const verificador = b64url(randomBytes(32));
	const q = new URLSearchParams({
		response_type: 'code',
		client_id,
		redirect_uri: VUELTA,
		code_challenge: b64url(createHash('sha256').update(verificador).digest()),
		code_challenge_method: 'S256',
		state: 'e2e',
		scope: 'lectura acciones'
	});
	await page.goto(`/oauth/autorizar?${q}`);
	await listo(page);
	await expect(page.getByRole('heading', { name: /Conectar Claude/ })).toBeVisible();
	const vuelta = page.waitForRequest((r) => r.url().startsWith(VUELTA));
	await page.route(`${VUELTA}**`, (r) => r.fulfill({ status: 200, body: 'ok' }));
	await page.getByRole('button', { name: 'Permitir' }).click();
	const code = new URL((await vuelta).url()).searchParams.get('code')!;

	const tk = await request.post('/api/oauth/token', {
		form: {
			grant_type: 'authorization_code',
			code,
			client_id,
			redirect_uri: VUELTA,
			code_verifier: verificador
		}
	});
	expect(tk.status()).toBe(200);
	const { access_token, refresh_token, scope } = await tk.json();
	expect(scope).toBe('lectura acciones');

	// Herramientas: lista, una lectura y una acción.
	const lista = await rpc(request, access_token, 'tools/list');
	expect(lista.body.result.tools.map((t: { name: string }) => t.name)).toContain(
		'crear_link_de_pago'
	);
	const pedidos = await rpc(request, access_token, 'tools/call', {
		name: 'listar_pedidos',
		arguments: {}
	});
	expect(pedidos.body.result.isError).toBeFalsy();
	const link = await rpc(request, access_token, 'tools/call', {
		name: 'crear_link_de_pago',
		arguments: { concepto: 'Seña de prueba', monto: 5000 }
	});
	expect(link.body.result.isError, JSON.stringify(link.body)).toBeFalsy();
	// Crear cupones necesita "Reembolsos y descuentos", que viene apagado: lo dice la herramienta.
	const cupon = await rpc(request, access_token, 'tools/call', {
		name: 'crear_cupon',
		arguments: { codigo: 'IAPRUEBA', tipo: 'porcentaje', valor: 10 }
	});
	expect(cupon.body.result.isError).toBe(true);
	expect(JSON.stringify(cupon.body)).toContain('Reembolsos y descuentos');

	// El panel la muestra; refresco rota; revocar corta.
	await page.goto('/panel/integraciones');
	await listo(page);
	const nuevo = await request.post('/api/oauth/token', {
		form: { grant_type: 'refresh_token', refresh_token, client_id }
	});
	expect(nuevo.status()).toBe(200);
	const { access_token: acceso2 } = await nuevo.json();
	expect((await rpc(request, access_token, 'ping')).status).toBe(401);
	await request.post('/api/oauth/revoke', { form: { token: acceso2 } });
	expect((await rpc(request, acceso2, 'ping')).status).toBe(401);
});
