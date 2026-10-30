import { describe, expect, it } from 'vitest';
import { volverValido } from './volver.ts';

describe('volverValido', () => {
	it('acepta el consentimiento de una IA y las páginas del panel', () => {
		for (const v of [
			'/oauth/autorizar?x=1',
			'/panel',
			'/panel/pedidos?estado=nuevo',
			'/onboarding?paso=3'
		])
			expect(volverValido(v)).toBe(true);
	});
	it('rechaza otros sitios y rutas que no son del panel', () => {
		for (const v of [
			'',
			null,
			'//evil.com/panel',
			'/panel//evil.com',
			'/\\evil.com',
			'https://evil.com',
			'/panelx',
			'/t/tienda'
		])
			expect(volverValido(v)).toBe(false);
	});
});
