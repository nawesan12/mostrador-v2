import { describe, expect, it } from 'vitest';
import { contraste } from '../acentos.ts';
import { estiloFinal, PRESETS } from './presets.ts';
import { paleta } from './tokens.ts';

describe('el borde de los campos se ve (WCAG 1.4.11, 3:1) en los 17 presets', () => {
	for (const id of Object.keys(PRESETS))
		it(id, () => {
			const p = paleta(estiloFinal(id, 0, {}));
			expect(contraste(p.bordeCampo, p.superficie)).toBeGreaterThanOrEqual(3);
			expect(contraste(p.bordeCampo, p.fondo)).toBeGreaterThanOrEqual(3);
		});
});
