import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });
const sinDesborde = (page: Page) =>
	page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);

test.beforeEach(async ({ page }) => {
	await page.addInitScript(() => localStorage.setItem('mostrador-cookies', 'nec'));
});

const PAGINAS = [
	['terminos', 'Términos y condiciones'],
	['privacidad', 'Política de privacidad'],
	['cookies', 'Política de cookies'],
	['arrepentimiento', 'Botón de arrepentimiento'],
	['defensa-del-consumidor', 'Defensa de las y los consumidores'],
	['tratamiento-de-datos', 'Acuerdo de tratamiento de datos'],
	['uso-aceptable', 'Uso aceptable y productos prohibidos'],
	['reembolsos', 'Política de reembolsos de planes']
];

for (const [slug, titulo] of PAGINAS) {
	test(`/legales/${slug} carga sin desborde`, async ({ page }) => {
		const res = await page.goto(`/legales/${slug}`);
		expect(res?.status()).toBe(200);
		await expect(page.locator('h1')).toHaveText(titulo);
		await expect(page).toHaveTitle(`${titulo} · Mostrador`);
		expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	});
}

test('/legales lleva a términos y una página inexistente da 404', async ({ page }) => {
	await page.goto('/legales');
	await expect(page).toHaveURL(/\/legales\/terminos$/);
	const res = await page.goto('/legales/no-existe');
	expect(res?.status()).toBe(404);
});

test('botón de arrepentimiento: valida y da el código de trámite', async ({ page }) => {
	await page.goto('/legales/arrepentimiento');
	await listo(page);
	const form = page.locator('form[action="?/revocar"]');
	await form.getByLabel('Email de la cuenta').fill('lucia');
	await form.getByLabel('Nombre y apellido').fill('Lucía Fernández');
	await form.getByLabel('CUIT o DNI').fill('123');
	await form.getByLabel('Plan contratado').fill('Pro mensual');
	await form.getByRole('button', { name: 'Enviar pedido de revocación' }).click();
	await expect(page.getByText('Revisá el email: tiene que ser el de la cuenta.')).toBeVisible();
	await expect(page.getByText('Poné tu DNI o tu CUIT, con o sin guiones.')).toBeVisible();
	// Lo escrito no se pierde.
	await expect(form.getByLabel('Nombre y apellido')).toHaveValue('Lucía Fernández');

	await form.getByLabel('Email de la cuenta').fill('lucia@laesquina.com.ar');
	await form.getByLabel('CUIT o DNI').fill('27-38412905-4');
	await form.getByRole('button', { name: 'Enviar pedido de revocación' }).click();
	await expect(page.getByText('Recibimos tu pedido')).toBeVisible();
	await expect(page.locator('.codigo')).toHaveText(/^ARR-\d{4}-\d{6}$/);

	await page.getByRole('button', { name: 'Hacer otro pedido' }).click();
	await expect(form.getByRole('button', { name: 'Enviar pedido de revocación' })).toBeVisible();
});

test('"Configurar cookies" del footer abre el panel de cookies', async ({ page }) => {
	await page.goto('/legales/cookies');
	await listo(page);
	await page.locator('footer').getByRole('button', { name: 'Configurar cookies' }).click();
	await expect(page.getByText('Usamos cookies', { exact: true })).toBeVisible();
	await expect(page.getByText('Medir campañas en Meta y Google')).toBeVisible();
});

test('la lista lateral marca la página actual', async ({ page }) => {
	await page.goto('/legales/uso-aceptable');
	await expect(
		page.getByRole('navigation', { name: 'Páginas legales' }).locator('[aria-current="page"]')
	).toHaveText('Uso aceptable');
});
