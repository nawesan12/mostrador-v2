import { expect, test, type Page } from '@playwright/test';
import { readFileSync } from 'node:fs';

const sinDesborde = (page: Page) =>
	page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);

const PAGINAS = [
	['/marca', 'Marca'],
	['/marca/papeleria', 'Papelería corporativa'],
	['/marca/redes', 'Redes'],
	['/marca/imprimibles', 'Imprimibles'],
	['/marca/og', null]
] as const;

/** Medidas de un PNG, leídas del encabezado IHDR. */
function medidasPng(ruta: string) {
	const b = readFileSync(ruta);
	expect(b.subarray(1, 4).toString()).toBe('PNG');
	return [b.readUInt32BE(16), b.readUInt32BE(20)];
}

async function abrir(page: Page, ruta: string) {
	const errores: string[] = [];
	page.on('pageerror', (e) => errores.push(e.message));
	// Las fotos de Unsplash no hacen falta para probar: sin red externa el PNG usa un hueco.
	await page.route('https://images.unsplash.com/**', (r) => r.abort());
	await page.goto(ruta);
	await page.locator('html[data-listo]').waitFor({ state: 'attached' });
	return errores;
}

test.describe('/marca', () => {
	for (const [ruta, titulo] of PAGINAS) {
		test(`${ruta}: sin errores ni desborde, noindex`, async ({ page }) => {
			const errores = await abrir(page, ruta);
			if (titulo) await expect(page.getByRole('heading', { level: 1 })).toHaveText(titulo);
			await expect(page.locator('meta[name="robots"]')).toHaveAttribute('content', /noindex/);
			expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
			expect(errores).toEqual([]);
		});
	}

	test('los datos de la persona se ven al instante en las piezas', async ({ page }) => {
		await abrir(page, '/marca/papeleria');
		await page.getByLabel('Nombre y apellido').fill('Ana Pérez');
		await page.getByLabel('Cargo').fill('Ventas');
		const tarjeta = page.locator('[data-lamina="papeleria-1a-tarjeta-dorso"]');
		await expect(tarjeta).toContainText('Ana Pérez');
		await expect(tarjeta).toContainText('Ventas');
		await expect(tarjeta).toContainText('tulocalweb.com.ar');
		// Titular, CUIT y domicilio salen de legales.ts: los mismos que el pie del sitio.
		const factura = page.locator('[data-lamina="papeleria-5a-factura-b-modelo"]');
		await expect(factura).toContainText('Nahuel Santillán');
		await expect(factura).toContainText('CUIT 20-44859234-1');
		await expect(factura).toContainText('Mar del Plata, provincia de Buenos Aires');
		await expect(page.locator('body')).not.toContainText('Corrientes 1234');
		// La factura y el remito son modelos, sin CAE real.
		await expect(factura).toContainText('CAE 00000000000000');
	});

	test('baja una pieza en PNG a su medida real', async ({ page }, info) => {
		await abrir(page, '/marca/og');
		const [descarga] = await Promise.all([
			page.waitForEvent('download'),
			page.locator('[data-lamina="og-precios"]').getByRole('button', { name: 'PNG' }).click()
		]);
		expect(descarga.suggestedFilename()).toBe('og-precios.png');
		const ruta = info.outputPath('og-precios.png');
		await descarga.saveAs(ruta);
		expect(medidasPng(ruta)).toEqual([1200, 630]);
	});

	test('"Descargar todo" baja un .zip con los PNG de la familia', async ({ page }, info) => {
		await abrir(page, '/marca/og');
		const [descarga] = await Promise.all([
			page.waitForEvent('download', { timeout: 60_000 }),
			page.getByRole('button', { name: 'Descargar todo' }).click()
		]);
		expect(descarga.suggestedFilename()).toBe('tulocalweb-og.zip');
		const ruta = info.outputPath('og.zip');
		await descarga.saveAs(ruta);
		const zip = readFileSync(ruta);
		expect(zip.subarray(0, 2).toString()).toBe('PK');
		// Un encabezado local por archivo: las 9 imágenes OG.
		const firma = Buffer.from([0x50, 0x4b, 0x03, 0x04]);
		let entradas = 0;
		for (let i = zip.indexOf(firma); i >= 0; i = zip.indexOf(firma, i + 4)) entradas++;
		expect(entradas).toBe(9);
		expect(zip.toString('latin1')).toContain('og-home.png');
	});
});
