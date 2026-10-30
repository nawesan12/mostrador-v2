import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });

test('mi tienda: ocultar, reordenar y cambiar colores queda sin publicar hasta publicar', async ({
	page
}) => {
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	await page.goto('/panel/mi-tienda');
	await listo(page);
	await expect(page.getByRole('button', { name: 'Todo publicado' })).toBeVisible();
	await expect(page.locator('.sitio .wa')).toBeVisible();
	// Los tres cambios se guardan uno detrás del otro: se esperan los tres antes de recargar.
	let guardados = 0;
	page.on('response', (r) => {
		if (
			r.url().endsWith('/panel/mi-tienda') &&
			r.request().method() === 'PUT' &&
			r.status() === 204
		)
			guardados++;
	});
	await page.getByRole('button', { name: 'Ocultar Botón de WhatsApp' }).click();
	await expect(page.locator('.sitio .wa')).toHaveCount(0);
	await page.getByRole('button', { name: 'Subir Reseñas' }).click();
	await page.getByRole('button', { name: 'Paleta 4' }).click();
	await expect(page.getByRole('button', { name: 'Publicar cambios' })).toBeVisible();
	await expect.poll(() => guardados, { timeout: 15_000 }).toBe(3);
	await page.reload();
	await listo(page);
	await expect(page.locator('.panelito .s-t').nth(2)).toContainText('Reseñas');
	await expect(page.getByRole('button', { name: 'Paleta 4' })).toHaveAttribute(
		'aria-pressed',
		'true'
	);
	await page.getByRole('button', { name: 'Publicar cambios' }).click();
	await expect(page.getByText('Cambios publicados')).toBeVisible();
	await expect(page.getByRole('button', { name: 'Todo publicado' })).toBeVisible();
});
