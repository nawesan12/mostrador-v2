import { expect, test, type Page } from '@playwright/test';

/**
 * El onboarding con una cuenta real: 6 pasos (sin Canales ni Tincho), retoma donde quedó (también
 * desde otro navegador), el «atrás» vuelve un paso, y quien ya lo terminó sigue en el panel.
 */
const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });
const titulo = (page: Page) => page.locator('.paso h1, .paso h2').first();
const seguir = (page: Page) => page.locator('.pie .seguir').click();

test('se retoma donde quedó, el atrás vuelve un paso y terminado va al panel', async ({
	page,
	browser
}) => {
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	await page.goto('/ingresar?m=registro');
	await listo(page);
	await page.getByRole('textbox', { name: 'Tu nombre' }).fill('Ana');
	await page.getByRole('textbox', { name: 'Nombre de la tienda' }).fill('Kiosco Ana');
	await page.getByRole('textbox', { name: 'Email' }).fill(`ana+${Date.now()}@ejemplo.com`);
	await page.getByLabel('Contraseña', { exact: true }).fill('Abcdefghijk1');
	await page.locator('label.terminos .caja-check').click();
	await page.getByRole('button', { name: 'Crear mi tienda' }).click();
	await expect(page).toHaveURL(/\/onboarding\?paso=1$/);
	await listo(page);
	await expect(page.locator('.ar-paso')).toHaveText('Paso 2 de 6');
	// Sin objetivos de ejemplo: con la API no los usaba nada.
	await expect(page.getByText('¿Qué querés resolver primero?')).toHaveCount(0);
	await page.getByRole('button', { name: 'Almacén' }).click();
	await seguir(page);

	// Shopify sin link: avisa en vez de seguir.
	await page.getByRole('button', { name: 'Shopify' }).click();
	await seguir(page);
	await expect(page.getByText(/Pegá el link de tu tienda/)).toBeVisible();
	await page.getByRole('button', { name: 'Empiezo de cero' }).click();
	await seguir(page);
	await expect(titulo(page)).toHaveText(/Elegí cómo se ve/);

	// Recargar queda en el paso; el atrás del navegador vuelve uno.
	await page.reload();
	await listo(page);
	await expect(titulo(page)).toHaveText(/Elegí cómo se ve/);
	await page.goBack();
	await expect(titulo(page)).toHaveText(/otra plataforma/);
	await page.goForward();
	await expect(titulo(page)).toHaveText(/Elegí cómo se ve/);

	// Otro navegador con la misma sesión y sin nada guardado: retoma en el mismo paso.
	const otro = await browser.newContext();
	await otro.addCookies(await page.context().cookies());
	const p2 = await otro.newPage();
	await p2.goto('/onboarding');
	await listo(p2);
	await expect(titulo(p2)).toHaveText(/Elegí cómo se ve/);
	await otro.close();

	// Cobros y envíos, y «¡Listo!» con el link real de la tienda.
	await seguir(page);
	await expect(titulo(page)).toHaveText('Cómo te pagan y cómo entregás.');
	await seguir(page);
	await expect(page.getByRole('heading', { name: 'Tu tienda ya está en línea.' })).toBeVisible();
	const tienda = page.getByRole('link', { name: 'Ver mi tienda' });
	await expect(tienda).toHaveAttribute('href', /\/t\/kiosco-ana/);
	await expect(page.getByText('Sumá a tu equipo')).toHaveCount(0);

	// Ya terminado: volver al onboarding lleva al panel.
	await page.goto('/onboarding');
	await expect(page).toHaveURL(/\/panel$/);

	// «Cargar mi primer producto» (desde «¡Listo!», que sí se puede recargar) abre el editor.
	await page.goto('/onboarding?paso=5');
	await listo(page);
	await page.getByRole('button', { name: /Cargar mi primer producto/ }).click();
	await expect(page).toHaveURL(/\/panel\/productos/);
	await expect(page.getByRole('heading', { name: 'Nuevo producto' })).toBeVisible();
});
