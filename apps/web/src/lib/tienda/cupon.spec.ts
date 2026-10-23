import { describe, expect, it } from 'vitest';
import casos from '../../../../api/internal/dominio/vidriera/testdata/cupones.json';
import { aplicarCupon, type CuponTienda, type Historial, type LineaCupon } from './cupon.ts';

describe('aplicarCupon (mismos casos que Go)', () => {
	const ahora = new Date(casos.ahora);
	for (const c of casos.casos)
		it(c.nombre, () => {
			expect(
				aplicarCupon(
					c.cupon as CuponTienda,
					c.lineas as LineaCupon[],
					c.historial as Historial,
					ahora
				)
			).toEqual(c.esperado);
		});
});
