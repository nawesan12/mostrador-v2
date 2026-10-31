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

test('marketing: el costo sigue a la audiencia, Tincho reescribe y la campaña sale', async ({
	page
}) => {
	await demo(page, '/panel/marketing');
	const vip = page.getByRole('group', { name: '¿A quién?' }).getByRole('button', { name: /^VIP/ });
	const n = Number((await vip.locator('span').textContent())!);
	await vip.click();
	await expect(page.getByRole('button', { name: `Enviar a ${n} personas` })).toBeVisible();
	await expect(page.locator('.nu-d').first()).toContainText(`${n} mensajes de marketing`);
	const antes = await page.getByLabel('Mensaje').inputValue();
	await page.getByRole('button', { name: '✦ Reescribir con Tincho' }).click();
	await expect(page.getByLabel('Mensaje')).not.toHaveValue(antes);
	await expect(page.locator('.b-txt')).toContainText('Lucía');
	await page.getByRole('button', { name: `Enviar a ${n} personas` }).click();
	await expect(page.getByText('Campaña en camino')).toBeVisible();
	await expect(page.locator('.campana').first()).toContainText('Enviando');
	await page.getByRole('button', { name: 'Programar' }).click();
	await expect(page.getByText('Campaña programada').first()).toBeVisible();
	await page.getByRole('link', { name: '+ Nuevo cupón' }).click();
	await expect(page.getByText('Para qué es')).toBeVisible();
});

test('canales: el reparto de ventas y sin interruptores que no hacían nada', async ({ page }) => {
	await demo(page, '/panel/canales');
	await expect(page.locator('.leyenda')).toContainText('Tienda online');
	// Pausar un canal no ocultaba nada en ningún lado: ya no hay interruptor.
	await expect(page.locator('.canal').getByRole('switch')).toHaveCount(0);
});

test('informes: cambia la serie y exporta el CSV', async ({ page }) => {
	await demo(page, '/panel/informes');
	await expect(page.locator('.barras .barra')).toHaveCount(12);
	await page
		.getByRole('group', { name: 'Informe' })
		.getByRole('button', { name: 'Clientes' })
		.click();
	await expect(page.locator('.g-cab')).toContainText('Clientes nuevos');
	const descarga = page.waitForEvent('download');
	await page.getByRole('link', { name: 'Exportar' }).click();
	const d = await descarga;
	expect(d.suggestedFilename()).toBe('informe-clientes-nuevos-12-meses.csv');
});
