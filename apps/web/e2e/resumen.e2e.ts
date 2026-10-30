import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });
const sinDesborde = (page: Page) =>
	page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);

async function demo(page: Page) {
	await page.addInitScript(() => localStorage.setItem('mostrador-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	await listo(page);
}

test('el Resumen muestra los números de la tienda y cambia de período', async ({ page }) => {
	await demo(page);
	await expect(page.getByRole('heading', { level: 1 })).toContainText('Valeria');
	await expect(page.getByText('Ventas · últimos 7 días')).toBeVisible();
	await expect(page.locator('.pedido').first()).toContainText('#1048');
	await expect(page.getByText('Meta de', { exact: false }).first()).toBeVisible();
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);

	await page.getByRole('group', { name: 'Período' }).getByRole('link', { name: '30 días' }).click();
	await expect(page).toHaveURL(/periodo=30d/);
	await expect(page.getByText('Ventas · últimos 30 días')).toBeVisible();
	await expect(page.locator('.kpi').first()).toContainText('vs. mes anterior');
	await page.getByRole('group', { name: 'Período' }).getByRole('link', { name: 'Hoy' }).click();
	await expect(page.locator('.kpi').first()).toContainText('vs. ayer');
});

test('la meta se edita y queda guardada', async ({ page }) => {
	await demo(page);
	await page.locator('.meta').getByRole('button', { name: 'Editar' }).click();
	const dialogo = page.getByRole('dialog');
	await dialogo.getByLabel('Meta en pesos').fill('abc');
	await dialogo.getByRole('button', { name: 'Guardar meta' }).click();
	await expect(dialogo.getByText('Escribí la meta en pesos')).toBeVisible();
	await dialogo.getByLabel('Meta en pesos').fill('$ 50.000.000');
	await dialogo.getByRole('button', { name: 'Guardar meta' }).click();
	await expect(dialogo).toBeHidden();
	await expect(page.getByText('Meta guardada')).toBeVisible();
	await page.reload();
	await listo(page);
	await expect(page.locator('.meta')).toContainText('$ 50 M');
});

test('el aviso de ARCA se descarta y Tincho manda el recordatorio una vez', async ({ page }) => {
	await demo(page);
	await expect(page.getByText(/Tu certificado de ARCA vence/)).toBeVisible();
	await page.getByRole('button', { name: 'Descartar el aviso' }).click();
	await expect(page.getByText(/Tu certificado de ARCA vence/)).toBeHidden();

	await page.getByRole('button', { name: 'Sí, enviar' }).click();
	await expect(page.getByText('Recordatorios enviados')).toBeVisible();
	await expect(page.getByText(/12 mensajes en cola/)).toBeVisible();
	await expect(page.getByRole('button', { name: 'Sí, enviar' })).toHaveCount(0);
	await expect(page.locator('.burbuja')).toContainText('Listo');
	await page.reload();
	await listo(page);
	await expect(page.getByText(/Tu certificado de ARCA vence/)).toHaveCount(0);
	await expect(page.locator('.burbuja')).toContainText('Listo');
});

test('los accesos y las tarjetas llevan a su sección', async ({ page }) => {
	await demo(page);
	await page.getByRole('link', { name: /Ver todos/ }).click();
	await expect(page).toHaveURL(/\/panel\/pedidos$/, { timeout: 15_000 });
	await page.goto('/panel');
	await listo(page);
	await page.getByRole('button', { name: 'Nuevo' }).click();
	await expect(page.getByRole('dialog', { name: 'Buscar' })).toBeVisible();
});
