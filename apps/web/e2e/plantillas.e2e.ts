import { expect, test } from '@playwright/test';
import { PLANTILLAS } from '../src/lib/plantillas.ts';
const slugDe = (n: string) =>
	n
		.normalize('NFD')
		.replace(/[\u0300-\u036f]/g, '')
		.toLowerCase()
		.replace(/&/g, 'y')
		.replace(/[^a-z0-9]+/g, '-')
		.replace(/^-|-$/g, '');

/** Las plantillas con diseño propio: su vitrina abre sin errores ni desborde. */
const PROPIAS = [
	'andes',
	'quebracho',
	'miga',
	'miga-clasica',
	'tijera',
	'pampa',
	'rosedal',
	'malbec',
	'mostaza',
	'yerbal'
];
for (const id of PROPIAS) {
	const p = PLANTILLAS.find((x) => x.id === id)!;
	const slug = slugDe(p.tienda);
	test(`vitrina ${id}`, async ({ page }) => {
		const errores: string[] = [];
		page.on('pageerror', (e) => errores.push(e.message));
		for (const ruta of ['', '/catalogo', '/nosotros']) {
			const r = await page.goto(`/t/${slug}${ruta}`);
			expect(r?.status(), ruta).toBe(200);
			const desborde = await page.evaluate(() => document.documentElement.scrollWidth - innerWidth);
			expect(desborde, `${id}${ruta}`).toBeLessThanOrEqual(0);
		}
		expect(errores).toEqual([]);
	});
}

/**
 * Turnos en cada plantilla: en una demo de esa plantilla se prenden los turnos con un servicio y
 * alguien que atiende, y la reserva y su link en el nav se dibujan sin errores ni desborde.
 */
for (const id of [...PROPIAS.filter((x) => x !== 'tijera'), 'ceibo', 'patio']) {
	const p = PLANTILLAS.find((x) => x.id === id)!;
	test(`turnos en ${id}`, async ({ page }) => {
		const errores: string[] = [];
		page.on('pageerror', (e) => errores.push(e.message));
		await page.addInitScript(() => localStorage.setItem('mostrador-cookies', 'nec'));
		await page.goto('/demo');
		await page.locator('html[data-listo]').waitFor({ state: 'attached' });
		await page.getByRole('button', { name: new RegExp(`^${p.nombre},`) }).click();
		await expect(page).toHaveURL(/\/panel$/);
		const ajustes = await (await page.request.get('/panel/turnos/ajustes/__data.json')).ok();
		expect(ajustes).toBe(true);
		const s = await page.request.post('/panel/turnos/servicios', {
			data: {
				nombre: 'Asesoría',
				descripcion: '',
				precio: 9000,
				duracion: 30,
				foto: null,
				etiqueta: null,
				activo: true
			}
		});
		expect(s.status()).toBe(201);
		const pr = await page.request.post('/panel/turnos/profesionales', {
			data: { nombre: 'Marta', rol: 'Atiende', foto: null, servicios: [], activo: true }
		});
		expect(pr.status()).toBe(201);
		const a = await page.request.put('/panel/turnos/ajustes', {
			data: {
				activo: true,
				sena: 30,
				anticipacionMin: 60,
				diasVista: 14,
				horario: [1, 2, 3, 4, 5, 6].map((dia) => ({ dia, desde: '09:00', hasta: '20:00' }))
			}
		});
		expect(a.ok()).toBe(true);
		const slug = slugDe(p.tienda);
		await page.goto(`/t/${slug}`);
		await expect(page.locator(`a[href="/t/${slug}/turnos"]`).first()).toBeVisible();
		await page.goto(`/t/${slug}/turnos`);
		await page.locator('html[data-listo]').waitFor({ state: 'attached' });
		// Malbec pregunta la edad antes de mostrar la tienda (el cartel sale al hidratar: se espera,
		// no se mira si ya está, porque si aparece después tapa el clic).
		if (id === 'malbec') await page.getByRole('button', { name: 'Sí, soy mayor' }).click();
		await page.getByRole('radiogroup', { name: 'Servicio' }).getByRole('radio').first().click();
		await expect(page.locator('[role=radiogroup][aria-label^="Horarios"]').first()).toBeVisible();
		const desborde = await page.evaluate(() => document.documentElement.scrollWidth - innerWidth);
		expect(desborde, `${id}/turnos`).toBeLessThanOrEqual(0);
		expect(errores).toEqual([]);
	});
}
