import { readFile } from 'node:fs/promises';
import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });
const sinDesborde = (page: Page) =>
	page.evaluate(() => document.documentElement.scrollWidth - innerWidth);

async function demo(page: Page, ruta: string) {
	await page.addInitScript(() => localStorage.setItem('mostrador-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	await page.goto(ruta);
	await listo(page);
}

test('ARCA: datos fiscales, pedido de certificado (.csr) y pasos en ARCA', async ({ page }) => {
	await demo(page, '/panel/integraciones');
	const arca = page.locator('.tarjeta', { hasText: 'ARCA' });
	await arca.getByRole('button', { name: 'Configurar' }).click();
	const d = page.getByRole('dialog');

	// La demo arranca conectada en homologación: abre en probar y ajustar.
	await expect(d).toContainText('Conectada en homologación');
	await expect(d).toContainText('Paso 4 de 4');
	await expect(d).toContainText('Certificado válido hasta');
	await d.getByRole('button', { name: 'Probar' }).click();
	await expect(d.getByRole('status')).toContainText('tienda de ejemplo');

	// Renovar: datos fiscales con el dígito verificador del CUIT.
	await d.getByRole('button', { name: 'Renovar el certificado' }).click();
	await expect(d).toContainText('Paso 1 de 4');
	await d.getByLabel('CUIT').fill('20448592342');
	await expect(d.getByText(/dígito verificador no da/)).toBeVisible();
	await d.getByLabel('CUIT').fill('20448592341');
	await expect(d.getByLabel('CUIT')).toHaveValue('20-44859234-1');
	await expect(d.getByText(/dígito verificador no da/)).toHaveCount(0);
	await d.getByLabel('Razón social').fill('Nahuel Santillán');
	await d.getByRole('button', { name: 'Monotributo' }).click();
	await expect(d).toContainText('factura C');
	await d.getByLabel('Punto de venta').fill('4');
	await expect(d).toContainText('ambiente de prueba');
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	await d.getByRole('button', { name: 'Generar pedido' }).click();

	// Pedido de certificado: se descarga y es un CSR de verdad.
	await expect(d).toContainText('Paso 2 de 4');
	const [descarga] = await Promise.all([
		page.waitForEvent('download'),
		d.getByRole('button', { name: /Descargar pedido/ }).click()
	]);
	expect(descarga.suggestedFilename()).toBe('tulocalweb-20448592341.csr');
	const csr = await readFile((await descarga.path())!, 'utf8');
	expect(csr.startsWith('-----BEGIN CERTIFICATE REQUEST-----')).toBe(true);
	expect(csr.trimEnd().endsWith('-----END CERTIFICATE REQUEST-----')).toBe(true);
	await expect(d).toContainText('Autoservicio de Acceso a APIs de Homologación');
	await expect(d).toContainText('wsfe');

	// En producción, los pasos son los de clave fiscal.
	await d.getByRole('button', { name: 'Atrás' }).click();
	await d.getByRole('button', { name: 'Producción' }).click();
	await expect(d).toContainText('facturas reales');
	await d.getByRole('button', { name: 'Generar pedido' }).click();
	await expect(d).toContainText('Administración de Certificados Digitales');
	await expect(d).toContainText('Administrador de Relaciones de Clave Fiscal');
	await expect(d).toContainText('Facturación electrónica');

	// Certificado: lo que no es un certificado se avisa; uno en la demo, que es de ejemplo.
	await d.getByRole('button', { name: 'Ya tengo el certificado' }).click();
	await d.getByLabel('Certificado').fill('hola');
	await d.getByRole('button', { name: 'Guardar certificado' }).click();
	await expect(d.getByRole('alert')).toContainText('no es un certificado');
	await d
		.getByLabel('Certificado')
		.fill('-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----');
	await d.getByRole('button', { name: 'Guardar certificado' }).click();
	await expect(d.getByRole('alert')).toContainText('tienda de ejemplo');
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
});

test('ARCA: facturar al cobrar se apaga y se deshace', async ({ page }) => {
	await demo(page, '/panel/integraciones?arca');
	const d = page.getByRole('dialog');
	const alCobrar = d.getByRole('switch', { name: 'Facturar al cobrar' });
	await expect(alCobrar).toHaveAttribute('aria-checked', 'true');
	await alCobrar.click();
	await expect(page.getByText('Facturar al cobrar, apagado')).toBeVisible();
	await expect(alCobrar).toHaveAttribute('aria-checked', 'false');
	await page.getByRole('button', { name: 'Deshacer' }).click();
	await expect(alCobrar).toHaveAttribute('aria-checked', 'true');
});

test('facturación: anular una factura con nota de crédito', async ({ page }) => {
	await demo(page, '/panel/facturacion?seg=ok');
	const fila = page.locator('.fila').first();
	const pedido = (await fila.locator('.q-s').textContent())!.match(/pedido #(\d+)/)![1];
	await fila.locator('.abrir').click();
	const d = page.getByRole('dialog');
	await expect(d).toContainText('Autorizada por ARCA');
	await d.getByRole('button', { name: 'Anular con nota de crédito' }).click();
	await expect(d).toContainText('No se puede deshacer');
	await d.getByRole('button', { name: 'Anular', exact: true }).click();
	await expect(page.getByText(/^Factura [AB] anulada$/)).toBeVisible();
	await expect(page.getByText(/^Nota de crédito por/)).toBeVisible();

	// La factura queda anulada y la nota de crédito, autorizada.
	await page.goto(`/panel/facturacion?seg=void&q=${pedido}`);
	await listo(page);
	await expect(page.locator('.fila', { hasText: `pedido #${pedido}` }).first()).toBeVisible();
	await page.goto(`/panel/facturacion?seg=ok&q=${pedido}`);
	await listo(page);
	await expect(page.locator('.fila', { hasText: 'Nota de crédito' }).first()).toBeVisible();
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
});
