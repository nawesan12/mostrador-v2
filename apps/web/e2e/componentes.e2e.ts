import { expect, test } from '@playwright/test';

const sinDesborde = (page: import('@playwright/test').Page) =>
	page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);

test.describe('/componentes', () => {
	test.beforeEach(async ({ page }) => {
		await page.goto('/componentes');
		await expect(page.getByRole('heading', { level: 1 })).toHaveText('Avisos, modales y hojas');
	});

	test('sin desborde horizontal', async ({ page }) => {
		expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	});

	test('el toast aparece, su acción abre la hoja y Esc la cierra', async ({ page }) => {
		await page.getByRole('button', { name: 'Toast', exact: true }).click();
		const toast = page.getByRole('status').filter({ hasText: 'Nuevo pedido por WhatsApp' });
		await expect(toast).toBeVisible();
		await toast.getByRole('button', { name: 'Ver pedido' }).click();
		await expect(toast).toHaveCount(0);
		const hoja = page.getByRole('dialog', { name: 'Pedido #1048' });
		await expect(hoja).toBeVisible();
		await page.keyboard.press('Escape');
		await expect(hoja).toHaveCount(0);
	});

	test('máximo 3 toasts a la vez', async ({ page }) => {
		const boton = page.getByRole('button', { name: 'Toast', exact: true });
		for (let i = 0; i < 5; i++) await boton.click();
		await expect(page.locator('.pila .toast')).toHaveCount(3);
	});

	test('la confirmación pausa y ofrece deshacer', async ({ page }) => {
		await page.getByRole('button', { name: 'Confirmación' }).click();
		const modal = page.getByRole('dialog', { name: '¿Pausar 4 productos sin stock?' });
		await expect(modal).toBeVisible();
		const ml = modal.getByRole('switch', { name: 'También pausar en Mercado Libre' });
		await expect(ml).toHaveAttribute('aria-checked', 'true');
		await ml.click();
		await expect(ml).toHaveAttribute('aria-checked', 'false');
		await modal.getByRole('button', { name: 'Pausar productos' }).click();
		await expect(modal).toHaveCount(0);
		const hecho = page.getByRole('status').filter({ hasText: '4 productos pausados' });
		await expect(hecho).toContainText('Sólo en tu tienda');
		await hecho.getByRole('button', { name: 'Deshacer' }).click();
		await expect(
			page.getByRole('status').filter({ hasText: 'volvieron a estar a la venta' })
		).toBeVisible();
	});

	test('el tema oscuro cambia las variables', async ({ page }) => {
		const fondo = () =>
			page.locator('.pagina').evaluate((el) => getComputedStyle(el).backgroundColor);
		const claro = await fondo();
		await page.getByRole('group', { name: 'Tema' }).getByRole('button', { name: 'Oscuro' }).click();
		await expect.poll(fondo).not.toBe(claro);
		expect(await fondo()).toBe('rgb(34, 20, 13)');
	});
});
