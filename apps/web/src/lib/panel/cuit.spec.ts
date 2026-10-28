import { describe, expect, it } from 'vitest';
import { cuitValido, formatoCuit, limpiarCuit } from './cuit.ts';

describe('cuitValido', () => {
	it('los mismos casos que el test de Go (arca.CUITValido)', () => {
		const casos: [string, boolean][] = [
			['20448592341', true],
			['20448592342', false],
			['30500010912', true],
			['2044859234', false],
			['2044859234a', false]
		];
		expect(casos.map(([c]) => cuitValido(c))).toEqual(casos.map(([, ok]) => ok));
	});

	it('con guiones se limpia antes; el formato los pone', () => {
		expect(cuitValido(limpiarCuit('20-44859234-1'))).toBe(true);
		expect(formatoCuit('20448592341')).toBe('20-44859234-1');
		expect(formatoCuit('2044')).toBe('20-44');
	});
});
