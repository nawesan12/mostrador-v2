import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });
const sinDesborde = (page: Page) =>
	page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);

async function demo(page: Page, ruta: string) {
	await page.addInitScript(() => localStorage.setItem('mostrador-cookies', 'nec'));
	await page.goto('/demo');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	await page.goto(ruta);
	await listo(page);
}

test('inventario: sumar uno, contar y ver el movimiento', async ({ page }) => {
	await demo(page, '/panel/inventario?orden=stock');
	await expect(page.getByText('Valor del inventario')).toBeVisible();
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	const fila = page.locator('.fila').first();
	const nombre = (await fila.locator('.q-t').textContent())!.trim();
	const antes = Number(await fila.locator('.stepper span').textContent());
	await fila.getByRole('button', { name: `Sumar uno a ${nombre}` }).click();
	await expect(
		page.locator('.fila').filter({ hasText: nombre }).locator('.stepper span')
	).toHaveText(String(antes + 1));
	await expect(page.getByText(`Ajuste manual · ${nombre}`)).toBeVisible();

	await page.getByRole('button', { name: 'Contar stock' }).click();
	const conteo = page.getByRole('dialog', { name: 'Contar stock' });
	await conteo.getByLabel(`Contado de ${nombre}`).fill('25');
	await conteo.getByRole('button', { name: 'Guardar conteo' }).click();
	await expect(page.getByText('Conteo guardado')).toBeVisible();
	await expect(page.getByText(`Conteo · ${nombre}`)).toBeVisible();
	await expect(
		page.locator('.fila').filter({ hasText: nombre }).locator('.stepper span')
	).toHaveText('25');
});

test('inventario: pedido a proveedores e imprimir códigos', async ({ page }) => {
	await demo(page, '/panel/inventario');
	await page.getByRole('button', { name: 'Pedir mercadería' }).click();
	const pedido = page.getByRole('dialog', { name: 'Pedido a proveedores' });
	await expect(pedido.getByText(/armé el pedido/)).toBeVisible();
	const r = await page.request.get('/panel/inventario/codigos');
	expect(await r.text()).toContain('<svg');
});

test('facturación: emitir una factura A pide CUIT', async ({ page }) => {
	await demo(page, '/panel/facturacion');
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	await page.getByRole('button', { name: 'Emitir factura' }).click();
	const m = page.getByRole('dialog');
	await m.getByRole('button', { name: 'Factura A' }).click();
	await m.getByLabel('Razón social').fill('Estudio Paz SRL');
	await m.getByLabel('Total').fill('121000');
	await m.getByRole('button', { name: 'Emitir' }).click();
	await expect(m.getByText(/necesita el CUIT/)).toBeVisible();
	await m.getByLabel('CUIT').fill('30-71234567-1');
	await m.getByRole('button', { name: 'Emitir' }).click();
	await expect(page.getByText('Factura A emitida')).toBeVisible();
	await expect(page.locator('.fila').first()).toContainText('Estudio Paz SRL');
});

test('envíos: etiqueta de un pedido listo lo pasa a en camino', async ({ page }) => {
	// Uno que se envía: los de retiro en el local no aparecen en Envíos.
	await demo(page, '/panel/pedidos?estado=para-preparar&envio=andreani');
	const numero = (await page.locator('.fila .q-s').first().textContent())!.match(/#(\d+)/)![1];
	await page
		.locator('.fila')
		.first()
		.getByRole('button', { name: new RegExp(`Preparar #${numero}`) })
		.click();
	await expect(page.getByText(`#${numero} preparado`)).toBeVisible();
	await page.goto(`/panel/envios?q=%23${numero}`);
	await listo(page);
	await expect(page.locator('.fila')).toHaveCount(1);
	await page
		.locator('.fila')
		.getByRole('button', { name: /^Etiqueta/ })
		.click();
	await expect(page.getByText(/1 etiqueta lista/)).toBeVisible();
	await expect(page.locator('.fila').getByRole('button', { name: /^Seguir/ })).toBeVisible();
});
