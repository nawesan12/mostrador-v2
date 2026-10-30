import { describe, expect, it } from 'vitest';
import { linkWhatsapp, normalizarWhatsapp } from './whatsapp.ts';

describe('normalizarWhatsapp', () => {
	// Los mismos casos que whatsapp.Normalizar en la API.
	it.each([
		['11 5555-0000', '5491155550000'],
		['+54 9 11 5555 0000', '5491155550000'],
		['54 11 5555 0000', '5491155550000'],
		['0351 15 555 0000', '5493515550000'],
		['351 555-0000', '5493515550000'],
		['3415551111', '5493415551111'],
		['123', '']
	])('%s → %s', (entrada, sale) => expect(normalizarWhatsapp(entrada)).toBe(sale));

	it('arma el link sólo si el número sirve', () => {
		expect(linkWhatsapp('341 555-1111')).toBe('https://wa.me/5493415551111');
		expect(linkWhatsapp('12')).toBeNull();
	});
});
