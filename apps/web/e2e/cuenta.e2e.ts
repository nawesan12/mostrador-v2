import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });

async function demo(page: Page, ruta: string) {
	await page.addInitScript(() => localStorage.setItem('mostrador-cookies', 'nec'));
	await page.goto('/demo');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	await page.goto(ruta);
	await listo(page);
}

test('plan: anual, pack de Tincho y cancelación en 3 pasos con deshacer', async ({ page }) => {
	await demo(page, '/panel/plan-y-facturacion');
	await page.getByRole('group', { name: 'Ciclo' }).getByRole('button', { name: /Anual/ }).click();
	await expect(page.locator('.c-s')).toContainText('plan Pro anual');
	await expect(page.locator('.plan.cur')).toContainText('$ 37.425');
	await page.getByRole('button', { name: /\+2\.000 respuestas/ }).click();
	await expect(page.getByText('2.000 respuestas más para Tincho')).toBeVisible();
	await expect(page.locator('.uso')).toContainText('de 5.000');

	await page.getByRole('button', { name: 'Cancelar plan' }).click();
	await page.getByRole('radio', { name: 'Me voy a otra plataforma' }).click();
	await page.getByRole('button', { name: 'Continuar' }).click();
	await expect(page.getByText('Antes de irte, compará')).toBeVisible();
	await page.getByRole('button', { name: 'Cancelar igual' }).click();
	await page.getByRole('button', { name: 'Cancelar mi plan' }).click();
	await expect(page.getByText('Plan cancelado')).toBeVisible();
	await expect(page.getByText('Tu plan está cancelado')).toBeVisible();
	await page.getByRole('button', { name: 'Reactivar' }).click();
	await expect(page.getByText('¿Necesitás frenar?')).toBeVisible();
});

test('usuarios: rol, permisos y quitar', async ({ page }) => {
	await demo(page, '/panel/usuarios-y-permisos');
	const martin = page.locator('.persona').filter({ hasText: 'Martín Ríos' });
	await martin.getByRole('group').getByRole('button', { name: 'Depósito' }).click();
	await expect(page.getByText('Martín ahora es Depósito')).toBeVisible();
	await expect(martin).toContainText('Stock');

	const celda = page.getByRole('button', { name: 'Ventas: Ver finanzas' });
	await expect(celda).toHaveAttribute('aria-pressed', 'false');
	await celda.click();
	await expect(celda).toHaveAttribute('aria-pressed', 'true');
	await page.getByRole('button', { name: 'Admin: Plan, usuarios y datos' }).click();
	await expect(page.getByRole('button', { name: 'Admin: Plan, usuarios y datos' })).toHaveAttribute(
		'aria-pressed',
		'false'
	);

	await page.getByRole('button', { name: 'Quitar a Carla Díaz' }).click();
	await expect(page.getByText('Carla Díaz ya no tiene acceso')).toBeVisible();
	await expect(page.locator('.persona')).toHaveCount(3);
});
