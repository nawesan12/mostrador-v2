import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });

test('centro de ayuda: buscar, abrir una sección y un artículo, y preguntarle a Tincho', async ({
	page
}) => {
	await page.addInitScript(() => localStorage.setItem('mostrador-cookies', 'nec'));
	await page.goto('/ayuda');
	await listo(page);
	await page.getByLabel('Buscar en la ayuda').fill('factura');
	await page.getByRole('button', { name: 'Buscar', exact: true }).click();
	await expect(page).toHaveURL(/\/ayuda\?q=factura/);
	await expect(page.getByText(/artículos? para «factura»/)).toBeVisible();
	await page
		.getByRole('link', { name: /Facturación y ARCA/ })
		.first()
		.click();
	await expect(page.getByRole('heading', { name: 'Facturación y ARCA' })).toBeVisible();
	await page.getByRole('link', { name: /Factura A, B o C/ }).click();
	await expect(
		page.getByRole('heading', { name: 'Factura A, B o C: cuál corresponde' })
	).toBeVisible();
	await page.getByRole('button', { name: 'Sí', exact: true }).click();
	await expect(page.getByText('¡Gracias!')).toBeVisible();
	await page.getByLabel('Pregunta para Tincho').fill('¿cómo pongo mi dominio?');
	await page.getByRole('button', { name: 'Preguntar' }).click();
	await expect(page.getByRole('link', { name: /Usar tu propio dominio/ })).toBeVisible();
	expect((await page.goto('/ayuda/no-existe'))?.status()).toBe(404);
});

test('ayuda del panel: las guías llevan al artículo y preguntar abre Tincho', async ({ page }) => {
	await page.addInitScript(() => localStorage.setItem('mostrador-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	await page.goto('/panel/ayuda');
	await listo(page);
	await page.getByLabel('Tu pregunta').fill('¿Cuánto vendí ayer?');
	await page.getByLabel('Tu pregunta').press('Enter');
	await expect(page.getByRole('dialog', { name: /Tincho/ })).toBeVisible();
	await page.keyboard.press('Escape');
	await page.getByRole('link', { name: /Vender en Mercado Libre/ }).click();
	await expect(page).toHaveURL(/\/ayuda\/productos\/mercado-libre$/);
});
