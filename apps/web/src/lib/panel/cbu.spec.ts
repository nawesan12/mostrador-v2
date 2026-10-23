import { describe, expect, it } from 'vitest';
import { cbuValido } from './cbu.ts';

describe('cbuValido', () => {
	it('acepta CBU/CVU con los verificadores bien (los mismos casos que Go)', () => {
		expect(cbuValido('2850590940090418135201')).toBe(true);
		expect(cbuValido('0170099220000067797370')).toBe(true);
	});
	it('rechaza un dígito mal o un largo distinto', () => {
		expect(cbuValido('2850590940090418135209')).toBe(false);
		expect(cbuValido('285059094009041813520')).toBe(false);
		expect(cbuValido('28505909400904181352a1')).toBe(false);
	});
});
