import { describe, expect, it } from 'vitest';
import { CATALOGOS } from '#lib/catalogos.ts';
import { PLANTILLAS } from '#lib/plantillas.ts';
import { demoDe } from './demo.ts';
import { resumenDemo } from './resumen.ts';

const BASE = Date.parse('2026-10-20T15:00:00Z');

describe('datos de la demo', () => {
	it('las 16 plantillas tienen catálogo con categorías y fotos', () => {
		expect(new Set(CATALOGOS.map((c) => c.plantilla))).toEqual(
			new Set(PLANTILLAS.map((p) => p.id))
		);
		for (const c of CATALOGOS) {
			expect(c.productos.length).toBeGreaterThanOrEqual(8);
			for (const p of c.productos) {
				expect(c.categorias).toContain(p.categoria);
				expect(p.foto).toMatch(/^\d{10,}-[0-9a-f]+$/);
				expect(p.precio).toBeGreaterThan(0);
			}
			expect(new Set(c.productos.map((p) => p.id)).size).toBe(c.productos.length);
		}
	});

	it('es siempre la misma para una plantilla y se fecha desde la base', () => {
		const a = demoDe('litoral', BASE)!;
		const b = demoDe('litoral', BASE + 1)!;
		expect(a.pedidos.map((p) => p.total)).toEqual(b.pedidos.map((p) => p.total));
		expect(b.pedidos[0].creado - a.pedidos[0].creado).toBe(1);
		expect(a.pedidos[0].numero).toBe(1048);
		expect(Math.max(...a.pedidos.map((p) => p.creado))).toBeLessThan(BASE);
		expect(Math.min(...a.pedidos.map((p) => p.creado))).toBeGreaterThan(BASE - 31 * 86_400_000);
	});

	it('los pedidos usan productos del catálogo y los totales cierran', () => {
		for (const c of CATALOGOS) {
			const d = demoDe(c.plantilla, BASE)!;
			const ids = new Set(c.productos.map((p) => p.id));
			for (const p of d.pedidos) {
				expect(p.items.length).toBeGreaterThan(0);
				for (const i of p.items) expect(ids.has(i.productoId)).toBe(true);
				expect(p.total).toBe(p.items.reduce((s, i) => s + i.precio * i.cantidad, 0));
				expect(d.clientes.some((x) => x.id === p.clienteId)).toBe(true);
			}
		}
	});

	it('el resumen cuenta los pedidos para preparar y respeta las leídas', () => {
		const d = demoDe('andes', BASE)!;
		const r = resumenDemo(d, BASE, 0, BASE);
		expect(r.insignias.pedidos).toBe(d.pedidos.filter((p) => p.estado === 'Para preparar').length);
		expect(r.notificaciones.filter((n) => !n.leida)).toHaveLength(6);
		expect(r.notificaciones[0].detalle).toContain('Hornero ya respondió');
		const leido = resumenDemo(d, BASE, BASE, BASE);
		expect(leido.notificaciones.every((n) => n.leida)).toBe(true);
	});
});
