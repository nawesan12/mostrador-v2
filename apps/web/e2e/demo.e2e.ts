import { expect, test } from '@playwright/test';

test('sin demo, el panel manda a elegir rubro', async ({ page }) => {
	await page.goto('/panel');
	await expect(page).toHaveURL(/\/demo$/);
	await expect(page.getByRole('heading', { level: 1 })).toHaveText('¿Qué vendés?');
	await expect(page.getByRole('button', { name: /Gastronomía/ })).toBeVisible();
	const desborde = await page.evaluate(
		() => document.documentElement.scrollWidth - document.documentElement.clientWidth
	);
	expect(desborde).toBeLessThanOrEqual(0);
});

test('elegir un rubro arma la tienda y la recuerda', async ({ page, context }) => {
	await page.goto('/demo');
	await page.locator('html[data-listo]').waitFor({ state: 'attached' });
	await page.getByRole('button', { name: /Gastronomía/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	await expect(page.getByText('Hola, Café Bruma')).toBeVisible();

	const cookie = (await context.cookies()).find((c) => c.name === 'mostrador_demo');
	expect(cookie?.httpOnly).toBe(true);
	expect(cookie?.sameSite).toBe('Lax');

	// Volver a /demo marca el rubro actual; elegir otro cambia la tienda.
	await page.goto('/demo');
	await page.locator('html[data-listo]').waitFor({ state: 'attached' });
	await expect(page.getByRole('button', { name: /Gastronomía/ })).toContainText('Tu demo');
	await page.getByRole('button', { name: /Tecnología/ }).click();
	await expect(page.getByText('Hola, Sonido Sur')).toBeVisible();
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
