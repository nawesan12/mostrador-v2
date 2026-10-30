import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });

async function demo(page: Page, ruta: string) {
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	await page.goto(ruta);
	await listo(page);
}

test('crear pedido sin pagar, armar el link y pagarlo: el pedido pasa a preparar', async ({
	page,
	context
}) => {
	await demo(page, '/panel/pedidos/nuevo');
	await page.getByLabel('Buscar cliente').fill('a');
	await page.getByLabel('Buscar cliente').fill('ar');
	await page.locator('.sugeridos button').first().click();
	const cliente = (await page.locator('.e-n').textContent())!.trim();
	await page.locator('.prod:not([disabled])').first().click();
	await page.locator('.prod:not([disabled])').nth(1).click();
	await expect(page.locator('.resumen .item')).toHaveCount(2);
	await page.getByRole('button', { name: 'Crear y armar el link' }).click();
	const modal = page.getByRole('dialog', { name: 'Crear link de pago' });
	await expect(modal).toBeVisible();
	await expect(modal.getByLabel('Concepto')).toHaveValue(/Pedido #\d+/);
	await expect(modal).toContainText(`¡Hola ${cliente.split(' ')[0]}!`);
	const numero = (await modal.getByLabel('Concepto').inputValue()).replace(/\D/g, '');
	await context.grantPermissions(['clipboard-read', 'clipboard-write']).catch(() => {});
	await modal.getByRole('button', { name: 'Crear y copiar link' }).click();
	const aviso = page.getByText(/\/p\/[2-9A-Z]{6} · \$/);
	await expect(aviso).toBeVisible();
	const codigo = (await aviso.textContent())!.match(/\/p\/([2-9A-Z]{6})/)![1];

	// Quien paga abre el link en otro navegador, sin la cookie de la demo.
	const otro = await page.context().browser()!.newContext();
	const p2 = await otro.newPage();
	await p2.goto(`/p/${codigo}`);
	await expect(p2.getByText(`Pedido #${numero}`)).toBeVisible();
	await p2.getByRole('button', { name: /^Pagar \$/ }).click();
	await expect(p2.getByText('¡Listo, pagaste!')).toBeVisible();
	await otro.close();

	await page.goto(`/panel/pedidos?pedido=${numero}`);
	await expect(page.getByRole('dialog', { name: `Pedido #${numero}` })).toContainText(
		'Para preparar'
	);
});

test('crear pedido ya pagado y desde ⌘K', async ({ page }) => {
	await demo(page, '/panel');
	await page.keyboard.press('ControlOrMeta+k');
	const buscar = page.getByRole('dialog').getByLabel('Buscar', { exact: true });
	await expect(buscar).toBeFocused();
	await buscar.fill('crear pedido');
	await page.keyboard.press('Enter');
	await expect(page).toHaveURL(/\/panel\/pedidos\/nuevo$/);
	await listo(page);
	await page.getByPlaceholder('Nombre', { exact: true }).fill('Rocío Prueba');
	await page.locator('.prod:not([disabled])').first().click();
	await page
		.getByRole('group', { name: 'Cómo se cobra' })
		.getByRole('button', { name: 'Ya está pagado' })
		.click();
	await page.getByRole('button', { name: 'Crear pedido', exact: true }).click();
	await expect(page.getByText(/^Pedido #\d+ creado$/)).toBeVisible();
	await expect(page.getByRole('dialog', { name: /Pedido #\d+/ })).toContainText('Rocío Prueba');
});

test('un link que no existe da 404', async ({ page }) => {
	const r = await page.goto('/p/ZZZZZZ');
	expect(r?.status()).toBe(404);
});
