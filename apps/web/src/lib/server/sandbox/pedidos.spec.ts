import { describe, expect, it } from 'vitest';
import { demoDe } from './demo.ts';
import { detalleDemo, listarDemo, transicionValida, type CambioPedido } from './pedidos.ts';

const BASE = Date.parse('2026-10-20T15:00:00Z');
const AHORA = BASE + 30 * 60_000;
const d = demoDe('litoral', BASE)!;
const sin = new Map<number, CambioPedido>();

describe('pedidos de la demo', () => {
	it('lista todo, ordena por más nuevos y suma bien', () => {
		const l = listarDemo(d, sin, { cuantos: 10 }, AHORA);
		expect(l.total).toBe(d.pedidos.length);
		expect(l.filas[0].numero).toBe(1048);
		expect(l.filas).toHaveLength(10);
		expect(l.suma).toBe(d.pedidos.reduce((s, p) => s + p.total, 0));
	});

	it('cada faceta se cuenta sin su propio filtro', () => {
		const wa = listarDemo(d, sin, { canal: ['wa'] }, AHORA);
		const todos = listarDemo(d, sin, {}, AHORA);
		expect(wa.total).toBe(d.pedidos.filter((p) => p.canal === 'wa').length);
		expect(wa.facetas.canal).toEqual(todos.facetas.canal);
		expect(Object.values(wa.etapas).reduce((s, n) => s + n, 0)).toBe(wa.total);
		expect(Object.values(wa.facetas.pago).reduce((s, n) => s + n, 0)).toBe(wa.total);
	});

	it('busca por número, cliente, producto o ciudad (sin importar acentos)', () => {
		expect(listarDemo(d, sin, { q: '#1048' }, AHORA).filas[0].numero).toBe(1048);
		const p = d.pedidos[5];
		const nombre = d.clientes.find((c) => c.id === p.clienteId)!.nombre;
		const sinTilde = nombre.normalize('NFD').replace(/[̀-ͯ]/g, '');
		expect(
			listarDemo(d, sin, { q: sinTilde }, AHORA).filas.some((f) => f.numero === p.numero)
		).toBe(true);
	});

	it('ordena por monto y por urgencia', () => {
		const m = listarDemo(d, sin, { orden: 'monto', cuantos: 5 }, AHORA).filas.map((f) => f.total);
		expect(m).toEqual([...m].sort((a, b) => b - a));
		const u = listarDemo(d, sin, { orden: 'urgentes', cuantos: 1 }, AHORA).filas[0];
		expect(['Para preparar', 'Listo para despachar']).toContain(u.estado);
	});

	it('aplica los cambios del visitante y arma el historial', () => {
		const p = d.pedidos.find((x) => x.estado === 'Para preparar')!;
		const cambios = new Map<number, CambioPedido>([
			[p.numero, { estado: 'Listo para despachar', en: { 'Listo para despachar': AHORA } }]
		]);
		const det = detalleDemo(d, cambios, p.numero, AHORA)!;
		expect(det.estado).toBe('Listo para despachar');
		expect(det.historial.find((h) => h.texto === 'Preparado')?.hecho).toBe(true);
		expect(det.comprador.pedidos).toBeGreaterThan(0);
		expect(detalleDemo(d, sin, 1, AHORA)).toBeNull();
	});

	it('sólo deja avanzar una etapa, volver una o cancelar', () => {
		expect(transicionValida('Para preparar', 'Listo para despachar')).toBe(true);
		expect(transicionValida('Listo para despachar', 'Para preparar')).toBe(true);
		expect(transicionValida('Para preparar', 'Entregado')).toBe(false);
		expect(transicionValida('Entregado', 'Cancelado')).toBe(false);
	});
});
