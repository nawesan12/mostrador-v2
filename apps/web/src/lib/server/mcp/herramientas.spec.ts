import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { afterEach, describe, expect, it } from 'vitest';
import { d1Local } from '../sandbox/d1-local.ts';
import { crear, DatosSandbox, dbSandbox } from '../sandbox/sandbox.ts';
import { herramientasPara, llamar, type Contexto } from './herramientas.ts';
import type { Permiso } from './oauth.ts';

const dir = join(import.meta.dirname, '../../../../migraciones-sandbox');
const migraciones = readdirSync(dir)
	.filter((f) => f.endsWith('.sql'))
	.sort()
	.map((f) => readFileSync(join(dir, f), 'utf8'));

let d1: ReturnType<typeof d1Local> | undefined;
afterEach(() => {
	d1?.cerrar();
	d1 = undefined;
});

async function contexto(permisos: Permiso[] = ['lectura', 'acciones']): Promise<Contexto> {
	d1 = d1Local(migraciones);
	const db = dbSandbox(d1);
	const sb = (await crear(db, 'moda'))!;
	return { datos: new DatosSandbox(db, sb), permisos, sitio: 'https://mostrador.test' };
}
const datosDe = (r: Awaited<ReturnType<typeof llamar>>) =>
	r.structuredContent as Record<string, never>;

describe('herramientas MCP', () => {
	it('sólo lectura: las acciones no se listan y se rechazan', async () => {
		const c = await contexto(['lectura']);
		const nombres = herramientasPara(c.permisos).map((h) => h.name);
		expect(nombres).toContain('listar_pedidos');
		expect(nombres).not.toContain('crear_cupon');
		const r = await llamar(c, 'crear_cupon', { codigo: 'HOLA10', tipo: 'porcentaje', valor: 10 });
		expect(r.isError).toBe(true);
	});

	it('lee pedidos y productos con montos en pesos', async () => {
		const c = await contexto();
		const pedidos = datosDe(await llamar(c, 'listar_pedidos', { limite: 3 }));
		expect((pedidos.pedidos as unknown[]).length).toBeLessThanOrEqual(3);
		const [p] = pedidos.pedidos as { numero: number; total: { texto: string } }[];
		expect(p.total.texto).toMatch(/^\$\u00a0\d/);
		const detalle = datosDe(await llamar(c, 'ver_pedido', { numero: p.numero }));
		expect(detalle.numero).toBe(p.numero);
		expect((await llamar(c, 'ver_pedido', { numero: 999999 })).isError).toBe(true);
		const informe = datosDe(await llamar(c, 'informe_de_ventas', { periodo: '12m' }));
		expect(informe.meses).toHaveLength(12);
	});

	it('ajusta stock y precio, y valida los argumentos', async () => {
		const c = await contexto();
		const [prod] = datosDe(await llamar(c, 'listar_productos', { limite: 1 }))
			.productos as unknown as { id: string; stock: number }[];
		const r = datosDe(
			await llamar(c, 'ajustar_producto', { id: prod.id, sumar_stock: 5, precio: 12345 })
		);
		expect(r.stock).toEqual({ antes: prod.stock, ahora: prod.stock + 5 });
		expect((await c.datos.producto(prod.id))?.precio).toBe(12345);
		expect((await llamar(c, 'ajustar_producto', { id: prod.id })).isError).toBe(true);
		expect((await llamar(c, 'ajustar_producto', { id: prod.id, stock: -1 })).isError).toBe(true);
	});

	it('respeta los interruptores del panel', async () => {
		const c = await contexto();
		// "Reembolsos y descuentos" arranca apagado: no se crean cupones.
		const no = await llamar(c, 'crear_cupon', { codigo: 'HOLA10', tipo: 'porcentaje', valor: 10 });
		expect(no.isError).toBe(true);
		expect(no.content[0].text).toContain('Reembolsos y descuentos');
		await c.datos.ajustarIntegracion({ tipo: 'permisoIa', id: 'plata', valor: true });
		const si = datosDe(
			await llamar(c, 'crear_cupon', { codigo: 'HOLA10', tipo: 'porcentaje', valor: 10 })
		);
		expect(si.codigo).toBe('HOLA10');
		// Un código largo se rechaza entero, no se recorta a 20.
		const largo = await llamar(c, 'crear_cupon', {
			codigo: 'VERANO2026-SUPEROFERTA',
			tipo: 'porcentaje',
			valor: 10
		});
		expect(largo.isError).toBe(true);
	});

	it('crea un link de pago con su dirección pública', async () => {
		const c = await contexto();
		const r = datosDe(await llamar(c, 'crear_link_de_pago', { monto: 15000, concepto: 'Seña' }));
		expect(r.url).toMatch(/^https:\/\/mostrador\.test\/p\//);
	});
});
