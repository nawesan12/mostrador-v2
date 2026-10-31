import { describe, expect, it } from 'vitest';
import CONTRATO from '@mostrador/contrato/direccion.json';
import { PLANTILLAS } from '#lib/plantillas.ts';
import { slugDe } from '#lib/slug.ts';
import {
	esNombreAutomatico,
	motivoDeFormato,
	normalizarSlug,
	RESERVADOS,
	slugDeNombre,
	slugFinal
} from './direccion.ts';

describe('la dirección de la tienda (decisiones §38)', () => {
	it('valida igual que Go: los casos del contrato', () => {
		for (const c of CONTRATO.casos)
			expect(motivoDeFormato(c.slug) ?? 'libre', c.slug).toBe(c.motivo);
	});

	it('las vitrinas de todas las plantillas están reservadas', () => {
		for (const p of PLANTILLAS) expect(RESERVADOS, p.tienda).toContain(slugDe(p.tienda));
	});

	it('lo que se escribe queda como slug', () => {
		expect(normalizarSlug('Café de Ana')).toBe('cafe-de-ana');
		expect(normalizarSlug('  Mate & Co!!')).toBe('mate-y-co-');
		expect(slugFinal('  Mate & Co!!')).toBe('mate-y-co');
		expect(normalizarSlug('--hola--mundo')).toBe('hola-mundo');
		expect(normalizarSlug('x'.repeat(80))).toHaveLength(60);
		expect(slugDeNombre('Almacén Don Pepe')).toBe('almacen-don-pepe');
	});

	it('reconoce el nombre que pone el registro con Google, Apple o Mercado Pago', () => {
		expect(esNombreAutomatico('La tienda de Nahuel')).toBe(true);
		expect(esNombreAutomatico('La tienda de José ')).toBe(true);
		expect(esNombreAutomatico('La tienda de la esquina')).toBe(false);
		expect(esNombreAutomatico('Almacén Don Pepe')).toBe(false);
	});
});
