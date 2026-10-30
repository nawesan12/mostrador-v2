import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });

test('finanzas: transferir deja en cero lo disponible y suma el movimiento', async ({ page }) => {
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	await page.goto('/panel/finanzas');
	await listo(page);
	await expect(page.getByText('Conectado con Caja')).toBeVisible();
	await expect(page.locator('.barras .col')).toHaveCount(14);
	const antes = (await page.locator('.d-v').textContent())!;
	expect(Number(antes.replace(/\D/g, ''))).toBeGreaterThan(0);
	await page.getByRole('button', { name: 'Transferir a mi banco' }).click();
	await expect(page.getByText('Transferencia en camino')).toBeVisible();
	await expect(page.locator('.d-v')).toHaveText('$ 0');
	await expect(page.locator('.mov').filter({ hasText: 'Transferencia a tu banco' })).toContainText(
		antes.trim()
	);
	await page.getByRole('button', { name: 'Transferir a mi banco' }).click();
	await expect(page.getByText('No hay nada para transferir')).toBeVisible();
});
