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

test('la portada tiene Open Graph con imagen absoluta y liviana', async ({ page, request }) => {
	await page.goto('/');
	const og = (p: string) => page.locator(`meta[property="og:${p}"]`).getAttribute('content');
	expect(await og('title')).toBe('Mostrador · Vendé en todos lados, pagá menos');
	expect(await og('url')).toBe('https://mostrador.iamnaw.dev/');
	const img = await og('image');
	expect(img).toBe('https://mostrador.iamnaw.dev/og/home.jpg');
	// La imagen se sirve desde este mismo deploy y pesa menos de 300 KB (WhatsApp).
	const res = await request.get('/og/home.jpg');
	expect(res.status()).toBe(200);
	expect(res.headers()['content-type']).toContain('image/jpeg');
	expect((await res.body()).byteLength).toBeLessThan(300_000);
	await expect(page.locator('meta[name="twitter:card"]')).toHaveAttribute(
		'content',
		'summary_large_image'
	);
});
