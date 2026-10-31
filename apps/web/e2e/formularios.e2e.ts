import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });
const sinDesborde = (page: Page) =>
	page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);

test.beforeEach(async ({ page }) => {
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
});

test('contacto: valida, conserva lo escrito y saluda por el nombre', async ({ page }) => {
	await page.goto('/contacto');
	await listo(page);
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	await page.getByLabel('Nombre').fill('Lucía Fernández');
	await page.getByLabel('Email').fill('lucia');
	await page.getByLabel('WhatsApp').fill('123');
	await page.getByText('$ 5 a 20 M').click();
	await page.getByText('Tiendanube', { exact: true }).click();
	await page.getByRole('button', { name: 'Quiero que me contacten' }).click();
	await expect(page.getByText('Revisá el email.')).toBeVisible();
	await expect(page.getByText('Revisá el número: con característica, sin 0 ni 15.')).toBeVisible();
	await expect(page.getByLabel('Nombre')).toHaveValue('Lucía Fernández');
	await expect(page.getByRole('radio', { name: '$ 5 a 20 M' })).toBeChecked();

	await page.getByLabel('Email').fill('lucia@tutienda.com.ar');
	await page.getByLabel('WhatsApp').fill('11 5555-0000');
	await page.getByRole('button', { name: 'Quiero que me contacten' }).click();
	await expect(page.getByText('¡Listo, Lucía!')).toBeVisible();
	await page.getByRole('button', { name: 'Enviar otra consulta' }).click();
	await expect(page.getByRole('button', { name: 'Quiero que me contacten' })).toBeVisible();
});

test('partners: postulación con especialidades', async ({ page }) => {
	await page.goto('/partners');
	await listo(page);
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	await page.getByRole('link', { name: 'Quiero ser partner' }).click();
	await page.getByRole('button', { name: 'Postularme' }).click();
	await expect(page.getByText('¿Cómo se llama tu agencia o estudio?')).toBeVisible();
	await page.getByLabel('Nombre').fill('Juan Gómez');
	await page.getByLabel('Email').fill('juan@agencia.com');
	await page.getByLabel('Agencia o estudio').fill('Estudio Brote');
	await page.getByText('Diseño', { exact: true }).click();
	await page.getByText('Desarrollo', { exact: true }).click();
	await page.getByRole('button', { name: 'Postularme' }).click();
	await expect(page.getByText('¡Bienvenido al programa!')).toBeVisible();
});

test('novedades: filtra por tipo y suma el email', async ({ page }) => {
	await page.goto('/novedades');
	await listo(page);
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	// Las mismas que lib/sitio/novedades.ts: todas, y después sólo los arreglos.
	await expect(page.locator('article')).toHaveCount(6);
	await page.getByRole('button', { name: 'Arreglo', exact: true }).click();
	await expect(page.locator('article')).toHaveCount(1);
	await expect(page.locator('article')).toContainText('Factura A con CUIT extranjero');
	await expect(page.getByText('Agosto 2026')).toHaveCount(0);
	await page.getByLabel('Tu email').fill('nada');
	await page.getByRole('button', { name: 'Sumarme' }).click();
	await expect(page.getByText('Revisá el email.')).toBeVisible();
	await page.getByLabel('Tu email').fill('lucia@tutienda.com.ar');
	await page.getByRole('button', { name: 'Sumarme' }).click();
	await expect(page.getByText('Listo, te sumamos.')).toBeVisible();
});

test('la baja de novedades pide confirmar con un botón', async ({ page }) => {
	await page.goto('/novedades/baja?token=abc');
	await listo(page);
	await expect(page.locator('h1')).toHaveText('¿Te damos de baja?');
	await page.getByRole('button', { name: 'Darme de baja' }).click();
	await expect(page.locator('h1')).toHaveText('Listo, no te escribimos más.');
});
