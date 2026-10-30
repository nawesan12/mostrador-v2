import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });
const sinDesborde = (page: Page) =>
	page.evaluate(() => document.documentElement.scrollWidth - innerWidth);

/** Una demo de peluquería (plantilla Tijera): la tienda es Navaja. */
async function demo(page: Page) {
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: /^Tijera/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	return (await page.locator('.t-d').first().textContent())!.split('/t/')[1].trim();
}

const horarios = (page: Page) =>
	page.locator('[role=radiogroup][aria-label^="Horarios"] [role=radio]');

/** Elige Corte con Bruno y el primer horario libre a partir de mañana; devuelve el inicio (ISO). */
async function elegir(page: Page, slug: string) {
	await page.goto(`/t/${slug}/turnos`);
	await listo(page);
	await page.getByRole('radiogroup', { name: 'Servicio' }).getByRole('radio').first().click();
	await page
		.getByRole('radiogroup', { name: 'Profesional' })
		.getByRole('radio', { name: /Bruno/ })
		.click();
	const dias = page
		.getByRole('radiogroup', { name: 'Día' })
		.locator('[role=radio]:not([disabled])');
	for (let i = 1; i < (await dias.count()); i++) {
		// Hasta que llegan los horarios del día nuevo se ven los del anterior: esperar la respuesta.
		const llegan = page.waitForResponse((r) => r.url().includes('/turnos/libres?'));
		await dias.nth(i).click();
		await llegan;
		await expect(horarios(page).first()).toBeVisible();
		const libre = page.locator(
			'[role=radiogroup][aria-label^="Horarios"] [role=radio]:not([disabled])'
		);
		if (await libre.count()) {
			// Si el clic cayó mientras cargaba, no queda elegido: se reintenta.
			await expect(async () => {
				await libre.first().click();
				await expect(libre.first()).toHaveAttribute('aria-checked', 'true', { timeout: 1000 });
			}).toPass();
			break;
		}
	}
	return page.locator('input[name=inicio]').inputValue();
}

async function datos(page: Page, nombre: string) {
	await page.getByRole('textbox', { name: 'Nombre' }).fill(nombre);
	await page.getByRole('textbox', { name: 'WhatsApp' }).fill('11 5555-0101');
	await page.getByRole('textbox', { name: 'Email' }).fill('lucia@ejemplo.com');
	await page.getByRole('button', { name: 'Confirmar turno' }).click();
}

/** El mismo horario, de nuevo en la tienda: ¿se ofrece? */
async function ofrecido(page: Page, slug: string, inicio: string) {
	const q = new URLSearchParams({
		servicio: 'corte',
		profesional: 'bruno',
		dia: inicio.slice(0, 10)
	});
	await page.goto(`/t/${slug}/turnos?${q}`);
	await listo(page);
	const h = horarios(page).filter({ hasText: inicio.slice(11, 16) });
	await expect(h).toBeVisible();
	return h.isEnabled();
}

test('turnos: se reserva de punta a punta, entra a la agenda y el horario ya no se ofrece', async ({
	page
}) => {
	const slug = await demo(page);
	const inicio = await elegir(page, slug);
	expect(inicio).toMatch(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:00-03:00$/);
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	await datos(page, 'Lucía Pérez');
	await expect(page.getByRole('heading', { name: '¡Listo, Lucía!' })).toBeVisible();
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);

	// Agregar al calendario: el .ics del turno.
	const href = await page.getByRole('link', { name: 'Agregar al calendario' }).getAttribute('href');
	const ics = await page.request.get(href!);
	expect(ics.headers()['content-type']).toContain('text/calendar');
	const cuerpo = await ics.text();
	expect(cuerpo).toContain('BEGIN:VCALENDAR');
	expect(cuerpo).toContain('SUMMARY:Corte con Bruno');

	// En la agenda del panel; se marca atendido.
	await page.goto(`/panel/turnos?dia=${inicio.slice(0, 10)}&vista=dia`);
	await listo(page);
	await page
		.getByRole('button', { name: /Lucía Pérez/ })
		.first()
		.click();
	await page.getByRole('button', { name: 'Atendido' }).click();
	await expect(page.getByText('Marcado como atendido')).toBeVisible();
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);

	// El mismo horario ya no se ofrece.
	expect(await ofrecido(page, slug, inicio)).toBe(false);
});

test('turnos: cancelar con el link del turno libera el horario', async ({ page }) => {
	const slug = await demo(page);
	const inicio = await elegir(page, slug);
	await datos(page, 'Tomás Acosta');
	await expect(page.getByRole('heading', { name: '¡Listo, Tomás!' })).toBeVisible();
	expect(await ofrecido(page, slug, inicio)).toBe(false);
	await page.goBack();
	await listo(page);
	await page.getByRole('button', { name: 'Cancelar turno' }).click();
	await page.getByRole('button', { name: 'Sí, cancelar el turno' }).click();
	await expect(page.getByRole('heading', { name: 'Turno cancelado' })).toBeVisible();
	expect(await ofrecido(page, slug, inicio)).toBe(true);
});

test('turnos: la vitrina muestra la reserva pero no se reserva', async ({ page }) => {
	await page.goto('/t/navaja');
	await expect(page.getByText(/tienda de ejemplo/)).toBeVisible();
	await expect(page.getByRole('heading', { name: 'El equipo' })).toBeVisible();
	const inicio = await elegir(page, 'navaja');
	expect(inicio).not.toBe('');
	await expect(page.getByRole('button', { name: 'Es una tienda de ejemplo' })).toBeDisabled();
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
});
