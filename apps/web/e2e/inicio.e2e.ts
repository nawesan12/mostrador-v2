import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });
const sinDesborde = (page: Page) =>
	page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);

test.beforeEach(async ({ page }) => {
	await page.addInitScript(() => localStorage.setItem('mostrador-cookies', 'nec'));
});

test('la portada carga con los encabezados de seguridad y sin desborde', async ({ page }) => {
	const res = await page.goto('/');
	expect(res?.status()).toBe(200);
	expect(res?.headers()['x-content-type-options']).toBe('nosniff');
	await listo(page);
	await expect(page.locator('h1:visible')).toContainText('Más seguro y por menos.');
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
});

test('la portada tiene Open Graph con imagen absoluta y liviana', async ({ page, request }) => {
	await page.goto('/');
	const og = (p: string) => page.locator(`meta[property="og:${p}"]`).getAttribute('content');
	expect(await og('title')).toBe('Mostrador · Vendé en todos lados, pagá menos');
	expect(await og('url')).toBe('https://mostrador.iamnaw.dev/');
	expect(await og('image')).toBe('https://mostrador.iamnaw.dev/og/home.jpg');
	const res = await request.get('/og/home.jpg');
	expect(res.status()).toBe(200);
	expect(res.headers()['content-type']).toContain('image/jpeg');
	expect((await res.body()).byteLength).toBeLessThan(300_000);
	await expect(page.locator('meta[name="twitter:card"]')).toHaveAttribute(
		'content',
		'summary_large_image'
	);
});

test('el CTA lleva al registro', async ({ page }) => {
	await page.goto('/');
	await listo(page);
	await page
		.getByRole('link', { name: /Crear mi tienda|Empezar/ })
		.first()
		.click();
	await expect(page).toHaveURL(/\/ingresar\?m=registro$/);
});

test('el chat de Hornero avanza solo y los controles se pueden tocar', async ({ page }) => {
	await page.goto('/');
	await listo(page);
	const seccion = page.locator('section#hornero');
	await seccion.scrollIntoViewIfNeeded();
	await expect(seccion.getByText('Hola! ¿Tienen el termo de 1 L en negro?')).toBeVisible({
		timeout: 5000
	});
	const ctl = seccion.getByRole('switch', { name: 'Pedir mi aprobación para descuentos' });
	await expect(ctl).toHaveAttribute('aria-checked', 'true');
	await ctl.click();
	await expect(ctl).toHaveAttribute('aria-checked', 'false');
});

test('las tres etapas usan los precios de los planes', async ({ page }) => {
	await page.goto('/');
	for (const p of ['$ 19.900', '$ 49.900', '$ 149.900'])
		await expect(page.getByText(p).first()).toBeAttached();
});

test.describe('menú mobile', () => {
	test.skip(({ isMobile }) => !isMobile, 'sólo en mobile');

	test('abre, navega a una sección y cierra con Esc', async ({ page }) => {
		await page.goto('/');
		await listo(page);
		const boton = page.getByRole('button', { name: 'Menú', exact: true });
		await boton.click();
		await expect(boton).toHaveAttribute('aria-expanded', 'true');
		await expect(page.getByRole('link', { name: /Hornero IA/ })).toBeVisible();
		await page.keyboard.press('Escape');
		await expect(boton).toHaveAttribute('aria-expanded', 'false');
		await boton.click();
		await page
			.getByRole('link', { name: /Conocé a Hornero/ })
			.last()
			.click();
		await expect(boton).toHaveAttribute('aria-expanded', 'false');
		await expect(page.locator('section#hornero h2')).toBeInViewport();
	});
});

test('tablet (1024 px): sin desborde y con el hero apilado', async ({ page }) => {
	await page.setViewportSize({ width: 1024, height: 800 });
	await page.goto('/');
	await listo(page);
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	await expect(page.locator('section.mob h1')).toBeVisible();
});

test('con movimiento reducido no hay animaciones infinitas', async ({ browser }) => {
	const ctx = await browser.newContext({ reducedMotion: 'reduce' });
	const page = await ctx.newPage();
	await page.addInitScript(() => localStorage.setItem('mostrador-cookies', 'nec'));
	await page.goto('/');
	await listo(page);
	const infinitas = await page.evaluate(
		() =>
			document.getAnimations().filter((a) => a.effect?.getComputedTiming().iterations === Infinity)
				.length
	);
	expect(infinitas).toBe(0);
	await ctx.close();
});
