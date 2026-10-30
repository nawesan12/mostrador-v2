import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });
const sinDesborde = (page: Page) =>
	page.evaluate(() => document.documentElement.scrollWidth - innerWidth);

async function demo(page: Page) {
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	return (await page.locator('.t-d').first().textContent())!.split('/t/')[1].trim();
}

test('reseñas: quien compró deja su reseña, el panel la ve con compra y la oculta', async ({
	page
}) => {
	const slug = await demo(page);
	// Comprar (como tienda.e2e.ts), con Mercado Pago: en la demo queda cobrado.
	await page.goto(`/t/${slug}/catalogo`);
	await page.goto(`/t/${slug}/producto/bombilla-alpaca`);
	await listo(page);
	await page.getByRole('button', { name: 'Agregar al carrito' }).click();
	await expect(page).toHaveURL(/\/carrito$/);
	await page.getByRole('link', { name: 'Finalizar compra' }).click();
	await page.getByLabel('Email').fill('lucia@ejemplo.com');
	await page.getByLabel('Nombre y apellido').fill('Lucía Pérez');
	await page.getByLabel('Teléfono').fill('11 5555-0101');
	await page.getByLabel('Código postal').fill('1414');
	await page.getByLabel(/Dirección/).fill('Gorriti 4500');
	await page.getByRole('radio', { name: /Correo Argentino/ }).click();
	await page.getByRole('radio', { name: /Mercado Pago/ }).click();
	await page.getByRole('button', { name: /^Pagar/ }).click();
	await expect(page.getByRole('heading', { name: '¡Gracias, Lucía!' })).toBeVisible();

	// La reseña, desde la confirmación (con el código del pedido).
	await expect(page.getByRole('heading', { name: '¿Cómo te fue? Dejá tu reseña' })).toBeVisible();
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	await page.locator('label', { hasText: '5 estrellas' }).click();
	await expect(page.getByRole('radio', { name: '5 estrellas' })).toBeChecked();
	await expect(page.getByLabel('Tu nombre')).toHaveValue('Lucía');
	await page.getByLabel(/Contanos/).fill('Lo usé toda la semana: impecable, lo recomiendo.');
	await page.getByRole('button', { name: 'Publicar mi reseña' }).click();
	await expect(page.getByRole('heading', { name: '¡Gracias por tu reseña!' })).toBeVisible();

	// En el panel: primera, con compra verificada. Ocultarla (con Deshacer).
	await page.goto('/panel/resenas-y-preguntas');
	await listo(page);
	await page.getByRole('button', { name: 'Reseñas', exact: true }).click();
	const mia = page.locator('.resena', { hasText: 'Lo usé toda la semana' });
	await expect(mia).toContainText('Compra verificada');
	await mia.getByRole('button', { name: 'Ocultar' }).click();
	await expect(page.getByText('Reseña oculta')).toBeVisible();
	await expect(mia).toContainText('Oculta en la tienda');
	await expect(mia.getByRole('button', { name: 'Mostrar' })).toBeVisible();
	await page.goto(`/t/${slug}`);
	await expect(page.getByText('Lo usé toda la semana')).toHaveCount(0);
});

test('preguntas: se pregunta en la ficha, se responde en el panel y se ve en la ficha', async ({
	page
}) => {
	const slug = await demo(page);
	await page.goto(`/t/${slug}/catalogo`);
	await page.locator('a[href*="/producto/"]').first().click();
	await expect(page).toHaveURL(/\/producto\//);
	await listo(page);
	const ficha = page.url();
	await page.getByRole('button', { name: 'Preguntale a la tienda' }).click();
	await page.getByLabel('Tu pregunta').fill('¿Lo tienen en otro color?');
	await page.getByLabel('Tu nombre').fill('Martina');
	await page.getByLabel(/Email/).fill('martina@ejemplo.com');
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	await page.getByRole('button', { name: 'Mandar pregunta' }).click();
	await expect(page.getByRole('status')).toContainText('Te avisamos cuando responda.');

	await page.goto('/panel/resenas-y-preguntas');
	await listo(page);
	const q = page.locator('.preguntas .tarjeta', { hasText: '¿Lo tienen en otro color?' });
	await expect(q).toContainText('Tienda');
	await q.getByLabel(/Tu respuesta/).fill('¡Hola! Sí, también en negro y en natural.');
	await q.getByRole('button', { name: 'Publicar' }).click();
	await expect(q).toContainText('Publicada en tu tienda');

	await page.goto(ficha);
	await listo(page);
	await expect(page.getByText('¿Lo tienen en otro color?')).toBeVisible();
	await expect(page.getByText('¡Hola! Sí, también en negro y en natural.')).toBeVisible();
});

test('la vitrina de una plantilla no guarda preguntas', async ({ page }) => {
	await page.goto('/t/tienda-lumbre');
	await page.locator('a[href*="/producto/"]').first().click();
	await listo(page);
	await page.getByRole('button', { name: 'Preguntale a la tienda' }).click();
	await page.getByLabel('Tu pregunta').fill('¿Hacen envíos?');
	await page.getByLabel('Tu nombre').fill('Martina');
	await page.getByRole('button', { name: 'Mandar pregunta' }).click();
	await expect(page.getByText(/tienda de ejemplo/).last()).toBeVisible();
});
