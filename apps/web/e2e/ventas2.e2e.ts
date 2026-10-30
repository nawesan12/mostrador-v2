import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });

async function demo(page: Page, ruta: string) {
	await page.addInitScript(() => localStorage.setItem('mostrador-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	await page.goto(ruta);
	await listo(page);
}

test('punto de venta: cobrar crea el pedido y descuenta stock', async ({ page }) => {
	await demo(page, '/panel/ventas-en-local');
	const prod = page.locator('.prod:not([disabled])').first();
	const nombre = (await prod.locator('.p-n').textContent())!.trim();
	const stock = Number((await prod.locator('.p-pie span').textContent())!.replace(/\D/g, ''));
	await prod.click();
	await prod.click();
	await expect(page.locator('.ticket .item')).toContainText(nombre);
	await page
		.getByRole('group', { name: 'Medio de pago' })
		.getByRole('button', { name: 'Efectivo' })
		.click();
	await page.getByRole('button', { name: /^Cobrar \$/ }).click();
	await expect(page.getByText(/^Cobrado \$/)).toBeVisible();
	await expect(page.locator('.ticket .item')).toHaveCount(0);
	await expect(page.locator('.prod').filter({ hasText: nombre }).locator('.p-pie span')).toHaveText(
		`${stock - 2} u`
	);
	await page.getByRole('button', { name: 'Ver pedido' }).click();
	// Con opciones, la grilla dice "Producto · Talle M" y el pedido lleva la variante aparte.
	await expect(page.getByRole('dialog', { name: /Pedido #\d+/ })).toContainText(
		nombre.split(' · ')[0]
	);
});

test('devoluciones: aprobar y deshacer', async ({ page }) => {
	await demo(page, '/panel/devoluciones');
	const dev = page.locator('.dev').first();
	await dev.getByRole('button', { name: 'Aprobar' }).click();
	await expect(page.getByText('Devolución aprobada')).toBeVisible();
	await expect(dev).toContainText('Aprobada');
	await page.getByRole('button', { name: 'Deshacer' }).click();
	await expect(dev.getByRole('button', { name: 'Aprobar' })).toBeVisible();
});

test('preguntas: editar y publicar la respuesta; ver las reseñas', async ({ page }) => {
	await demo(page, '/panel/resenas-y-preguntas');
	const q = page.locator('.preguntas .tarjeta').first();
	await q.getByRole('button', { name: 'Editar' }).click();
	await q.getByLabel('Respuesta').fill('¡Hola! Sí, enviamos a todo el país.');
	await q.getByRole('button', { name: 'Publicar' }).click();
	await expect(page.getByText('Respuesta publicada')).toBeVisible();
	await expect(q).toContainText('Publicada en Mercado Libre');
	await page.getByRole('button', { name: 'Reseñas' }).click();
	await expect(page.getByText(/reseñas · \d+% recomendaría/)).toBeVisible();
});
