import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });
const sinDesborde = (page: Page) =>
	page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);

async function demo(page: Page, ruta: string) {
	await page.addInitScript(() => localStorage.setItem('mostrador-cookies', 'nec'));
	await page.goto('/demo');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	await page.goto(ruta);
	await listo(page);
}

test('clientes: segmentos en la URL, ficha y cupón', async ({ page }) => {
	await demo(page, '/panel/clientes');
	await expect(page.getByRole('heading', { name: 'Clientes' })).toBeVisible();
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	await page.getByRole('group', { name: 'Segmentos' }).getByRole('button', { name: /VIP/ }).click();
	await expect(page).toHaveURL(/seg=vip/);
	await expect(page.locator('.chip-f, .fila').first()).toBeVisible();

	await page.locator('.fila .abrir').first().click();
	const ficha = page.getByRole('dialog');
	await expect(ficha.getByText('Compras de los últimos 12 meses')).toBeVisible();
	await ficha.getByRole('button', { name: /Mandar cupón/ }).click();
	await expect(page.getByText(/Cupón del 10% para/)).toBeVisible();
	await expect(ficha.getByRole('button', { name: 'Cupón enviado' })).toBeDisabled();
});

test('clientes: alta y etiqueta', async ({ page }) => {
	await demo(page, '/panel/clientes');
	await page.getByRole('button', { name: 'Nuevo cliente' }).click();
	const alta = page.getByRole('dialog');
	await alta.getByLabel('Nombre y apellido').fill('Rosa Pérez Prueba');
	await alta.getByLabel('WhatsApp').fill('11 5555-1234');
	await alta.getByRole('button', { name: 'Guardar cliente' }).click();
	await expect(page.getByText('Rosa Pérez Prueba ya es cliente')).toBeVisible();
	await page.getByRole('textbox', { name: 'Buscar clientes' }).fill('Rosa Pérez Prueba');
	await expect(page.locator('.fila')).toHaveCount(1);
	await page.locator('.fila .tilde').click();
	await page.locator('.lote').getByRole('button', { name: 'Etiquetar' }).click();
	await page.getByRole('dialog').getByLabel('Etiqueta').fill('Feria');
	await page.getByRole('dialog').getByRole('button', { name: 'Etiquetar' }).click();
	await expect(page.getByText('Etiqueta «Feria»')).toBeVisible();
	await page.getByRole('textbox', { name: 'Buscar clientes' }).fill('Feria');
	await expect(page.locator('.fila')).toHaveCount(1);
});

test('productos: pausar, editar, crear y exportar', async ({ page }) => {
	await demo(page, '/panel/productos');
	await expect(page.getByText('Cargá productos con una foto')).toBeVisible();
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);

	const primero = page.locator('.fila').first();
	const nombre = (await primero.locator('.q-t').textContent())!.trim();
	await primero.getByRole('switch').click();
	await expect(page.getByText(`${nombre} pausado`)).toBeVisible();
	await page
		.getByRole('group', { name: 'Segmentos' })
		.getByRole('button', { name: /Pausados/ })
		.click();
	await expect(page.locator('.fila').filter({ hasText: nombre })).toBeVisible();

	await page.locator('.fila .abrir').first().click();
	const editor = page.getByRole('dialog', { name: 'Editar producto' });
	await editor.getByLabel('Nombre').fill(`${nombre} edición especial`);
	await editor.getByRole('button', { name: '✦ Mejorar con Hornero' }).click();
	await expect(editor.locator('#desc-prod')).toHaveValue(/envío a todo el país/);
	await editor.getByRole('button', { name: 'Guardar y publicar' }).click();
	await expect(page.getByText('Producto guardado')).toBeVisible();

	await page.getByRole('button', { name: 'Nuevo producto' }).click();
	const nuevo = page.getByRole('dialog', { name: 'Nuevo producto' });
	await nuevo.getByLabel('Nombre').fill('Termo de prueba');
	await nuevo.getByLabel('Precio base').fill('45000');
	await nuevo.getByLabel('Precio base').blur();
	await nuevo.getByRole('button', { name: 'Guardar y publicar' }).click();
	await expect(page.getByText('Producto creado')).toBeVisible();
	await page.goto('/panel/productos?q=Termo%20de%20prueba');
	await listo(page);
	await expect(page.locator('.fila')).toHaveCount(1);
	await expect(page.locator('.fila')).toContainText('$ 45.000');

	const csv = await (await page.request.get('/panel/listas/productos/exportar')).text();
	expect(csv.split('\r\n')[0]).toContain('Nombre;SKU;Categoría;Precio');
	expect(csv).toContain('Termo de prueba');
});

test('productos: importar un CSV', async ({ page }) => {
	await demo(page, '/panel/productos');
	await page.locator('input[type=file]').setInputFiles({
		name: 'productos.csv',
		mimeType: 'text/csv',
		buffer: Buffer.from(
			'nombre;precio;stock;categoria\nBombilla importada;1.700;12;Mates\nMatera importada;36500;3;Mates\n'
		)
	});
	await expect(page.getByText('2 productos importados')).toBeVisible();
	await page.goto('/panel/productos?q=importada');
	await listo(page);
	await expect(page.locator('.fila')).toHaveCount(2);
});

test('productos: importar desde Empretienda pide un link válido', async ({ page }) => {
	await demo(page, '/panel/productos');
	await page.getByRole('button', { name: 'Desde Empretienda' }).click();
	const importar = page.getByRole('button', { name: 'Importar', exact: true });
	await expect(importar).toBeDisabled();
	await page.getByLabel('Link de tu tienda').fill('no existe nada');
	await importar.click();
	await expect(page.getByRole('alert')).toContainText('mitienda.empretienda.com.ar');
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
});
