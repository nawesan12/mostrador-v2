import { expect, test, type Browser, type Page } from '@playwright/test';

/**
 * Tiempo real (docs/decisiones.md §44) de punta a punta: dos personas del mismo equipo (dueña a
 * 1440 y Ventas a 375) y la dueña de otra tienda, cada una en su navegador, más quien compra en la
 * tienda online. Lo que cambia por la API llega donde debe y no donde no.
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
			// Una IP por pedido (la que pasaría el BFF): los topes de registro son por IP.
			'x-cliente-ip': `10.${rand()}.${rand()}.${rand()}`,
			...(token ? { authorization: `Bearer ${token}` } : {})
		},
		body: cuerpo === undefined ? undefined : JSON.stringify(cuerpo)
	});
	const texto = await r.text();
	if (!r.ok) throw new Error(`${metodo} ${ruta}: ${r.status} ${texto}`);
	return (texto ? JSON.parse(texto) : undefined) as T;
}

async function registrar(nombre: string, tienda: string, email: string) {
	return api<{ token: string; tiendaId: string }>('POST', '/v1/auth/registro', {
		nombre,
		tienda,
		email,
		clave: 'Abcdefghijk1'
	});
}

/** El link de la invitación que mandó el worker (Mailpit). */
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

type Persona = { page: Page; marcos: string[] };

/** Un navegador con la sesión puesta, en el panel, mirando lo que llega por el WebSocket. */
async function enElPanel(
	browser: Browser,
	baseURL: string,
	email: string,
	ancho: number
): Promise<Persona> {
	const ctx = await browser.newContext({
		baseURL,
		viewport: { width: ancho, height: ancho < 760 ? 812 : 900 }
	});
	await ctx.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	const page = await ctx.newPage();
	const marcos: string[] = [];
	page.on('websocket', (ws) => {
		if (!ws.url().includes('/v1/tiempo-real')) return;
		ws.on('framereceived', (f) => marcos.push(String(f.payload)));
	});
	// Entra como cualquiera: email y contraseña (la cookie __Host- la pone el servidor).
	await page.goto('/ingresar');
	await page.locator('html[data-listo]').waitFor({ state: 'attached' });
	await page.getByRole('textbox', { name: 'Email' }).fill(email);
	await page.getByLabel('Contraseña', { exact: true }).fill('Abcdefghijk1');
	await page.getByRole('button', { name: 'Ingresar', exact: true }).click();
	await expect(page).toHaveURL(/\/panel/);
	await expect.poll(() => marcos.some((m) => m.includes('"hola"')), { timeout: 15_000 }).toBe(true);
	return { page, marcos };
}

const llego = (p: Persona, tipo: string) => p.marcos.some((m) => m.includes(`"tipo":"${tipo}"`));

test('los avisos llegan al equipo según su rol, nunca a otra tienda, y a quien compra', async ({
	browser,
	baseURL
}) => {
	test.setTimeout(150_000);
	const s = Date.now();
	const ana = await registrar('Ana', `Sol ${s}`, `ana+${s}@sol.com.ar`);
	await registrar('Eva', `Luna ${s}`, `eva+${s}@luna.com.ar`);
	// Beto entra al equipo de Sol con el rol Ventas.
	const emailBeto = `beto+${s}@sol.com.ar`;
	await api('POST', '/v1/panel/equipo', { contacto: emailBeto, rol: 'ventas' }, ana.token);
	await api('POST', '/v1/auth/invitacion/registro', {
		token: await tokenInvitacion(emailBeto),
		nombre: 'Beto',
		clave: 'Abcdefghijk1'
	});

	// Sol vende por transferencia y tiene un producto con una sola unidad.
	const conf = await api<{ negocio: Record<string, unknown> }>(
		'GET',
		'/v1/panel/configuracion',
		undefined,
		ana.token
	);
	await api(
		'PUT',
		'/v1/panel/configuracion/negocio',
		{ ...conf.negocio, alias: 'sol.tiempo.real' },
		ana.token
	);
	const prod = await api<{ id: string; slug: string }>(
		'POST',
		'/v1/panel/productos',
		{
			nombre: 'Mate imperial',
			precio: 1_000_000,
			stock: 1,
			canales: { web: true },
			publicado: true
		},
		ana.token
	);
	const { slug } = await api<{ slug: string }>('GET', '/v1/panel/tienda', undefined, ana.token);

	const pAna = await enElPanel(browser, baseURL!, `ana+${s}@sol.com.ar`, 1440);
	const pBeto = await enElPanel(browser, baseURL!, emailBeto, 375);
	const pEva = await enElPanel(browser, baseURL!, `eva+${s}@luna.com.ar`, 1440);

	// Alguien mira el producto en la tienda online.
	const compradora = await browser.newPage({ baseURL });
	await compradora.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	await compradora.goto(`/t/${slug}/producto/${prod.slug}`);
	await compradora.locator('html[data-listo]').waitFor({ state: 'attached' });

	// Otra persona compra la última unidad por la tienda online.
	const compra = await api<{ numero: number; codigo: string }>(
		'POST',
		`/v1/publico/tiendas/${slug}/pedidos`,
		{
			carrito: [{ id: prod.id, cantidad: 1 }],
			email: `lu+${s}@compras.com.ar`,
			nombre: 'Lu Pérez',
			telefono: '1155556666',
			envio: 'retiro',
			pago: 'transferencia'
		}
	);

	// Al equipo de Sol: «Pedido nuevo» con «Ver», en la compu y en el celular.
	for (const p of [pAna, pBeto]) {
		const aviso = p.page.locator('.toast', { hasText: `Pedido nuevo #${compra.numero}` });
		await expect(aviso).toBeVisible({ timeout: 15_000 });
		await expect(aviso.getByRole('button', { name: 'Ver' })).toBeVisible();
	}
	// La compradora que miraba el producto se entera de que se agotó.
	await expect(compradora.getByText('Se agotó Mate imperial')).toBeVisible({ timeout: 15_000 });

	// Un cupón: lo ve quien tiene Catálogo (la dueña), no Ventas.
	await api(
		'POST',
		'/v1/panel/cupones',
		{
			codigo: `TR${String(s).slice(-6)}`,
			tipo: 'pct',
			valor: 10,
			alcance: 'todo',
			quien: 'todos',
			minimo: 0,
			limite: 0,
			porCliente: 'libre',
			vigencia: 'sin',
			canales: ['web'],
			opciones: { hornero: false, combinable: false, unicos: false, link: false }
		},
		ana.token
	);
	await expect.poll(() => llego(pAna, 'cupon.nuevo'), { timeout: 10_000 }).toBe(true);

	// La página del pedido de quien compró: al marcarlo pagado, «¡Recibimos tu pago!».
	const pedido = await browser.newPage({ baseURL, viewport: { width: 375, height: 812 } });
	await pedido.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	await pedido.goto(`/t/${slug}/pedido/${compra.codigo}`);
	await pedido.locator('html[data-listo]').waitFor({ state: 'attached' });
	await expect(pedido.getByText('sol.tiempo.real')).toBeVisible();
	await api(
		'POST',
		`/v1/panel/pedidos/${compra.numero}/estado`,
		{ estado: 'Para preparar' },
		ana.token
	);
	await expect(pedido.getByText('¡Recibimos tu pago!')).toBeVisible({ timeout: 15_000 });
	// Pagado, ya no le pide transferir.
	await expect(pedido.getByText('sol.tiempo.real')).toHaveCount(0);

	// Lo que no le toca a nadie de afuera: Eva no recibió nada de Sol; Beto, ni cupones ni plata.
	await new Promise((ok) => setTimeout(ok, 1000));
	expect(pEva.marcos.filter((m) => !/"(hola|latido)"/.test(m))).toEqual([]);
	expect(llego(pBeto, 'cupon.nuevo')).toBe(false);
	expect(llego(pBeto, 'finanzas.cambio')).toBe(false);
	expect(llego(pBeto, 'pago.acreditado')).toBe(true);
	// Ningún aviso lleva datos de negocio: sólo tipo e id.
	for (const m of [...pAna.marcos, ...pBeto.marcos]) {
		expect(Object.keys(JSON.parse(m)).every((k) => ['tipo', 'id', 'lote'].includes(k))).toBe(true);
	}
	await pAna.page.screenshot({ path: 'test-results/tiempo-real-1440.png' });
	await pBeto.page.screenshot({ path: 'test-results/tiempo-real-375.png' });

	// Le cambian el rol a Beto: su panel se refresca solo y ahora recibe cupones.
	const equipo = await api<{ equipo: { id: string; contacto: string }[] }>(
		'GET',
		'/v1/panel/configuracion',
		undefined,
		ana.token
	);
	const idBeto = equipo.equipo.find((m) => m.contacto === emailBeto)?.id;
	expect(idBeto).toBeTruthy();
	const antes = pBeto.marcos.filter((m) => m.includes('"hola"')).length;
	await api('PUT', `/v1/panel/equipo/${idBeto}`, { rol: 'admin' }, ana.token);
	await expect
		.poll(() => pBeto.marcos.filter((m) => m.includes('"hola"')).length, { timeout: 15_000 })
		.toBeGreaterThan(antes);
});

test('sin conexión de tiempo real el panel anda igual y avisa «Reconectando…» si dura', async ({
	browser,
	baseURL
}) => {
	const s = Date.now();
	const email = `cora+${s}@sol.com.ar`;
	await registrar('Cora', `Sol ${s}`, email);
	for (const ancho of [1440, 375]) {
		const ctx = await browser.newContext({
			baseURL,
			viewport: { width: ancho, height: ancho < 760 ? 812 : 900 }
		});
		await ctx.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
		const page = await ctx.newPage();
		// El WebSocket no llega nunca a la API: se corta apenas abre (un proxy que no lo deja pasar).
		let intentos = 0;
		await page.routeWebSocket(/\/v1\/tiempo-real/, (ws) => {
			intentos++;
			void ws.close({ code: 1011, reason: 'sin api' });
		});
		await page.goto('/ingresar');
		await page.locator('html[data-listo]').waitFor({ state: 'attached' });
		await page.getByRole('textbox', { name: 'Email' }).fill(email);
		await page.getByLabel('Contraseña', { exact: true }).fill('Abcdefghijk1');
		await page.getByRole('button', { name: 'Ingresar', exact: true }).click();
		await expect(page).toHaveURL(/\/panel/);
		// Un parpadeo no se muestra; un corte que dura, sí (también en el celular).
		await expect(page.getByRole('status').filter({ hasText: 'Reconectando…' })).toBeVisible({
			timeout: 15_000
		});
		expect(intentos).toBeGreaterThan(1);
		await page.screenshot({ path: `test-results/tiempo-real-reconectando-${ancho}.png` });
		// El panel sigue andando: se navega y se ve como siempre.
		await page.goto('/panel/pedidos');
		await expect(page.getByRole('heading', { name: 'Pedidos', exact: true })).toBeVisible();
		await ctx.close();
	}
});
