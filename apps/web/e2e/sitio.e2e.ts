import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });
const sinDesborde = (page: Page) =>
	page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);

test.beforeEach(async ({ page }) => {
	await page.addInitScript(() => localStorage.setItem('mostrador-cookies', 'nec'));
});

for (const [ruta, h1] of [
	['/seguridad', 'Tu negocio no se cae. Y nada se pierde.'],
	['/nosotros', 'Hacemos que vender sea más barato y más simple para el que empieza.']
]) {
	test(`${ruta} carga sin desborde`, async ({ page }) => {
		const res = await page.goto(ruta);
		expect(res?.status()).toBe(200);
		await expect(page.locator('h1')).toHaveText(h1);
		expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	});
}

test('las preguntas de seguridad se abren de a una', async ({ page }) => {
	await page.goto('/seguridad');
	await listo(page);
	const primera = page.getByRole('button', {
		name: '¿Qué pasa si se cae Tu Local Web en pleno Hot Sale?'
	});
	const segunda = page.getByRole('button', { name: '¿Pueden ver los datos de mis clientes?' });
	await expect(primera).toHaveAttribute('aria-expanded', 'true');
	await segunda.click();
	await expect(segunda).toHaveAttribute('aria-expanded', 'true');
	await expect(primera).toHaveAttribute('aria-expanded', 'false');
	await expect(page.getByText('Solo el personal de soporte que vos autorices')).toBeVisible();
});

test('una ruta inexistente muestra el 404 del sitio', async ({ page }) => {
	const res = await page.goto('/esto-no-existe');
	expect(res?.status()).toBe(404);
	await expect(page.locator('h1')).toHaveText('El horno está vacío.');
	await expect(page.locator('meta[name="robots"]')).toHaveAttribute('content', 'noindex');
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
});

test('migrar: la pestaña sale de ?desde= y el botón lleva al registro con el origen', async ({
	page
}) => {
	await page.goto('/migrar?desde=mercadolibre');
	await listo(page);
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	await expect(page.getByRole('button', { name: 'Desde Mercado Libre' })).toHaveAttribute(
		'aria-pressed',
		'true'
	);
	await expect(page.getByText('Traemos tus publicaciones de Mercado Libre')).toBeVisible();
	await page.getByRole('button', { name: 'Desde Tiendanube' }).click();
	await expect(page).toHaveURL(/desde=tiendanube/);
	await page.getByLabel('Link de tu tienda en Tiendanube').fill('lucia.mitiendanube.com');
	await page.getByRole('button', { name: 'Empezar' }).click();
	await expect(page).toHaveURL(/\/onboarding\?desde=tiendanube&origen=lucia\.mitiendanube\.com$/);
});
