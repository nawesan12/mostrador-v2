import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });

async function demo(page: Page) {
	await page.addInitScript(() => localStorage.setItem('mostrador-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	await page.goto('/panel/pedidos');
	await listo(page);
}

test('lista, filtra por la URL y vuelve a todos', async ({ page }) => {
	await demo(page);
	await expect(page.getByRole('heading', { name: 'Pedidos' })).toBeVisible();
	await expect(page.locator('.fila').first()).toContainText('#1048');
	expect(
		await page.evaluate(
			() => document.documentElement.scrollWidth - document.documentElement.clientWidth
		)
	).toBeLessThanOrEqual(0);

	await page
		.getByRole('group', { name: 'Etapas' })
		.getByRole('button', { name: /Para preparar/ })
		.click();
	await expect(page).toHaveURL(/estado=para-preparar/);
	await expect(page.locator('.fila .p-e, .fila .accion').first()).toBeVisible();
	await page.getByRole('button', { name: 'Sin pagar' }).first().click();
	await expect(page).toHaveURL(/estado=sin-pagar$/);
	await page.getByRole('textbox', { name: 'Buscar pedidos' }).fill('zzzz');
	await expect(page.getByText('Ningún pedido con esos filtros')).toBeVisible();
	await page.getByRole('button', { name: 'Limpiar filtros' }).click();
	await expect(page.locator('.fila').first()).toContainText('#1048');

	// Al recargar, los filtros siguen ahí.
	await page.goto('/panel/pedidos?canal=ml&orden=monto');
	await listo(page);
	await expect(page.getByText('Mercado Libre').first()).toBeVisible();
	await expect(page.getByRole('button', { name: /Orden: Mayor monto/ })).toBeVisible();
});

test('preparar un pedido, deshacerlo y verlo en la hoja', async ({ page }) => {
	await demo(page);
	await page.goto('/panel/pedidos?estado=para-preparar');
	await listo(page);
	const fila = page.locator('.fila').first();
	const numero = (await fila.locator('.q-s').textContent())!.match(/#(\d+)/)![1];
	await fila.getByRole('button', { name: new RegExp(`Preparar #${numero}`) }).click();
	await expect(page.getByText(`#${numero} preparado`)).toBeVisible();
	await page.getByRole('button', { name: 'Deshacer' }).click();
	await expect(page.locator('.fila').filter({ hasText: `#${numero}` })).toBeVisible();

	await page
		.locator('.fila')
		.filter({ hasText: `#${numero}` })
		.locator('.abrir')
		.click();
	const hoja = page.getByRole('dialog', { name: `Pedido #${numero}` });
	await expect(hoja.getByText('Historial')).toBeVisible();
	await hoja.getByRole('button', { name: 'Marcar como preparado' }).click();
	await expect(hoja.getByText('Listo para despachar').first()).toBeVisible();
});

test('seleccionar varios y prepararlos en lote', async ({ page }) => {
	await demo(page);
	await page.goto('/panel/pedidos?estado=para-preparar');
	await listo(page);
	const antes = Number(await page.locator('.etapa').nth(1).locator('.e-n').textContent());
	await page.locator('.fila .tilde').nth(0).click();
	await page.locator('.fila .tilde').nth(1).click();
	await expect(page.locator('.lote')).toContainText('2 seleccionados');
	await expect(page.locator('.lote')).toBeVisible();
	await page.locator('.lote').getByRole('button', { name: 'Preparar' }).click();
	await expect(page.getByText('Preparar · 2 pedidos')).toBeVisible();
	await expect(page.locator('.etapa').nth(1).locator('.e-n')).toHaveText(String(antes - 2));
});

test('el Resumen abre un pedido y exporta CSV', async ({ page }) => {
	await demo(page);
	await page.goto('/panel/pedidos?pedido=1048');
	await listo(page);
	await expect(page.getByRole('dialog', { name: 'Pedido #1048' })).toBeVisible();
	const r = await page.request.get('/panel/pedidos/exportar?canal=wa');
	expect(r.headers()['content-type']).toContain('text/csv');
	const csv = await r.text();
	expect(csv.split('\r\n')[0]).toContain('Número;Fecha;Cliente');
	expect(csv).toContain(';wa;');
});
