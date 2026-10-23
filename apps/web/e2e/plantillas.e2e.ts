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
