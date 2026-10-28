import { describe, expect, it } from 'vitest';
import casos from '../../../../../api/internal/dominio/vidriera/testdata/rutas.json';
import { haciaValido, normalizarRuta } from './rutas.ts';

// La misma tabla que Go (vidriera/redireccion_test.go).
describe('paths de la tienda vieja', () => {
	it.each(casos.normalizar)('normaliza $entrada', ({ entrada, sale }) => {
		expect(normalizarRuta(entrada)).toBe(sale);
	});
	it.each(casos.hacia)('destino $entrada', ({ entrada, valido }) => {
		expect(haciaValido(entrada)).toBe(valido);
	});
});
