import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });
const sinDesborde = (page: Page) =>
	page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);

/** Arma la demo de Litoral (Mate & Co) y entra al panel. */
async function demo(page: Page) {
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	await listo(page);
}

test('el armazón muestra la tienda de la demo y navega entre secciones', async ({
	page,
	isMobile
}) => {
	await demo(page);
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	await expect(page.locator('.barra .t-n')).toHaveText('Mate & Co');
	if (!isMobile) {
		await page.getByRole('link', { name: 'Ventas', exact: true }).click();
		await expect(page).toHaveURL(/\/panel\/pedidos$/);
	}
	// Ir directo a una sección: se ve y queda marcada en la barra.
	await page.goto('/panel/mis-datos');
	await expect(page.getByRole('heading', { name: 'Mis datos' })).toBeVisible();
	await expect(page.locator('.barra a[aria-current="page"]')).toContainText('Mis datos');
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
});

test('una sección que no existe da 404', async ({ page }) => {
	await demo(page);
	const res = await page.goto('/panel/no-existe');
	expect(res?.status()).toBe(404);
});

test('el tema se recuerda', async ({ page }) => {
	await demo(page);
	await page.getByRole('button', { name: 'Pasar a tema oscuro' }).click();
	await expect(page.locator('.panel')).toHaveAttribute('data-theme', 'dark');
	await page.reload();
	await listo(page);
	await expect(page.locator('.panel')).toHaveAttribute('data-theme', 'dark');
});

test('notificaciones: filtra, tocar una la marca y marca todo como leído', async ({ page }) => {
	await demo(page);
	const campanita = page.locator('header.top').getByRole('button', { name: /^Notificaciones/ });
	// La insignia con el número sin leer (§45).
	await expect(campanita).toHaveAccessibleName('Notificaciones, 6 sin leer');
	await expect(campanita.locator('.badge')).toHaveText('6');
	await campanita.click();
	const caja = page.getByRole('dialog', { name: 'Notificaciones' });
	await expect(caja.locator('.n')).toHaveCount(6);
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	const caja1 = await caja.boundingBox();
	const vista = page.viewportSize()!;
	expect(caja1!.x + caja1!.width).toBeLessThanOrEqual(vista.width);
	// Tocar una: se cierra, queda leída y la insignia baja.
	await caja
		.locator('.n')
		.filter({ hasText: /Tincho ya respondió/ })
		.click();
	await expect(caja).toBeHidden();
	await expect(campanita).toHaveAccessibleName('Notificaciones, 5 sin leer');
	await campanita.click();
	await expect(
		caja
			.locator('.n')
			.filter({ hasText: /Tincho ya respondió/ })
			.getByLabel('Sin leer')
	).toHaveCount(0);
	await caja.getByRole('button', { name: 'Sistema' }).click();
	await expect(caja.locator('.n')).toHaveCount(2);
	await caja.getByRole('button', { name: 'Marcar todo como leído' }).click();
	await expect(caja.getByLabel('Sin leer')).toHaveCount(0);
	await expect(caja.getByRole('button', { name: 'Marcar todo como leído' })).toHaveCount(0);
	await page.reload();
	await listo(page);
	await expect(campanita).toHaveAccessibleName('Notificaciones');
	await expect(campanita.locator('.badge')).toHaveCount(0);
});

test('⌘K busca y lleva a la sección', async ({ page }) => {
	await demo(page);
	await page.keyboard.press('ControlOrMeta+k');
	const cmdk = page.getByRole('dialog', { name: 'Buscar' });
	await expect(cmdk.getByRole('textbox', { name: 'Buscar' })).toBeFocused();
	await page.keyboard.type('inform');
	// Tincho siempre arriba; la sección que coincide queda elegida.
	await expect(cmdk.getByRole('option')).toHaveCount(2);
	await expect(cmdk.getByRole('option', { name: /Informes/ })).toHaveAttribute(
		'aria-selected',
		'true'
	);
	await page.keyboard.press('Enter');
	await expect(page).toHaveURL(/\/panel\/informes$/);
	await expect(cmdk).toBeHidden();
	// Sin coincidencias, Enter se lo pregunta a Tincho.
	await page.keyboard.press('ControlOrMeta+k');
	await expect(cmdk.getByRole('textbox', { name: 'Buscar' })).toBeFocused();
	await page.keyboard.type('qué me conviene vender');
	await page.keyboard.press('Enter');
	await expect(page.getByRole('dialog', { name: 'Tincho' })).toBeVisible();
});

test.describe('celular', () => {
	test.skip(({ isMobile }) => !isMobile, 'sólo en mobile');

	test('barra de abajo y "Más" abre las secciones; Escape la cierra', async ({ page }) => {
		await demo(page);
		const abajo = page.getByRole('navigation', { name: 'Accesos' });
		await expect(abajo).toBeVisible();
		await abajo.getByRole('link', { name: /Pedidos/ }).click();
		await expect(page).toHaveURL(/\/panel\/pedidos$/);
		await abajo.getByRole('button', { name: 'Más' }).click();
		await expect(page.locator('.barra')).toHaveClass(/abierta/);
		await page.keyboard.press('Escape');
		await expect(page.locator('.barra')).not.toHaveClass(/abierta/);
	});
});
