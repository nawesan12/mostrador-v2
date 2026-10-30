import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });
const sinDesborde = (page: Page) =>
	page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);

async function demo(page: Page, ruta: string) {
	await page.addInitScript(() => localStorage.setItem('mostrador-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	await page.goto(ruta);
	await listo(page);
}

test('tarjeta A6: se arma con vista previa en vivo, se guarda y se imprime por pedido', async ({
	page
}) => {
	await demo(page, '/panel/tarjeta-para-paquetes');
	await expect(page.getByRole('heading', { name: 'Tarjeta para paquetes' })).toBeVisible();
	const vista = page.getByRole('region', { name: 'Vista previa' });
	await expect(vista).toContainText('¡Gracias, Lucía!');
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);

	// Texto con variable, otro modelo y el QR del dorso a Instagram.
	await page
		.getByLabel(/^Título/)
		.first()
		.fill('¡Gracias,');
	await page
		.getByRole('group', { name: 'Sumar un dato del pedido' })
		.getByRole('button', { name: '{nombre}' })
		.click();
	await expect(vista).toContainText('¡Gracias, Lucía');
	await page.getByRole('button', { name: 'Color de la tienda' }).click();
	await page.getByRole('button', { name: 'Seguir en Instagram' }).click();
	await expect(page.getByText('Poné la cuenta de Instagram.')).toBeVisible();
	await page.getByLabel('Cuenta de Instagram').fill('mateyco');
	await expect(vista).toContainText('@mateyco');
	await page.getByRole('button', { name: 'Guardar' }).click();
	await expect(page.getByText('Tarjeta guardada')).toBeVisible();

	// Lo guardado sigue al volver.
	await page.reload();
	await listo(page);
	await expect(page.getByLabel('Cuenta de Instagram')).toHaveValue('mateyco');

	// Imprimir: la del pedido lleva el nombre de quien compró; en lote, 4 por A4 con los dorsos.
	await page.goto('/panel/pedidos/imprimir?n=1048');
	await listo(page);
	await expect(page.getByText('1 tarjeta')).toBeVisible();
	await expect(page.locator('.hoja.a6')).toHaveCount(2);
	await expect(page.locator('.hoja').first()).toContainText('¡Gracias, ');
	await page.goto('/panel/pedidos/imprimir?n=1048,1047,1046,1045,1044');
	await listo(page);
	await expect(page.getByText('5 tarjetas')).toBeVisible();
	await expect(page.locator('.hoja.a4')).toHaveCount(4);
	await page.getByRole('button', { name: 'Una por hoja A6' }).click();
	await expect(page).toHaveURL(/f=a6/);
	await expect(page.locator('.hoja.a6')).toHaveCount(10);
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
});
