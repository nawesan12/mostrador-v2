import { describe, expect, it } from 'vitest';
import casos from '@mostrador/contrato/casos/combinar.json';
import { combinarDescuentos } from './cupon.svelte.ts';

describe('combo y cupón juntos (los mismos casos que Go)', () => {
	for (const c of casos.casos)
		it(c.nombre, () => {
			expect(combinarDescuentos(c.subtotal, c.combo, c.cupon, c.combinable)).toEqual(c.esperado);
		});
});
