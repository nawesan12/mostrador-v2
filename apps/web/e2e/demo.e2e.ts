import { expect, test } from '@playwright/test';

test('sin demo, el panel manda a elegir rubro', async ({ page }) => {
	await page.goto('/panel');
	await expect(page).toHaveURL(/\/demo$/);
	await expect(page.getByRole('heading', { level: 1 })).toHaveText('Probá Mostrador');
	await expect(page.getByRole('button', { name: /^Mostaza/ })).toBeVisible();
	const desborde = await page.evaluate(
		() => document.documentElement.scrollWidth - document.documentElement.clientWidth
	);
	expect(desborde).toBeLessThanOrEqual(0);
});

test('elegir un rubro arma la tienda y la recuerda', async ({ page, context }) => {
	await page.goto('/demo');
	await page.locator('html[data-listo]').waitFor({ state: 'attached' });
	await page.getByRole('button', { name: /^Mostaza/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	await expect(page.getByText('Hola, Café Bruma')).toBeVisible();

	const cookie = (await context.cookies()).find((c) => c.name === 'mostrador_demo');
	expect(cookie?.httpOnly).toBe(true);
	expect(cookie?.sameSite).toBe('Lax');

	// Volver a /demo marca el rubro actual; elegir otro cambia la tienda.
	await page.goto('/demo');
	await page.locator('html[data-listo]').waitFor({ state: 'attached' });
	await expect(page.getByRole('button', { name: /^Mostaza/ })).toContainText('Elegida');
	// Filtrar por rubro y buscar.
	await page.getByRole('group', { name: 'Rubro' }).getByRole('button', { name: 'Moda' }).click();
	await expect(page.getByRole('button', { name: /^Mostaza/ })).toHaveCount(0);
	await page.getByRole('button', { name: /^Pampa/ }).click();
	await expect(page.getByText('Hola, PMPA Store')).toBeVisible();
	await page.goto('/demo');
	await page.locator('html[data-listo]').waitFor({ state: 'attached' });
	await page.getByRole('textbox', { name: 'Buscar estilo o rubro' }).fill('zzz');
	await expect(page.getByText('No hay plantillas con «zzz».')).toBeVisible();
});

test('una cookie inventada no entra al panel', async ({ page, context, baseURL }) => {
	await context.addCookies([{ name: 'mostrador_demo', value: 'inventada', url: baseURL! }]);
	await page.goto('/panel');
	await expect(page).toHaveURL(/\/demo$/);
});

test('el cron interno exige el secreto', async ({ request }) => {
	const res = await request.post('/api/internal/cron', { data: {} });
	expect(res.status()).toBe(401);
});
