import { expect, test, type Page } from '@playwright/test';

const MAILPIT = process.env.E2E_MAILPIT ?? 'http://127.0.0.1:8026';
const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });

/** Último mail que llegó a Mailpit para esa dirección (texto plano). */
/**
 * El mail más nuevo a `para` que cumple `busca`: a la misma casilla llegan otros (bienvenida, «Entraron
 * a tu cuenta») y el que se espera puede tardar un poco más.
 */
async function ultimoMail(para: string, busca: RegExp = /./): Promise<string> {
	for (let i = 0; i < 40; i++) {
		const r = await fetch(`${MAILPIT}/api/v1/search?query=${encodeURIComponent(`to:${para}`)}`);
		const j = (await r.json()) as { messages: { ID: string }[] };
		for (const { ID } of j.messages ?? []) {
			const m = (await (await fetch(`${MAILPIT}/api/v1/message/${ID}`)).json()) as {
				Text: string;
			};
			if (busca.test(m.Text)) return m.Text;
		}
		await new Promise((res) => setTimeout(res, 250));
	}
	throw new Error(`No llegó el mail esperado a ${para}`);
}

test('registro, salir, recuperar por mail, contraseña nueva e ingreso', async ({
	page,
	context
}) => {
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	const email = `lucia+${Date.now()}@laesquina.com.ar`;

	// Registro: crea cuenta y tienda y sigue en el onboarding, que se guarda en la API.
	await page.goto('/ingresar?m=registro');
	await listo(page);
	await page.getByRole('textbox', { name: 'Tu nombre' }).fill('Lucía');
	await page.getByRole('textbox', { name: 'Email' }).fill(email);
	await page.getByLabel('Contraseña', { exact: true }).fill('Abcdefghijk1');
	await page.locator('label.terminos .caja-check').click();
	await page.getByRole('button', { name: 'Crear mi tienda' }).click();
	await expect(page).toHaveURL(/\/onboarding\?paso=1$/);
	await listo(page);
	await expect(page.getByRole('heading', { name: /Contanos un poco de tu negocio/ })).toBeVisible();
	// Del paso 2 al 6 sin elegir nada nuevo, y al panel.
	for (let i = 0; i < 4; i++) await page.locator('.pie .seguir').click();
	await expect(page.getByRole('heading', { name: 'Tu tienda ya está en línea.' })).toBeVisible();
	await page.getByRole('button', { name: /Ir a mi panel/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	// Sin nombre en el registro ni cambio en «Tu negocio», queda el provisorio.
	await expect(page.locator('.barra .t-n')).toHaveText('La tienda de Lucía');
	const cookie = (await context.cookies()).find((c) => c.name.endsWith('tlw_sesion'));
	expect(cookie?.httpOnly).toBe(true);

	// El mismo email no se registra dos veces.
	await context.clearCookies();
	await page.goto('/ingresar?m=registro');
	await listo(page);
	await page.getByRole('textbox', { name: 'Tu nombre' }).fill('Otra');
	await page.getByRole('textbox', { name: 'Email' }).fill(email);
	await page.getByLabel('Contraseña', { exact: true }).fill('Abcdefghijk1');
	await page.locator('label.terminos .caja-check').click();
	await page.getByRole('button', { name: 'Crear mi tienda' }).click();
	await expect(page.getByText(/Ese email ya tiene una cuenta/)).toBeVisible();

	// Sin sesión, el panel manda a ingresar.
	await page.goto('/panel');
	await expect(page).toHaveURL(/\/ingresar$/);

	// Contraseña incorrecta.
	await listo(page);
	await page.getByRole('textbox', { name: 'Email' }).fill(email);
	await page.getByLabel('Contraseña', { exact: true }).fill('otra-clave');
	await page.getByRole('button', { name: 'Ingresar', exact: true }).click();
	await expect(page.getByText('El email o la contraseña no son correctos.')).toBeVisible();

	// Recuperar: llega el mail con el link.
	await page.getByRole('link', { name: 'Olvidé mi contraseña' }).click();
	await page.getByRole('textbox', { name: 'Email de tu cuenta' }).fill(email);
	await page.getByRole('button', { name: 'Mandame el link' }).click();
	await expect(page.getByRole('heading', { level: 1 })).toHaveText('Revisá tu email');
	const texto = await ultimoMail(email, /m=restablecer&token=/);
	const link = texto.match(/https?:\/\/\S+m=restablecer&token=[A-Za-z0-9_-]+/)?.[0];
	expect(link).toBeTruthy();

	// Contraseña nueva desde el link (el sitio del mail apunta al front de desarrollo).
	await page.goto(new URL(link!).pathname + new URL(link!).search);
	await listo(page);
	await expect(page.getByRole('heading', { level: 1 })).toHaveText('Elegí tu nueva contraseña');
	await page.getByLabel('Contraseña nueva', { exact: true }).fill('Nuevaclave123');
	await page.getByRole('button', { name: 'Guardar contraseña' }).click();
	await expect(page.getByText('Listo, tu contraseña cambió.')).toBeVisible();

	// El link se usa una sola vez.
	await page.goto(new URL(link!).pathname + new URL(link!).search);
	await listo(page);
	await page.getByLabel('Contraseña nueva', { exact: true }).fill('Otraclave1234');
	await page.getByRole('button', { name: 'Guardar contraseña' }).click();
	await expect(page.getByText(/El link venció o ya se usó/)).toBeVisible();

	// Ingreso con la nueva y salida.
	await page.goto('/ingresar');
	await listo(page);
	await page.getByRole('textbox', { name: 'Email' }).fill(email);
	await page.getByLabel('Contraseña', { exact: true }).fill('Nuevaclave123');
	await page.getByRole('button', { name: 'Ingresar', exact: true }).click();
	await expect(page).toHaveURL(/\/panel$/);
	await page.request.post('/salir', { headers: { origin: 'http://localhost:4174' } });
	await page.goto('/panel');
	await expect(page).toHaveURL(/\/ingresar$/);
});

test('ingreso con proveedores apagados y link por WhatsApp', async ({ page }) => {
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	// Sin la clave publicable de Clerk, no se ofrecen Google ni Apple (docs/decisiones.md §31).
	await page.goto('/ingresar');
	await listo(page);
	await expect(page.getByRole('button', { name: 'Seguir con Google' })).toHaveCount(0);

	// WhatsApp: en desarrollo la API lo manda al log; responde igual haya o no cuenta con ese número.
	await page.goto('/ingresar?m=whatsapp');
	await listo(page);
	await page.getByLabel('WhatsApp de tu cuenta').fill('11 5555-0000');
	await page.getByRole('button', { name: 'Mandame el link' }).click();
	await expect(page.getByRole('heading', { name: 'Revisá tu WhatsApp' })).toBeVisible();

	// Un link inventado no entra.
	await page.goto('/ingresar?m=enlace&token=inventado');
	await listo(page);
	await page.getByRole('button', { name: 'Entrar a mi panel' }).click();
	await expect(page.getByRole('alert')).toContainText(/venció o ya se usó/);
});
