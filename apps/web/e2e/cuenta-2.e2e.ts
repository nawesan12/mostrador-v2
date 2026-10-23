import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });

async function demo(page: Page, ruta: string) {
	await page.addInitScript(() => localStorage.setItem('mostrador-cookies', 'nec'));
	await page.goto('/demo');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	await page.goto(ruta);
	await listo(page);
}

test('dominio: conectar, queda verificando y se puede quitar', async ({ page }) => {
	await demo(page, '/panel/dominio');
	await page.getByLabel('Tu dominio').fill('https://MiTienda.com.ar/');
	await page.getByRole('button', { name: 'Conectar' }).click();
	const d = page.locator('.dominio').filter({ hasText: 'mitienda.com.ar' });
	await expect(d).toContainText('Verificando');
	await expect(page.locator('.chequeos')).toContainText('Registro CNAME de www');
	await page.getByRole('button', { name: 'Quitar mitienda.com.ar' }).click();
	await expect(d).toHaveCount(0);
});

test('notificaciones: un aviso por canal queda guardado', async ({ page }) => {
	await demo(page, '/panel/notificaciones');
	const sw = page.getByRole('switch', { name: 'Venta nueva por Email' });
	await expect(sw).toHaveAttribute('aria-checked', 'false');
	const guardado = page.waitForResponse(
		(r) => r.url().endsWith('/preferencias') && r.status() === 204
	);
	await sw.click();
	await guardado;
	await page.reload();
	await listo(page);
	await expect(page.getByRole('switch', { name: 'Venta nueva por Email' })).toHaveAttribute(
		'aria-checked',
		'true'
	);
});

test('mis datos: exportar en CSV y todo en ZIP; cerrar pide el nombre', async ({ page }) => {
	await demo(page, '/panel/mis-datos');
	const csv = page.waitForEvent('download');
	await page
		.locator('.export')
		.filter({ hasText: 'Pedidos' })
		.getByRole('button', { name: 'CSV' })
		.click();
	const a = await csv;
	expect(a.suggestedFilename()).toMatch(/^pedidos-\d{4}-\d{2}-\d{2}\.csv$/);
	const zipDl = page.waitForEvent('download');
	await page.getByRole('button', { name: 'Exportar todo' }).click();
	expect((await zipDl).suggestedFilename()).toMatch(/\.zip$/);
	await expect(page.locator('.recientes')).toContainText('Pedidos');

	await page.getByRole('button', { name: 'Cerrar cuenta' }).click();
	const dialogo = page.getByRole('dialog');
	await expect(dialogo.getByRole('button', { name: 'Cerrar cuenta' })).toBeDisabled();
	await dialogo.getByLabel('Nombre de la tienda').fill('mate & co');
	await dialogo.getByRole('button', { name: 'Cerrar cuenta' }).click();
	await expect(page).toHaveURL(/\/$/);
	await page.goto('/panel');
	await expect(page).toHaveURL(/\/demo$/);
});
