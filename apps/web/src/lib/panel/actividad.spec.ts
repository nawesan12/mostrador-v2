import { describe, expect, it } from 'vitest';
import { consultaDe, cuando, filtroDe, nombreCategoria, type Evento } from './actividad.ts';
import { filasDemo, paginaDemo } from '../server/sandbox/actividad.ts';

const ev = (i: number, x: Partial<Evento> = {}): Evento => ({
	id: String(i),
	tipo: 'equipo',
	quien: 'Vos',
	rol: 'Dueña',
	categoria: 'producto',
	foto: null,
	texto: `cambió algo ${i}`,
	cuando: new Date(Date.UTC(2026, 9, 30, 15, 0) - i * 60_000).toISOString(),
	donde: 'Panel',
	deshacer: false,
	deshecho: false,
	...x
});

describe('actividad', () => {
	it('lee de la URL sólo filtros válidos', () => {
		const f = filtroDe(
			new URLSearchParams('persona=abc-1&tipo=venta&desde=2026-10-01&hasta=ayer&cursor=x y')
		);
		expect(f).toEqual({ persona: 'abc-1', tipo: 'venta', desde: '2026-10-01' });
		expect(filtroDe(new URLSearchParams('tipo=nada'))).toEqual({});
		expect(consultaDe({ tipo: 'pedido', desde: '2026-10-01' })).toBe(
			'tipo=pedido&desde=2026-10-01'
		);
	});

	it('dice cuándo como se lee', () => {
		const ahora = Date.UTC(2026, 9, 30, 15, 0);
		expect(cuando(new Date(ahora - 20_000).toISOString(), ahora)).toBe('recién');
		expect(cuando(new Date(ahora - 5 * 60_000).toISOString(), ahora)).toBe('hace 5 min');
		expect(cuando(new Date(ahora - 3 * 3_600_000).toISOString(), ahora)).toMatch(/^hoy \d\d:\d\d$/);
		expect(cuando(new Date(ahora - 24 * 3_600_000).toISOString(), ahora)).toMatch(/^ayer /);
		expect(nombreCategoria('venta')).toBe('Punto de venta');
		expect(nombreCategoria('nada')).toBe('Otro');
	});

	it('la demo filtra y pagina como la API', () => {
		const todos = [
			...Array.from({ length: 60 }, (_, i) => ev(i)),
			ev(100, { quien: 'Carla', rol: 'Local', categoria: 'venta' })
		];
		const p1 = paginaDemo(todos, {});
		expect(p1.items).toHaveLength(50);
		expect(p1.siguiente).toBe('50');
		const p2 = paginaDemo(todos, { cursor: p1.siguiente! });
		expect(p2.items).toHaveLength(11);
		expect(p2.siguiente).toBeNull();
		expect(paginaDemo(todos, { persona: 'equipo' }).items.map((e) => e.quien)).toEqual(['Carla']);
		expect(paginaDemo(todos, { tipo: 'venta' }).items).toHaveLength(1);
		expect(paginaDemo(todos, { desde: '2026-10-31' }).items).toHaveLength(0);
		const filas = filasDemo(todos, { tipo: 'venta' });
		expect(filas[0]).toEqual(['Fecha', 'Hora', 'Quién', 'Rol', 'Tipo', 'Qué hizo', 'Desde']);
		expect(filas[1].slice(2, 5)).toEqual(['Carla', 'Local', 'Punto de venta']);
	});
});
