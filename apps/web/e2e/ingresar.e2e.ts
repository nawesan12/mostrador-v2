import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });
const sinDesborde = (page: Page) =>
	page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);

test.beforeEach(async ({ page }) => {
	// El banner de cookies tapa parte del formulario en mobile: en estos tests ya se eligió.
	await page.addInitScript(() => localStorage.setItem('mostrador-cookies', 'nec'));
});

test('login: errores en criollo, mostrar contraseña y en la demo lleva a /demo', async ({
	page
}) => {
	await page.goto('/ingresar');
	await listo(page);
	await expect(page.getByRole('heading', { level: 1 })).toHaveText('Hola de nuevo');
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);

	await page.getByRole('button', { name: 'Ingresar', exact: true }).click();
	await expect(page.getByText('Escribí tu email.')).toBeVisible();
	await expect(page.getByText('Escribí tu contraseña.')).toBeVisible();

	await page.getByRole('textbox', { name: 'Email' }).fill('lucia@laesquina.com.ar');
	const clave = page.getByLabel('Contraseña', { exact: true });
	await clave.fill('una-clave');
	await expect(clave).toHaveAttribute('type', 'password');
	await page.getByRole('button', { name: 'Mostrar contraseña' }).click();
	await expect(clave).toHaveAttribute('type', 'text');

	await page.getByRole('button', { name: 'Ingresar', exact: true }).click();
	await expect(page).toHaveURL(/\/demo$/);
});

test('registro: medidor de fuerza y términos obligatorios', async ({ page }) => {
	await page.goto('/ingresar?m=registro');
	await listo(page);
	await expect(page.getByRole('heading', { level: 1 })).toHaveText('Creá tu tienda gratis');
	const clave = page.getByLabel('Contraseña', { exact: true });
	await clave.fill('Abcdefghijkl');
	await expect(page.getByText('Bastante segura · sumá un número o un símbolo')).toBeVisible();
	await clave.fill('Abcdefghijk1');
	await expect(page.getByText('Muy segura')).toBeVisible();

	await page.getByRole('textbox', { name: 'Tu nombre' }).fill('Lucía');
	await page.getByRole('textbox', { name: 'Nombre de la tienda' }).fill('La Esquina');
	await page.getByRole('textbox', { name: 'Email' }).fill('lucia@laesquina.com.ar');
	// Escribir la contraseña después no puede borrar lo anterior (bug que tuvo esta pantalla).
	await clave.fill('Abcdefghijk12');
	await expect(page.getByRole('textbox', { name: 'Tu nombre' })).toHaveValue('Lucía');
	await page.getByRole('button', { name: 'Crear mi tienda' }).click();
	await expect(
		page.getByText('Para crear la tienda tenés que aceptar los términos.')
	).toBeVisible();
	// Lo escrito no se pierde.
	await expect(page.getByRole('textbox', { name: 'Nombre de la tienda' })).toHaveValue(
		'La Esquina'
	);
});

test('recuperar por email muestra "Revisá tu email" sin poner el email en la URL', async ({
	page
}) => {
	await page.goto('/ingresar');
	await listo(page);
	await page.getByRole('link', { name: 'Olvidé mi contraseña' }).click();
	await expect(page.getByRole('heading', { level: 1 })).toHaveText('¿Te olvidaste la contraseña?');
	await page.getByRole('textbox', { name: 'Email de tu cuenta' }).fill('lucia@laesquina.com.ar');
	await page.getByRole('button', { name: 'Mandame el link' }).click();
	await expect(page.getByRole('heading', { level: 1 })).toHaveText('Revisá tu email');
	await expect(page.getByText('lucia@laesquina.com.ar')).toBeVisible();
	expect(page.url()).not.toContain('lucia');
	await expect(page.getByRole('button', { name: /Reenviar en 0:4\d/ })).toBeDisabled();
});

test('recuperar por WhatsApp valida el número', async ({ page }) => {
	await page.goto('/ingresar?m=recuperar');
	await listo(page);
	await page
		.getByRole('group', { name: 'Canal' })
		.getByRole('button', { name: 'WhatsApp' })
		.click();
	await page.getByRole('textbox', { name: 'WhatsApp de tu cuenta' }).fill('123');
	await page.getByRole('button', { name: 'Mandame el link' }).click();
	await expect(page.getByText(/Escribí el número con código de área/)).toBeVisible();
});

test.describe('banner de cookies', () => {
	test.beforeEach(async ({ page }) => {
		// Una sola vez: al recargar tiene que quedar la elección guardada.
		await page.addInitScript(() => {
			if (sessionStorage.getItem('e2e-cookies-limpias')) return;
			sessionStorage.setItem('e2e-cookies-limpias', '1');
			localStorage.removeItem('mostrador-cookies');
		});
	});

	test('configurar y guardar recuerda la elección', async ({ page }) => {
		await page.goto('/ingresar');
		await listo(page);
		const banner = page.getByRole('dialog', { name: 'Cookies' });
		await expect(banner).toBeVisible();
		await banner.getByRole('button', { name: 'Configurar' }).click();
		await banner.getByRole('switch', { name: 'Marketing' }).click();
		await banner.getByRole('button', { name: 'Guardar mi elección' }).click();
		await expect(banner).toHaveCount(0);
		expect(await page.evaluate(() => localStorage.getItem('mostrador-cookies'))).toBe(
			'{"an":true,"mk":true}'
		);
		await page.reload();
		await listo(page);
		await expect(page.getByRole('dialog', { name: 'Cookies' })).toHaveCount(0);
	});
});
