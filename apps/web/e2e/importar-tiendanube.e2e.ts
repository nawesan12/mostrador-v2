import { expect, test, type Page } from '@playwright/test';

/**
 * Importar desde Tiendanube en la demo, de punta a punta: el OAuth real del Worker contra un
 * Tiendanube falso (e2e/tiendanube-falso.ts) y la autorización interceptada en el navegador.
 */
const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });
const sinDesborde = (page: Page) =>
	page.evaluate(() => document.documentElement.scrollWidth - innerWidth);

test('importar desde Tiendanube: conectar, progreso de la cola y resumen', async ({ page }) => {
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);

	// La autorización de la app: el Worker manda a Tiendanube (303) y Tiendanube vuelve a la URL
	// registrada con el código y el state. El navegador no sigue esa redirección (no sale a la red):
	// se lee adónde mandó el Worker y se hace la vuelta como la haría Tiendanube.
	let autorizar = '';
	let state = '';
	await page.route('**/panel/importar/tiendanube', async (r) => {
		const ida = await r.fetch({ maxRedirects: 0 });
		autorizar = ida.headers()['location'] ?? '';
		state = new URL(autorizar).searchParams.get('state') ?? '';
		const vuelta = new URL('/panel/importar/tiendanube/vuelta', r.request().url());
		vuelta.search = new URLSearchParams({ code: 'codigo-e2e', state }).toString();
		await r.fulfill({ status: 302, headers: { location: vuelta.toString() } });
	});
	// Por las dudas, nada de tiendanube.com sale del navegador.
	await page.route(/^https:\/\/[^/]*tiendanube\.com\//, (r) => r.abort());

	await page.goto('/panel/productos');
	await listo(page);
	await page.getByRole('button', { name: 'Importar tienda' }).click();
	const modal = page.getByRole('dialog');
	const conectar = modal.getByRole('link', { name: 'Conectar Tiendanube' });
	await expect(conectar).toBeVisible();
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	await conectar.click();

	// De vuelta en Productos, el modal sigue la importación hasta el final.
	await expect(page).toHaveURL(/\/panel\/productos$/);
	expect(autorizar).toMatch(/^https:\/\/www\.tiendanube\.com\/apps\/321\/authorize\?state=/);
	expect(state.length).toBeGreaterThan(20);
	await expect(modal.getByRole('status')).toHaveText(/Tu tienda de Tiendanube ya está/, {
		timeout: 30_000
	});
	await expect(modal.getByText('Productos 3')).toBeVisible();
	await expect(modal.getByText('Clientes 2')).toBeVisible();
	await expect(modal.getByText('Pedidos 1')).toBeVisible();
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	await modal.getByRole('button', { name: 'Listo' }).click();
	await expect(page.getByText('Master Ball').first()).toBeVisible();

	// El mismo state no sirve dos veces: vuelve con el aviso.
	await page.goto(`/panel/importar/tiendanube/vuelta?code=codigo-e2e&state=${state}`);
	await expect(page.getByRole('dialog').getByRole('alert')).toHaveText(/venció/);
});
