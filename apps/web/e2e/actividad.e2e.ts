import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });

async function demo(page: Page) {
	await page.addInitScript(() => localStorage.setItem('mostrador-cookies', 'nec'));
	await page.goto('/demo');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
}

test('actividad: lo que hizo el visitante aparece y se deshace', async ({ page }) => {
	await demo(page);
	await page.goto('/panel/cupones');
	await listo(page);
	const cupon = page.locator('.cupon').filter({ hasText: 'FINDE10' });
	await cupon.getByRole('switch').click();
	await expect(cupon).toContainText('Pausado por vos');

	await page.goto('/panel/actividad');
	await listo(page);
	const evento = page.locator('.evento').filter({ hasText: 'pausó el cupón FINDE10' });
	await expect(evento).toBeVisible();
	await page.getByRole('group', { name: 'Ver' }).getByRole('button', { name: 'Sistema' }).click();
	await expect(evento).toHaveCount(0);
	await page.getByRole('group', { name: 'Ver' }).getByRole('button', { name: 'Todo' }).click();
	await evento.getByRole('button', { name: 'Deshacer' }).click();
	await expect(page.getByText('Cambio deshecho')).toBeVisible();
	await expect(
		page.locator('.evento').filter({ hasText: 'pausó el cupón FINDE10 — deshecho' })
	).toBeVisible();

	await page.goto('/panel/cupones');
	await listo(page);
	await expect(page.locator('.cupon').filter({ hasText: 'FINDE10' })).toContainText('Activo');
});

test('para hacer hoy: marcar suma al anillo y la insignia baja', async ({ page, isMobile }) => {
	await demo(page);
	await page.goto('/panel/para-hacer-hoy');
	await listo(page);
	const total = await page.locator('.tarea').count();
	await expect(page.locator('.cuenta')).toHaveText(`0/${total}`);
	await page.getByRole('checkbox', { name: /Marcar Renovar certificado/ }).click();
	await expect(page.locator('.cuenta')).toHaveText(`1/${total}`);
	if (!isMobile)
		await expect(page.locator('.barra a[aria-current="page"]')).toContainText(String(total - 1));
	await page.getByRole('button', { name: 'Preparar' }).click();
	await expect(page).toHaveURL(/\/panel\/pedidos\?estado=para-preparar/);
	await page.goBack();
	await expect(page.locator('.cuenta')).toHaveText(`2/${total}`);
});
