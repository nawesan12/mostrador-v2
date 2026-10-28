import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });
const sinDesborde = (page: Page) =>
	page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);

async function ir(page: Page, ruta: string) {
	await page.goto(ruta);
	await listo(page);
	expect(await sinDesborde(page), ruta).toBeLessThanOrEqual(0);
}

test('equipo: el backoffice de ejemplo cambia la navegación según el rol', async ({ page }) => {
	await page.addInitScript(() => localStorage.setItem('mostrador-cookies', 'nec'));
	await ir(page, '/staff');
	await expect(page.getByText(/backoffice de ejemplo/)).toBeVisible();
	const nav = page.getByRole('navigation', { name: 'Secciones del equipo' });
	await expect(nav.getByRole('link', { name: 'Auditoría' })).toBeVisible();

	// Soporte no ve facturación, partners ni auditoría.
	await page
		.getByRole('group', { name: 'Ver como' })
		.getByRole('button', { name: 'Soporte' })
		.click();
	await expect(nav.getByRole('link', { name: 'Facturación' })).toHaveCount(0);
	await expect(nav.getByRole('link', { name: 'Auditoría' })).toHaveCount(0);
	await expect(nav.getByRole('link', { name: 'Consultas' })).toBeVisible();
	const r = await page.request.get('/staff/facturacion');
	expect(r.status()).toBe(403);
	await page
		.getByRole('group', { name: 'Ver como' })
		.getByRole('button', { name: 'Dueño' })
		.click();
	await expect(nav.getByRole('link', { name: 'Facturación' })).toBeVisible();
});

test('equipo: tiendas, ficha y acciones (en la demo no se aplican)', async ({ page }) => {
	await page.addInitScript(() => localStorage.setItem('mostrador-cookies', 'nec'));
	await ir(page, '/staff/tiendas');
	await expect(page.getByRole('heading', { name: 'Tiendas' })).toBeVisible();
	await page.locator('.fila .abrir').first().click();
	await expect(page).toHaveURL(/\/staff\/tiendas\/.+/);
	await listo(page);
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	await expect(page.getByText('Plan y suscripción')).toBeVisible();

	await page.getByRole('button', { name: 'Regalar pack de IA' }).click();
	await page.getByRole('dialog').getByRole('button', { name: 'Regalar pack de IA' }).click();
	await expect(page.getByText('En la demo no se aplica')).toBeVisible();

	// Cerrar pide motivo y escribir la dirección de la tienda.
	await page.getByRole('button', { name: 'Cerrar la tienda' }).click();
	await expect(page.getByRole('dialog').getByText(/Cerrar es definitivo/)).toBeVisible();
	await page.keyboard.press('Escape');
});

test('equipo: consultas con plazo legal, partners, facturación y auditoría', async ({ page }) => {
	await page.addInitScript(() => localStorage.setItem('mostrador-cookies', 'nec'));
	await ir(page, '/staff/consultas');
	await expect(page.getByRole('heading', { name: 'Consultas' })).toBeVisible();
	await page.locator('a.fila').first().click();
	await listo(page);
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	await expect(page.getByText('Mensaje', { exact: true })).toBeVisible();

	await ir(page, '/staff/consultas?estado=nueva');
	await ir(page, '/staff/partners');
	await expect(page.getByText('Postulaciones nuevas')).toBeVisible();
	await ir(page, '/staff/facturacion');
	await expect(page.getByText('Suscripciones', { exact: true })).toBeVisible();
	await ir(page, '/staff/auditoria');
	await expect(page.locator('.fila').first()).toBeVisible();
});
