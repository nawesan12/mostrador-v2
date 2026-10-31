import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });

async function demo(page: Page, ruta: string) {
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	await page.goto('/prueba');
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

test('usuarios: invitar, rol fijo y quitar', async ({ page }) => {
	await demo(page, '/panel/usuarios-y-permisos');
	const martin = page.locator('.persona').filter({ hasText: 'Martín Ríos' });
	await martin.getByRole('group').getByRole('button', { name: 'Local' }).click();
	await expect(page.getByText('Martín ahora es Local')).toBeVisible();
	await expect(martin).toContainText('Punto de venta');
	// Roles fijos: la tarjeta explica qué puede cada uno (sin matriz para tocar).
	await expect(page.getByText('Qué puede cada rol')).toBeVisible();

	await page.getByRole('button', { name: 'Invitar persona' }).click();
	await page.getByLabel('Email').fill('juli@mail.com');
	await page
		.getByRole('group', { name: 'Rol', exact: true })
		.getByRole('button', { name: /^Administración/ })
		.click();
	await page.getByRole('button', { name: 'Mandar invitación' }).click();
	const juli = page.locator('.persona').filter({ hasText: 'juli@mail.com' });
	await expect(juli).toContainText('Invitación enviada');
	await juli.getByRole('button', { name: 'Reenviar' }).click();
	await expect(page.getByText('Invitación reenviada')).toBeVisible();
	await juli.getByRole('button', { name: 'Cancelar la invitación a juli@mail.com' }).click();
	await expect(page.locator('.persona').filter({ hasText: 'juli@mail.com' })).toHaveCount(0);

	await page.getByRole('button', { name: 'Quitar a Carla Díaz' }).click();
	await expect(page.getByText('Carla Díaz ya no tiene acceso')).toBeVisible();
	await expect(page.locator('.persona')).toHaveCount(3);
});
