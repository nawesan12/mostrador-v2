import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });
const sinDesborde = (page: Page) =>
	page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);

test.beforeEach(async ({ page }) => {
	await page.addInitScript(() => localStorage.setItem('mostrador-cookies', 'nec'));
});

test('/estado muestra los siete componentes y cambia de período', async ({ page }) => {
	const res = await page.goto('/estado');
	expect(res?.status()).toBe(200);
	await listo(page);
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	for (const c of [
		'Tiendas online',
		'Panel de administración',
		'Checkout y pagos',
		'WhatsApp y Hornero (IA)',
		'Facturación electrónica (ARCA)',
		'Sincronización con Mercado Libre',
		'API y webhooks'
	])
		await expect(page.getByText(c, { exact: true })).toBeVisible();
	await expect(page.locator('.barras').first().locator('span')).toHaveCount(90);
	await page.getByRole('link', { name: '7 días' }).click();
	await expect(page).toHaveURL(/dias=7/);
	await expect(page.locator('.barras').first().locator('span')).toHaveCount(7);
	await expect(page.getByRole('link', { name: '7 días' })).toHaveAttribute('aria-current', 'true');
});

test('/estado: suscripción a los avisos y baja', async ({ page }) => {
	await page.goto('/estado');
	await listo(page);
	const campo = page.getByLabel('Tu email o WhatsApp');
	await campo.fill('123');
	await page.getByRole('button', { name: 'Suscribirme' }).click();
	await expect(page.getByText('Poné un email o un WhatsApp con característica.')).toBeVisible();
	await campo.fill('lucia@tutienda.com.ar');
	await page.getByRole('button', { name: 'Suscribirme' }).click();
	await expect(page.getByText('Listo, te avisamos por email.')).toBeVisible();

	await page.goto('/estado/baja?token=nada');
	await listo(page);
	await page.getByRole('button', { name: 'Dejar de recibirlos' }).click();
	await expect(page.locator('h1')).toHaveText('Listo, no te avisamos más.');
});

test('/estado/rss.xml es un RSS válido', async ({ request }) => {
	const res = await request.get('/estado/rss.xml');
	expect(res.status()).toBe(200);
	expect(res.headers()['content-type']).toContain('application/rss+xml');
	expect(await res.text()).toContain('<title>Estado de Mostrador</title>');
});
