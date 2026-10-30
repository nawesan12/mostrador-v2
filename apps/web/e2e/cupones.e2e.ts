import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });

async function demo(page: Page) {
	await page.addInitScript(() => localStorage.setItem('mostrador-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	await page.goto('/panel/cupones');
	await listo(page);
}

test('crear un cupón con el armador, pausarlo y duplicarlo', async ({ page }) => {
	await demo(page);
	await page.getByRole('button', { name: 'Nuevo cupón' }).click();
	await page.getByRole('button', { name: /^Que vuelvan/ }).click();
	await expect(page.getByLabel('Código', { exact: true })).toHaveValue('VOLVE15');
	await page.getByLabel('Código', { exact: true }).fill('prueba 20');
	await expect(page.getByLabel('Código', { exact: true })).toHaveValue('PRUEBA20');
	await page
		.getByRole('group', { name: 'Tipo de descuento' })
		.getByRole('button', { name: 'Porcentaje' })
		.click();
	await expect(page.locator('.tk-h')).toHaveText('15 % off');
	await expect(page.locator('.calcula')).toContainText('Margen que te queda');
	await page.getByRole('button', { name: 'Crear PRUEBA20' }).click();
	await expect(page.getByText('PRUEBA20 está activo')).toBeVisible();

	const cupon = page.locator('.cupon').filter({ hasText: 'PRUEBA20' }).first();
	await expect(cupon).toContainText('Activo');
	await cupon.getByRole('switch', { name: 'Activar o pausar PRUEBA20' }).click();
	await expect(cupon).toContainText('Pausado por vos');
	await cupon.getByRole('button', { name: 'Duplicar PRUEBA20' }).click();
	await expect(page.getByText('PRUEBA20-2 quedó pausado para que lo edites')).toBeVisible();

	// Repetir un código activo no se puede.
	await page.getByRole('button', { name: 'Nuevo cupón' }).click();
	await page.getByLabel('Código', { exact: true }).fill('FINDE10');
	await page.getByRole('button', { name: 'Crear FINDE10' }).click();
	await expect(page.getByText('Ya tenés un cupón FINDE10.')).toBeVisible();
});

test('consejo de Tincho, automáticas y relámpago', async ({ page }) => {
	await demo(page);
	await page.getByRole('button', { name: 'Poner mínimo $ 25.000' }).click();
	await expect(page.getByText('BIENVENIDA actualizado')).toBeVisible();
	await expect(page.locator('.consejo')).toHaveCount(0);
	await expect(page.locator('.cupon').filter({ hasText: 'BIENVENIDA' })).toContainText(
		'mín. $ 25.000'
	);

	await page
		.getByRole('group', { name: 'Ver' })
		.getByRole('button', { name: /Automáticas/ })
		.click();
	const cumple = page.locator('.auto').filter({ hasText: 'Cumpleaños' });
	await expect(cumple).toContainText('Apagada');
	await cumple.getByRole('switch').click();
	await expect(cumple).not.toContainText('Apagada');

	await page
		.getByRole('group', { name: 'Ver' })
		.getByRole('button', { name: /Relámpago/ })
		.click();
	await expect(page.locator('.v-et')).toContainText('EN VIVO');
	await page.getByRole('button', { name: 'Liberar 20 más' }).click();
	await expect(page.locator('.vb-t')).toContainText('de 70');
	await page.getByRole('button', { name: 'Terminar ahora' }).click();
	await expect(page.locator('.v-et')).toContainText('TERMINÓ');
	await page.getByRole('button', { name: 'Deshacer' }).click();
	await expect(page.locator('.v-et')).toContainText('EN VIVO');
});
