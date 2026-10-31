import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });
const sinDesborde = (page: Page) =>
	page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);
const seguir = (page: Page) => page.locator('.pie .seguir').click();

test.beforeEach(async ({ page }) => {
	await page.addInitScript(() => {
		localStorage.setItem('tlw-cookies', 'nec');
	});
});

test('el onboarding se recorre entero y termina en el panel con la demo armada', async ({
	page
}) => {
	await page.goto('/onboarding');
	await listo(page);
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	await expect(page.locator('h1').first()).toHaveText('Abrí Tu Local Web. No arrancás de cero.');

	// 1 → 2: en la demo, cualquier botón de cuenta sigue.
	await page.getByRole('button', { name: /Continuar con Mercado Pago/ }).click();
	await expect(page.getByRole('heading', { name: /Encontramos tu negocio/ })).toBeVisible();
	await expect(page.getByText('Valeria Sosa')).toBeVisible({ timeout: 5000 });
	const objetivo = page.getByRole('button', { name: 'Facturar sin complicarme' });
	await objetivo.click();
	await expect(objetivo).toHaveAttribute('aria-pressed', 'true');

	// 3: conectar.
	await seguir(page);
	await expect(page.getByRole('heading', { name: /Conectá lo que ya usás/ })).toBeVisible();
	await page.getByRole('button', { name: 'Conectar las recomendadas' }).click();
	await expect(page.getByText('5 de 8 conectadas.', { exact: false })).toBeVisible({
		timeout: 6000
	});

	// 4: importación y conflicto de precios.
	await seguir(page);
	await expect(
		page.getByRole('heading', { name: 'Listo. Tu negocio ya está adentro.' })
	).toBeVisible({
		timeout: 12000
	});
	await page.getByRole('button', { name: 'Usar Mercado Libre (+10%)' }).click();
	await expect(page.getByText(/mantenemos tu recargo del 10%/)).toBeVisible();

	// 5: plantilla desde la galería.
	await seguir(page);
	await page.getByRole('button', { name: /Ver las \d+ plantillas/ }).click();
	const gal = page.getByRole('dialog');
	await gal.getByRole('button', { name: /Moda/ }).first().click();
	await gal.getByRole('button', { name: /^Ceibo/ }).click();
	await expect(gal).toBeHidden();

	// 6 y 7.
	await seguir(page);
	await page.getByRole('switch', { name: 'Moto propia en Rosario' }).click();
	await expect(page.getByRole('switch', { name: 'Moto propia en Rosario' })).toHaveAttribute(
		'aria-checked',
		'true'
	);
	await seguir(page);
	await page.getByRole('button', { name: 'Canchero' }).click();
	await expect(page.getByText(/Holaa Lu!/)).toBeVisible();

	// 8: listo → panel.
	await page.locator('.pie .seguir').click();
	await expect(page.getByRole('heading', { name: 'Tu Local Web ya está listo.' })).toBeVisible();
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	await page.getByRole('button', { name: /Ir a mi panel/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	await expect(page.locator('.barra .t-n')).toHaveText('Tienda Lumbre');
});

test('lo elegido sobrevive a recargar y "Atrás" vuelve un paso', async ({ page }) => {
	await page.goto('/onboarding');
	await listo(page);
	await page.getByRole('button', { name: /Continuar con Google/ }).click();
	await seguir(page);
	await expect(page.getByRole('heading', { name: /Conectá lo que ya usás/ })).toBeVisible();
	await page.reload();
	await listo(page);
	await expect(page.getByRole('heading', { name: /Conectá lo que ya usás/ })).toBeVisible();
	await page.locator('.pie .atras').click();
	await expect(page.getByRole('heading', { name: /Encontramos tu negocio/ })).toBeVisible();
});

test('"Usar esta plantilla" de la landing arranca el onboarding con esa plantilla', async ({
	page
}) => {
	await page.goto('/');
	await listo(page);
	await page.getByRole('button', { name: 'Ver plantillas' }).click();
	await page
		.getByRole('dialog')
		.getByRole('link', { name: /Usar esta plantilla/ })
		.click();
	await expect(page).toHaveURL(/\/onboarding(\?|$)/);
	await listo(page);
	const tpl = await page.evaluate(() => localStorage.getItem('ob-tpl'));
	expect(tpl).toBeTruthy();
});
