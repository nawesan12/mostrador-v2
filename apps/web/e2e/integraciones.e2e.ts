import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });
const sinDesborde = (page: Page) =>
	page.evaluate(() => document.documentElement.scrollWidth - innerWidth);

async function demo(page: Page, ruta: string) {
	await page.addInitScript(() => localStorage.setItem('mostrador-cookies', 'nec'));
	await page.goto('/demo');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	await page.goto(ruta);
	await listo(page);
}

test('integraciones: conectar, configurar y desconectar', async ({ page }) => {
	await demo(page, '/panel/integraciones');
	await expect(page.getByRole('heading', { name: 'Integraciones' })).toBeVisible();
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);

	// Filtrar por categoría.
	await page.getByRole('button', { name: 'Envíos', exact: true }).click();
	await expect(page.locator('.tarjeta')).toHaveCount(3);
	await page.getByRole('button', { name: 'Todas', exact: true }).click();

	// Conectar una disponible.
	const oca = page.locator('.tarjeta', { hasText: 'OCA' });
	await oca.getByRole('button', { name: 'Conectar' }).click();
	await expect(oca.getByText('● Conectada')).toBeVisible();

	// Mercado Pago: configurar muestra la cuenta; desconectar y volver a conectar (en la demo vuelve sola).
	const mp = page.locator('.tarjeta', { hasText: 'Mercado Pago' });
	await mp.getByRole('button', { name: 'Configurar' }).click();
	await expect(page.getByRole('dialog')).toContainText('no cobra comisión');
	await page.getByRole('dialog').getByRole('button', { name: 'Desconectar' }).click();
	await expect(mp.getByRole('button', { name: 'Conectar' })).toBeVisible();
	await mp.getByRole('button', { name: 'Conectar' }).click();
	await expect(page).toHaveURL(/\/panel\/integraciones$/);
	await expect(page.getByRole('status').filter({ hasText: 'Mercado Pago conectado' })).toHaveCount(
		1
	);
	await expect(mp.getByRole('button', { name: 'Configurar' })).toBeVisible();

	// IA: conectar Claude y prender un permiso.
	await page.getByRole('button', { name: /Claude/ }).click();
	await expect(page.getByRole('button', { name: /Claude/ })).toContainText('Conectado', {
		timeout: 5000
	});
	const responder = page.getByRole('switch', { name: /Responder mensajes/ });
	await responder.click();
	await expect(responder).toHaveAttribute('aria-checked', 'true');
});
