import { expect, test } from '@playwright/test';

test('la portada carga, sin desborde horizontal y con los encabezados de seguridad', async ({
	page
}) => {
	const res = await page.goto('/');
	expect(res?.status()).toBe(200);
	expect(res?.headers()['x-content-type-options']).toBe('nosniff');
	await expect(page.getByRole('heading', { level: 1 })).toContainText('Todo lo que tienen');
	const desborde = await page.evaluate(
		() => document.documentElement.scrollWidth - document.documentElement.clientWidth
	);
	expect(desborde).toBeLessThanOrEqual(0);
});
