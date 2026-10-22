import { describe, expect, it } from 'vitest';
import { demoDe } from './demo.ts';
import {
	DEF_CLIENTES,
	DEF_PRODUCTOS,
	listarGenerico,
	registrosClientes,
	registrosProductos
} from './listas.ts';

const BASE = Date.parse('2026-10-20T15:00:00Z');
const AHORA = BASE + 30 * 60_000;
const d = demoDe('litoral', BASE)!;

describe('motor de listas de la demo', () => {
	const clientes = registrosClientes(d, BASE, AHORA, new Map());
	const productos = registrosProductos(d, AHORA, new Map(), new Map());

	it('los clientes salen de los pedidos y caen en algún segmento', () => {
		expect(clientes).toHaveLength(d.clientes.length);
		const conPedidos = clientes.find((c) => c.id === d.pedidos[0].clienteId)!;
		expect(conPedidos.pedidos).toBeGreaterThanOrEqual(1);
		expect(conPedidos.ultima).toBeLessThan(1);
		expect(new Set(clientes.map((c) => c.seg)).size).toBeGreaterThanOrEqual(3);
		expect(clientes.every((c) => c.meses.length === 12)).toBe(true);
	});

	it('cada segmento y faceta se cuenta sin su propio filtro', () => {
		const todos = listarGenerico(clientes, DEF_CLIENTES, {});
		const vip = listarGenerico(clientes, DEF_CLIENTES, { seg: ['vip'] });
		expect(vip.segmentos).toEqual(todos.segmentos);
		expect(vip.total).toBe(todos.segmentos.vip);
		const wa = listarGenerico(clientes, DEF_CLIENTES, { facetas: { canal: ['wa'] } });
		expect(wa.facetas.canal).toEqual(todos.facetas.canal);
		expect(Object.values(wa.segmentos).reduce((s, n) => s + n, 0)).toBe(wa.total);
	});

	it('ordena, busca sin tildes y respeta el rango y las banderas', () => {
		const ltv = listarGenerico(clientes, DEF_CLIENTES, { cuantos: 5 }).filas as unknown as {
			gastado: number;
		}[];
		expect(ltv.map((c) => c.gastado)).toEqual([...ltv.map((c) => c.gastado)].sort((a, b) => b - a));
		const alguien = clientes.find((c) => /[áéíóú]/.test(c.nombre))!;
		const sinTilde = alguien.nombre.normalize('NFD').replace(/[̀-ͯ]/g, '');
		expect(listarGenerico(clientes, DEF_CLIENTES, { q: sinTilde }).total).toBeGreaterThan(0);
		const caros = listarGenerico(clientes, DEF_CLIENTES, { rango: 500_000, banderas: ['wa'] });
		expect(
			(caros.filas as unknown as { gastado: number; wa: boolean }[]).every(
				(c) => c.gastado >= 500_000 && c.wa
			)
		).toBe(true);
	});

	it('los productos traen ventas, stock y estado; los cambios del visitante mandan', () => {
		expect(productos).toHaveLength(d.catalogo.productos.length);
		const top = listarGenerico(productos, DEF_PRODUCTOS, {}).filas[0] as unknown as {
			vendidos30: number;
		};
		expect(top.vendidos30).toBe(Math.max(...productos.map((p) => p.vendidos30)));
		const sinStock = productos.filter((p) => p.stock === 0);
		expect(sinStock.every((p) => p.estado === 'out')).toBe(true);
		const id = d.catalogo.productos[0].id;
		const cambiado = registrosProductos(
			d,
			AHORA,
			new Map([[id, { precio: 99_990, publicado: false }]]),
			new Map([[id, 5]])
		);
		const p = cambiado.find((x) => x.id === id)!;
		expect(p.precio).toBe(99_990);
		expect(p.estado).toBe('paused');
		expect(p.stock).toBe((d.stock[id] ?? 0) + 5);
		const nuevo = registrosProductos(
			d,
			AHORA,
			new Map([
				['n1', { nuevo: true, nombre: 'Termo nuevo', precio: 1000, publicado: true, creado: 1 }]
			]),
			new Map()
		);
		expect(nuevo[0].nombre).toBe('Termo nuevo');
	});
});
