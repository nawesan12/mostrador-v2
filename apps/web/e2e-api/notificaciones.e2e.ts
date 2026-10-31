import { expect, test, type Browser, type Page } from '@playwright/test';

/**
 * La campanita (docs/decisiones.md §45) de punta a punta con la API, el worker y el tiempo real: la
 * dueña a 1440 y Ventas a 375. Las notificaciones salen de los mismos eventos que los mails, llegan
 * sin recargar, cada una ve lo suyo (Ventas sin montos) y tocar una abre su página y la marca leída.
 */
const API = process.env.E2E_API ?? 'http://127.0.0.1:4000';
const MAILPIT = process.env.E2E_MAILPIT ?? 'http://127.0.0.1:8026';
const CABECERAS = { 'content-type': 'application/json', 'x-mostrador-bff': 'clave-bff-e2e' };

const rand = () => Math.floor(Math.random() * 250) + 1;

async function api<T = unknown>(metodo: string, ruta: string, cuerpo?: unknown, token?: string) {
	const r = await fetch(API + ruta, {
		method: metodo,
		headers: {
			...CABECERAS,
			'x-cliente-ip': `10.${rand()}.${rand()}.${rand()}`,
			...(token ? { authorization: `Bearer ${token}` } : {})
		},
		body: cuerpo === undefined ? undefined : JSON.stringify(cuerpo)
	});
	const texto = await r.text();
	if (!r.ok) throw new Error(`${metodo} ${ruta}: ${r.status} ${texto}`);
	return (texto ? JSON.parse(texto) : undefined) as T;
}

async function tokenInvitacion(para: string): Promise<string> {
	for (let i = 0; i < 80; i++) {
		const r = await fetch(`${MAILPIT}/api/v1/search?query=${encodeURIComponent(`to:${para}`)}`);
		const j = (await r.json()) as { messages: { ID: string }[] };
		for (const { ID } of j.messages ?? []) {
			const m = (await (await fetch(`${MAILPIT}/api/v1/message/${ID}`)).json()) as { Text: string };
			const t = m.Text.match(/invitacion\?token=([\w-]+)/);
			if (t) return t[1];
		}
		await new Promise((ok) => setTimeout(ok, 250));
	}
	throw new Error(`No llegó la invitación a ${para}`);
}

/** Un navegador con la sesión puesta, en el panel, con el tiempo real conectado. */
async function enElPanel(browser: Browser, baseURL: string, email: string, ancho: number) {
	const ctx = await browser.newContext({
		baseURL,
		viewport: { width: ancho, height: ancho < 760 ? 812 : 900 }
	});
	await ctx.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	const page = await ctx.newPage();
	let hola = false;
	page.on('websocket', (ws) => {
		if (!ws.url().includes('/v1/tiempo-real')) return;
		ws.on('framereceived', (f) => {
			if (String(f.payload).includes('"hola"')) hola = true;
		});
	});
	await page.goto('/ingresar');
	await page.locator('html[data-listo]').waitFor({ state: 'attached' });
	await page.getByRole('textbox', { name: 'Email' }).fill(email);
	await page.getByLabel('Contraseña', { exact: true }).fill('Abcdefghijk1');
	await page.getByRole('button', { name: 'Ingresar', exact: true }).click();
	await expect(page).toHaveURL(/\/panel/);
	await expect.poll(() => hola, { timeout: 15_000 }).toBe(true);
	return page;
}

const campanita = (p: Page) =>
	p.locator('header.top').getByRole('button', { name: /^Notificaciones/ });
const sinDesborde = (p: Page) =>
	p.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);

test('la campanita: llega sola, cada una ve lo suyo y tocar abre y marca leída', async ({
	browser,
	baseURL
}) => {
	test.setTimeout(150_000);
	const s = Date.now();
	const emailAna = `ana+${s}@sol.com.ar`;
	const ana = await api<{ token: string }>('POST', '/v1/auth/registro', {
		nombre: 'Ana',
		tienda: `Sol ${s}`,
		email: emailAna,
		clave: 'Abcdefghijk1'
	});
	const pAna = await enElPanel(browser, baseURL!, emailAna, 1440);

	// Vacía de verdad: sin insignia y lo dice.
	await expect(campanita(pAna)).toHaveAccessibleName('Notificaciones');
	await campanita(pAna).click();
	const cajaAna = pAna.getByRole('dialog', { name: 'Notificaciones' });
	await expect(
		cajaAna.getByText('No tenés notificaciones por ahora.', { exact: false })
	).toBeVisible();
	await pAna.keyboard.press('Escape');

	// Beto acepta la invitación (como Ventas): a Ana le llega sin recargar.
	const emailBeto = `beto+${s}@sol.com.ar`;
	await api('POST', '/v1/panel/equipo', { contacto: emailBeto, rol: 'ventas' }, ana.token);
	await api('POST', '/v1/auth/invitacion/registro', {
		token: await tokenInvitacion(emailBeto),
		nombre: 'Beto',
		clave: 'Abcdefghijk1'
	});
	await expect(campanita(pAna)).toHaveAccessibleName('Notificaciones, 1 sin leer', {
		timeout: 15_000
	});

	// Una venta por transferencia en la tienda online.
	const conf = await api<{ negocio: Record<string, unknown> }>(
		'GET',
		'/v1/panel/configuracion',
		undefined,
		ana.token
	);
	await api(
		'PUT',
		'/v1/panel/configuracion/negocio',
		{ ...conf.negocio, alias: 'sol.campanita' },
		ana.token
	);
	const prod = await api<{ id: string }>(
		'POST',
		'/v1/panel/productos',
		{ nombre: 'Mate imperial', precio: 25_000, stock: 5, canales: { web: true }, publicado: true },
		ana.token
	);
	const { slug } = await api<{ slug: string }>('GET', '/v1/panel/tienda', undefined, ana.token);
	const pBeto = await enElPanel(browser, baseURL!, emailBeto, 375);
	await expect(campanita(pBeto)).toHaveAccessibleName('Notificaciones');
	const compra = await api<{ numero: number }>('POST', `/v1/publico/tiendas/${slug}/pedidos`, {
		carrito: [{ id: prod.id, cantidad: 2 }],
		email: `lu+${s}@compras.com.ar`,
		nombre: 'Lu Pérez',
		telefono: '1155556666',
		envio: 'retiro',
		pago: 'transferencia'
	});
	const titulo = `Nueva venta #${compra.numero}`;
	await expect(campanita(pAna)).toHaveAccessibleName('Notificaciones, 2 sin leer', {
		timeout: 15_000
	});
	await expect(campanita(pBeto)).toHaveAccessibleName('Notificaciones, 1 sin leer', {
		timeout: 15_000
	});

	// Ana ve el monto; Beto (Ventas) no. Beto no se entera de su propia llegada al equipo.
	await campanita(pAna).click();
	const ventaAna = cajaAna.locator('.n', { hasText: titulo });
	await expect(ventaAna).toContainText('$');
	await expect(ventaAna).toContainText('Transferencia por confirmar');
	await expect(cajaAna.locator('.n', { hasText: 'Beto se sumó al equipo' })).toBeVisible();
	expect(await sinDesborde(pAna)).toBeLessThanOrEqual(0);
	await pAna.screenshot({ path: 'test-results/campanita-1440.png' });

	await campanita(pBeto).click();
	const cajaBeto = pBeto.getByRole('dialog', { name: 'Notificaciones' });
	await expect(cajaBeto.locator('.n')).toHaveCount(1);
	const ventaBeto = cajaBeto.locator('.n', { hasText: titulo });
	await expect(ventaBeto).toContainText('2 productos · Transferencia por confirmar');
	await expect(ventaBeto).not.toContainText('$');
	expect(await sinDesborde(pBeto)).toBeLessThanOrEqual(0);
	const caja = (await cajaBeto.boundingBox())!;
	expect(caja.x).toBeGreaterThanOrEqual(0);
	expect(caja.x + caja.width).toBeLessThanOrEqual(375);
	await pBeto.screenshot({ path: 'test-results/campanita-375.png' });

	// Tocarla abre el pedido y la marca leída.
	await ventaBeto.click();
	await expect(pBeto).toHaveURL(new RegExp(`/panel/pedidos\\?pedido=${compra.numero}$`));
	await expect(campanita(pBeto)).toHaveAccessibleName('Notificaciones', { timeout: 10_000 });

	// «Marcar todo como leído» en la de Ana.
	await cajaAna.getByRole('button', { name: 'Marcar todo como leído' }).click();
	await expect(campanita(pAna)).toHaveAccessibleName('Notificaciones');
	await expect(cajaAna.getByLabel('Sin leer')).toHaveCount(0);
	// Y quedó guardado.
	await pAna.reload();
	await expect(campanita(pAna)).toHaveAccessibleName('Notificaciones');
});
