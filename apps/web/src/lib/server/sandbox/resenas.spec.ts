import { describe, expect, it } from 'vitest';
import { demoDe } from './demo.ts';
import { devolucionesDemo, resenasDemo } from './resenas.ts';

const BASE = Date.parse('2026-10-20T15:00:00Z');
const d = demoDe('litoral', BASE)!;

describe('devoluciones y reseñas de la demo', () => {
	it('las devoluciones salen de pedidos entregados y respetan lo resuelto', () => {
		const { devoluciones, tasa } = devolucionesDemo(d, new Map());
		expect(devoluciones.length).toBeGreaterThan(0);
		expect(devoluciones.every((x) => x.estado === 'Pendiente')).toBe(true);
		expect(tasa).toBeLessThan(5);
		const id = devoluciones[0].id;
		expect(devolucionesDemo(d, new Map([[id, 'Aprobada']])).devoluciones[0].estado).toBe(
			'Aprobada'
		);
	});

	it('el resumen de reseñas cierra y las preguntas guardan la respuesta', () => {
		const r = resenasDemo(d, BASE, new Map([['q1', 'Sí, llega en 48 h.']]));
		expect(r.total).toBe(d.resenas.length);
		expect(r.porEstrellas.reduce((s, n) => s + n, 0)).toBe(r.total);
		expect(r.preguntas[0].respuesta).toBe('Sí, llega en 48 h.');
		expect(r.preguntas[1].respuesta).toBeNull();
		expect(r.preguntas[1].sugerida.length).toBeGreaterThan(10);
	});
});
