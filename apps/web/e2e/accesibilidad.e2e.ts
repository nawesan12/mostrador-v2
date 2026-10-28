import AxeBuilder from '@axe-core/playwright';
import { expect, test, type Page } from '@playwright/test';

/**
 * WCAG 2.1 A y AA con axe en las páginas principales. El texto de marca ya usa los tonos que pasan
 * (terracota #B0441D, verde #237550, ámbar #8A5F0E); color-contrast todavía no corta la prueba por
 * los logos de terceros (exentos), las miniaturas decorativas de la demo y algunos chips.
 */
const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });

async function auditar(page: Page, ruta: string) {
	await page.goto(ruta);
	await listo(page);
	const r = await new AxeBuilder({ page })
		.withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
		.disableRules(['color-contrast'])
		.analyze();
	expect(
		r.violations.map((v) => `${v.id}: ${v.nodes.map((n) => n.target.join(' ')).join(', ')}`),
		ruta
	).toEqual([]);
}

test('accesibilidad: sitio público', async ({ page }) => {
	await page.addInitScript(() => localStorage.setItem('mostrador-cookies', 'nec'));
	for (const ruta of ['/', '/demo', '/ingresar', '/migrar', '/ayuda', '/nosotros', '/onboarding'])
		await auditar(page, ruta);
});

test('accesibilidad: panel', async ({ page }) => {
	await page.addInitScript(() => localStorage.setItem('mostrador-cookies', 'nec'));
	await page.goto('/demo');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	for (const ruta of ['/panel', '/panel/pedidos', '/panel/productos', '/panel/clientes'])
		await auditar(page, ruta);
});
