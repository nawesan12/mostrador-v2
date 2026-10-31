import { describe, expect, it } from 'vitest';
import { claves } from '#lib/tiempo-real/reglas.ts';
import { esRutaPanel, insignia } from './notificaciones.ts';

describe('la campanita (decisiones §45)', () => {
	it('sólo abre páginas del panel', () => {
		for (const r of [
			'/panel',
			'/panel/pedidos/12',
			'/panel/productos?editar=0b4c1f2e-9d1a-4b7e-8a5c-2f6e1d3c4b5a',
			'/panel/resenas-y-preguntas'
		])
			expect(esRutaPanel(r), r).toBe(true);
		for (const r of [
			'https://otro.com/panel',
			'//otro.com',
			'/panelx',
			'/t/mi-tienda',
			'/panel//otro.com',
			'/panel/../salir',
			'javascript:alert(1)'
		])
			expect(esRutaPanel(r), r).toBe(false);
	});

	it('la insignia cuenta hasta 9', () => {
		expect(insignia(0)).toBe('');
		expect(insignia(-1)).toBe('');
		expect(insignia(3)).toBe('3');
		expect(insignia(9)).toBe('9');
		expect(insignia(10)).toBe('9+');
		expect(insignia(Number.NaN)).toBe('');
	});

	it('el aviso de tiempo real recarga sólo la campanita', () => {
		expect(claves({ tipo: 'notificacion.nueva' })).toEqual(['tienda:notificaciones']);
		expect(claves({ tipo: 'notificacion.leida' })).toEqual(['tienda:notificaciones']);
	});
});
