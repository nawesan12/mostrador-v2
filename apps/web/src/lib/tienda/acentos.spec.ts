import { describe, expect, it } from 'vitest';
import { PLANTILLAS } from '#lib/plantillas.ts';
import { acentoElegido, acentos, CON_ACENTO, contraste } from './acentos.ts';

describe('acentos', () => {
	it('cada plantilla tiene cuatro opciones distintas, la primera es la suya', () => {
		for (const p of PLANTILLAS) {
			const a = acentos(p.id).map((x) => x.color);
			expect(a, p.id).toHaveLength(4);
			expect(new Set(a).size, p.id).toBe(4);
		}
		expect(acentos('ceibo')[0].color).toBe('#9C4A2F');
	});

	it('las alternativas tienen contraste AA con el texto que va encima', () => {
		for (const id of CON_ACENTO)
			for (const { color, sobre } of acentos(id).slice(1))
				expect(contraste(color, sobre), `${id} ${color}`).toBeGreaterThanOrEqual(4.5);
	});

	it('la opción 0 no pisa nada', () => {
		expect(acentoElegido('yerbal', 0)).toBeNull();
		expect(acentoElegido('yerbal', 1)).toBe('#B0441D');
	});
});
