import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });

/** Sin scroll horizontal (a 375 px tampoco). */
async function sinDesborde(page: Page) {
	const ancho = await page.evaluate(() => document.documentElement.scrollWidth);
	expect(ancho).toBeLessThanOrEqual(page.viewportSize()!.width);
}

test('panel de partners de ejemplo: inicio, tiendas, link, comisiones, perfil y materiales', async ({
	page
}) => {
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	await page.goto('/partners');
	await listo(page);
	await page.getByRole('button', { name: 'Ver el panel de partners de ejemplo' }).click();
	await expect(page).toHaveURL(/\/partners\/panel$/);
	await listo(page);
	const nav = page.getByRole('navigation', { name: 'Secciones' });

	// Inicio: el mes, el próximo pago, el nivel y las tareas.
	await expect(page.getByRole('heading', { name: 'Hola, Ana' })).toBeVisible();
	await expect(page.getByText('Pampa Code · Partner Plata')).toBeVisible();
	await expect(page.getByText('Comisión del mes')).toBeVisible();
	await expect(page.getByRole('progressbar')).toBeVisible();
	await expect(page.getByText(/para Partner Oro/)).toBeVisible();
	await expect(
		page.getByText('Cargá tus datos de cobro para recibir las comisiones')
	).toBeVisible();
	await sinDesborde(page);

	// Mis tiendas: las 19 de ejemplo, filtros y alta de una tienda.
	await nav.getByRole('link', { name: 'Mis tiendas' }).click();
	await expect(page).toHaveURL(/\/partners\/panel\/tiendas$/);
	const lista = page.getByRole('list', { name: 'Tiendas referidas' });
	await expect(lista.getByRole('listitem')).toHaveCount(19);
	await page.getByRole('button', { name: /^En prueba/ }).click();
	await expect(lista.getByRole('listitem')).toHaveCount(3);
	await expect(page.getByText('Tiendas donde colaborás')).toBeVisible();
	await page.getByRole('button', { name: 'Sumar una tienda' }).click();
	await page.getByLabel('Email de la dueña').fill('lu@tienda.com');
	await expect(page.getByRole('link', { name: 'Armar el mail' })).toHaveAttribute(
		'href',
		/^mailto:lu%40tienda\.com\?.*empezar%3Fref%3Dpampa-code/
	);
	await page.getByRole('button', { name: 'Cancelar' }).click();
	await sinDesborde(page);

	// Link y cupón: link único, QR y números.
	await nav.getByRole('link', { name: 'Link y cupón' }).click();
	await expect(page.getByTestId('link-referido')).toHaveText(/\/empezar\?ref=pampa-code$/);
	await expect(page.getByTestId('cupon')).toHaveText('PAMPACODE20');
	await expect(page.getByAltText('QR de tu link de partner')).toBeVisible();
	await expect(page.getByText('Clics este mes')).toBeVisible();
	const bajada = page.waitForEvent('download');
	await page.getByRole('button', { name: 'SVG para imprenta' }).click();
	expect((await bajada).suggestedFilename()).toBe('qr-pampa-code.svg');
	await sinDesborde(page);

	// Comisiones: subir la factura del mes aprobado y cargar los datos de cobro.
	await nav.getByRole('link', { name: 'Comisiones y pagos' }).click();
	const pagos = page.getByRole('list', { name: 'Pagos' });
	await expect(pagos.getByRole('listitem')).toHaveCount(4);
	await pagos.getByRole('button', { name: /falta tu factura/ }).click();
	await page.getByLabel('Link de tu factura').fill('http://inseguro.com/f.pdf');
	await page.getByRole('button', { name: 'Cargar factura' }).click();
	await expect(page.getByRole('alert')).toContainText('link de la factura');
	await page.getByLabel('Link de tu factura').fill('https://drive.google.com/factura-de-prueba');
	await page.getByRole('button', { name: 'Cargar factura' }).click();
	await expect(page.getByText('Factura cargada', { exact: true })).toBeVisible();
	await expect(pagos.getByRole('button', { name: /factura cargada/ })).toBeVisible();

	await page.getByLabel('Alias').fill('pampa.code.mp');
	await page.getByLabel('Titular de la cuenta').fill('Pampa Code SRL');
	await page.getByLabel('CUIT').fill('20-12345678-0');
	await page.getByRole('button', { name: 'Guardar datos de cobro' }).click();
	await expect(page.getByRole('alert')).toContainText('CUIT');
	await page.getByLabel('CUIT').fill('20-12345678-6');
	await page.getByRole('button', { name: 'Guardar datos de cobro' }).click();
	await expect(page.getByText('Datos de cobro guardados', { exact: true })).toBeVisible();
	await expect(page.getByText('Alias · pampa.code.mp')).toBeVisible();
	await sinDesborde(page);

	// Perfil público: se edita y la vista previa acompaña; queda guardado.
	await nav.getByRole('link', { name: 'Perfil público' }).click();
	const previa = page.getByTestId('vista-previa');
	await expect(previa).toContainText('Pampa Code');
	await page.getByLabel('Ciudad').fill('Santa Fe');
	await page
		.getByRole('group', { name: 'Especialidades' })
		.getByRole('button', { name: 'Marketing' })
		.click();
	await expect(previa).toContainText('Santa Fe');
	await page.getByRole('button', { name: 'Guardar perfil' }).click();
	await expect(page.getByText('Perfil guardado', { exact: true })).toBeVisible();
	await page.reload();
	await listo(page);
	await expect(page.getByLabel('Ciudad')).toHaveValue('Santa Fe');
	await sinDesborde(page);

	// Materiales: las piezas de /marca con el ref puesto.
	await nav.getByRole('link', { name: 'Materiales' }).click();
	await expect(page.getByRole('link', { name: 'Abrir con mi link' }).first()).toHaveAttribute(
		'href',
		/\?ref=pampa-code$/
	);
	await expect(page.getByText('Textos para compartir')).toBeVisible();
	await sinDesborde(page);

	// Inicio ya no pide los datos de cobro.
	await nav.getByRole('link', { name: 'Inicio' }).click();
	await expect(page.getByText('Cargá tus datos de cobro para recibir las comisiones')).toHaveCount(
		0
	);
});

test('los QR de Imprimibles llevan el ref del partner', async ({ page }) => {
	await page.goto('/marca/imprimibles?ref=pampa-code');
	await listo(page);
	await expect(page.getByAltText('QR · Crear mi tienda').first()).toBeVisible();
});
