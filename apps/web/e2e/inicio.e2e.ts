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
	await expect(page).toHaveURL(/\/onboarding$/);
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

	test('con el menú abierto no aparece la burbuja de Hornero', async ({ page }) => {
		await page.goto('/');
		await listo(page);
		const avatar = page.getByRole('button', { name: 'Hornero', exact: true });
		await expect(avatar).toBeVisible({ timeout: 6000 });
		await page.getByRole('button', { name: 'Menú', exact: true }).click();
		await expect(avatar).toBeHidden();
		await page.keyboard.press('Escape');
		await expect(avatar).toBeVisible();
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

test('el modal de plantillas recorre, elige y lleva al registro con la plantilla', async ({
	page
}) => {
	await page.goto('/');
	await listo(page);
	await page.getByRole('button', { name: /Ver las 16 plantillas/ }).click();
	const modal = page.getByRole('dialog');
	await expect(modal).toBeVisible();
	await expect(modal.getByText('Plantilla Ceibo')).toBeAttached();
	await modal.getByRole('button', { name: 'Plantilla siguiente' }).click();
	await expect(modal.getByText('Plantilla Ceibo')).not.toBeAttached();
	await page.keyboard.press('Escape');
	await expect(modal).toBeHidden();
	await page.getByRole('button', { name: /Ver las 16 plantillas/ }).click();
	await page
		.getByRole('dialog')
		.getByRole('link', { name: /Usar esta plantilla/ })
		.click();
	await expect(page).toHaveURL(/\/onboarding$/);
	expect(await page.evaluate(() => localStorage.getItem('ob-tpl'))).toBeTruthy();
});

test('integraciones filtra por categoría', async ({ page }) => {
	await page.goto('/');
	await listo(page);
	const sec = page.locator('section#integraciones');
	const tarjetas = sec.locator('.integ');
	await expect(tarjetas).toHaveCount(54);
	await sec.getByRole('button', { name: /^Envíos/ }).click();
	await expect(tarjetas).toHaveCount(8);
	await expect(sec.locator('.n', { hasText: 'Andreani' })).toBeVisible();
});

test('precios: la calculadora y el ciclo anual cambian los montos', async ({ page }) => {
	await page.goto('/');
	await listo(page);
	const sec = page.locator('section#precios');
	await sec.getByLabel('Vendo por mes').fill('1000000');
	await expect(sec.getByText('Plan Emprende · 0% de comisión de Mostrador')).toBeVisible();
	await sec.getByLabel('Vendo por mes').fill('20000000');
	await expect(sec.getByText('Plan Negocio · 0% de comisión de Mostrador')).toBeVisible();
	await sec.getByRole('button', { name: /Anual/ }).click();
	await expect(sec.getByText('$ 37.425').first()).toBeVisible();
	await expect(sec.getByText('$ 449.100 por año')).toBeVisible();
});

test('el conector de IA confirma los cambios', async ({ page }) => {
	await page.goto('/');
	await listo(page);
	const sec = page.locator('section#ia');
	await sec.getByRole('button', { name: /ChatGPT/ }).click();
	await expect(sec.getByRole('button', { name: /ChatGPT/ })).toHaveAttribute(
		'aria-pressed',
		'true'
	);
	await sec.getByRole('button', { name: 'Confirmar' }).click();
	await expect(sec.getByRole('button', { name: 'Hecho' })).toBeVisible();
});

test('el widget de Hornero aparece, se minimiza y vuelve', async ({ page, isMobile }) => {
	await page.goto('/');
	await listo(page);
	if (isMobile) {
		// En el celular arranca minimizado para no tapar la pantalla.
		await expect(page.getByText('¿Te ayudo?')).toBeVisible({ timeout: 6000 });
		await page.getByRole('button', { name: 'Hornero', exact: true }).click();
	}
	await expect(page.getByText('¡Hola! Soy Hornero')).toBeVisible({ timeout: 6000 });
	await page.getByRole('button', { name: 'Minimizar' }).click();
	await expect(page.getByText('¿Te ayudo?')).toBeVisible();
	await page.getByRole('button', { name: 'Hornero', exact: true }).click();
	await expect(page.getByText('¡Hola! Soy Hornero')).toBeVisible();
});

test('el footer tiene el botón de arrepentimiento', async ({ page }) => {
	await page.goto('/');
	await expect(
		page.locator('footer').getByRole('link', { name: 'Botón de arrepentimiento' }).first()
	).toHaveAttribute('href', '/legales/arrepentimiento');
});

test('el widget espera a que se cierre el banner de cookies', async ({ browser }) => {
	const ctx = await browser.newContext();
	const page = await ctx.newPage();
	await page.goto('/');
	await listo(page);
	await expect(page.getByText('Usamos cookies')).toBeVisible();
	await page.waitForTimeout(3000);
	await expect(page.getByRole('button', { name: 'Hornero', exact: true })).toBeHidden();
	await page.getByRole('button', { name: 'Solo necesarias' }).click();
	await expect(page.getByRole('button', { name: 'Hornero', exact: true })).toBeVisible();
	await ctx.close();
});
