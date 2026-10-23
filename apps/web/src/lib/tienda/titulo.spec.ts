import { describe, expect, it } from 'vitest';
import { partirTitulo } from './titulo.ts';

describe('partirTitulo', () => {
	it('no deja palabras cortas colgando', () => {
		expect(partirTitulo('Orgánico y a granel')).toEqual(['Orgánico', 'y a granel']);
		expect(partirTitulo('El mate de cada mañana')).toEqual(['El mate', 'de cada mañana']);
	});
	it('una palabra sola queda entera', () => {
		expect(partirTitulo('Lumbre')).toEqual(['Lumbre', '']);
		expect(partirTitulo('Muebles hechos a mano')).toEqual(['Muebles', 'hechos a mano']);
	});
});
