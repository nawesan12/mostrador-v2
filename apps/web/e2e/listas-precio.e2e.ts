import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });

test('aumento por inflación: aplicar, deshacer y programar', async ({ page }) => {
	await page.addInitScript(() => localStorage.setItem('mostrador-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	await page.goto('/panel/listas-de-precio');
	await listo(page);

	const fila = page.locator('.tabla .fila:not(.cab-t)').first();
	const hoy = await fila.locator('.t-a').textContent();
	await page.getByRole('button', { name: /^Mayorista/ }).click();
	await expect(page.locator('.cab-t')).toContainText('Nuevo · Mayorista');
	await page.getByRole('button', { name: /^Minorista/ }).click();
	await page.getByRole('button', { name: 'Múltiplo de 500' }).click();
	const nuevo = (await fila.locator('b').textContent())!;
	expect(Number(nuevo.replace(/\D/g, '')) % 500).toBe(0);

	await page.getByRole('button', { name: /^Aplicar a \d+ productos$/ }).click();
	await expect(page.getByText('Precios actualizados +8%')).toBeVisible();
	await expect(fila.locator('.t-a')).toHaveText(nuevo);
	await page.getByRole('button', { name: 'Deshacer' }).click();
	await expect(fila.locator('.t-a')).toHaveText(hoy!);

	await page.getByRole('button', { name: 'Programar para el lunes' }).click();
	await expect(page.getByText('Aumento programado')).toBeVisible();
	await expect(page.locator('.prog')).toContainText('Programado: +8% el lunes');
	await page.locator('.prog').getByRole('button', { name: 'Cancelar' }).click();
	await expect(page.locator('.prog')).toHaveCount(0);
});
