import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });

test('configuración: negocio, equipo, seguridad y plan', async ({ page, isMobile }) => {
	await page.addInitScript(() => localStorage.setItem('mostrador-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	await page.goto('/panel/configuracion');
	await listo(page);
	const nav = page.getByRole('navigation', { name: 'Secciones' });

	await page.getByLabel('Horario').fill('Lun a vie · 9 a 18 h');
	await page.getByRole('button', { name: 'Guardar cambios' }).click();
	await expect(page.getByText('Datos guardados')).toBeVisible();
	await page.reload();
	await listo(page);
	await expect(page.getByLabel('Horario')).toHaveValue('Lun a vie · 9 a 18 h');

	await nav.getByRole('button', { name: 'Equipo y permisos' }).click();
	await page.getByRole('button', { name: 'Invitar' }).click();
	await page.getByLabel('Email o WhatsApp').fill('juli@mail.com');
	await page
		.getByRole('group', { name: 'Rol' })
		.getByRole('button', { name: /^Depósito/ })
		.click();
	await page.getByRole('button', { name: 'Mandar invitación' }).click();
	await expect(page.locator('.miembro').filter({ hasText: 'juli@mail.com' })).toContainText(
		'Depósito'
	);
	await expect(page.getByText('Equipo · 5 personas')).toBeVisible();

	await nav.getByRole('button', { name: 'Seguridad' }).click();
	await page
		.locator('.sesion')
		.filter({ hasText: 'App en iPhone' })
		.getByRole('button', { name: 'Cerrar' })
		.click();
	await expect(page.locator('.sesion')).toHaveCount(2);

	await nav.getByRole('button', { name: 'Plan' }).click();
	await page.getByRole('button', { name: 'Elegir Emprende' }).click();
	await expect(page.getByText('Plan Emprende elegido')).toBeVisible();
	await expect(page.locator('.plan.actual')).toContainText('Emprende');
	if (!isMobile) await expect(page.locator('.prueba')).toContainText('Emprende');
});
