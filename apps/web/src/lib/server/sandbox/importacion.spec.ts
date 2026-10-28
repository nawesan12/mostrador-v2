import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { afterEach, describe, expect, it } from 'vitest';
import { d1Local } from './d1-local.ts';
import { crear, DatosSandbox, dbSandbox, purgar, VIDA_MS } from './sandbox.ts';
import { redireccionDe } from './importacion.ts';
import {
	ErrorDefinitivo,
	FUENTES,
	type Etapa,
	type Fuente,
	type Pagina
} from '../importacion/fuente.ts';
import type { RegCliente } from './listas.ts';
import type { Demo } from './demo.ts';

const dir = join(import.meta.dirname, '../../../../migraciones-sandbox');
const migraciones = readdirSync(dir)
	.filter((f) => f.endsWith('.sql'))
	.sort()
	.map((f) => readFileSync(join(dir, f), 'utf8'));

let d1: ReturnType<typeof d1Local> | undefined;
afterEach(() => {
	d1?.cerrar();
	d1 = undefined;
	delete FUENTES.tiendanube;
});

const DIA = 86_400_000;

/** Fuente en memoria: el cursor es el número de página. */
function falsa(paginas: Partial<Record<Etapa, Omit<Pagina, 'siguiente'>[]>>, fallar?: Error) {
	const f: Fuente & { pedidas: number } = {
		pedidas: 0,
		async pagina(etapa, cursor) {
			f.pedidas++;
			if (fallar) throw fallar;
			const n = typeof cursor === 'number' ? cursor : 0;
			const ps = paginas[etapa] ?? [];
			if (n >= ps.length) return { siguiente: null };
			return { ...ps[n], siguiente: n + 1 < ps.length ? n + 1 : null };
		}
	};
	return f;
}

async function nueva() {
	d1 = d1Local(migraciones);
	const db = dbSandbox(d1);
	const sb = (await crear(db, 'moda'))!;
	return { db, sb, datos: new DatosSandbox(db, sb) };
}

const FUENTE = () =>
	falsa({
		productos: [
			{
				productos: [
					{
						externo: 'p1',
						url: 'https://vieja.mitiendanube.com/productos/remera-lisa/',
						producto: {
							nombre: 'Remera lisa',
							precio: 15000,
							sku: 'RL-1',
							stock: 4,
							fotos: ['https://cdn.ejemplo.com/1.jpg']
						}
					},
					{ externo: 'p2', url: '/productos/x', producto: { nombre: 'x', precio: 1 } }
				]
			},
			{
				productos: [
					// Mismo SKU: repetido.
					{
						externo: 'p3',
						url: '/productos/otra',
						producto: { nombre: 'Otra', precio: 1, sku: 'RL-1' }
					}
				]
			}
		],
		clientes: [
			{
				clientes: [
					{
						externo: 'c1',
						nombre: 'Lucía Fernández',
						email: 'Lucia@Example.com',
						telefono: '341 555-0101'
					},
					{ externo: 'c2', nombre: 'Sin Compras', email: 'nadie@example.com' },
					{ externo: 'c3', nombre: 'Don Sin Email' },
					{ externo: 'c4', nombre: '', email: 'malo' }
				]
			}
		],
		pedidos: [
			{
				pedidos: [
					{
						externo: 'o1',
						numero: '#101',
						fecha: Date.now() - 40 * DIA,
						email: 'LUCIA@example.com',
						nombre: 'Lucía F.',
						total: 50000,
						estado: 'paid',
						items: [{ nombre: 'Remera lisa', variante: '', cantidad: 1, precio: 50000 }]
					},
					{
						externo: 'o1',
						numero: '#101',
						fecha: Date.now(),
						total: 1,
						estado: '',
						items: []
					},
					{
						externo: 'o2',
						numero: '#102',
						fecha: Date.now() - 10 * DIA,
						email: 'otra@example.com',
						nombre: 'Otra Persona',
						total: 20000,
						estado: 'shipped',
						items: []
					}
				]
			}
		],
		redirecciones: [
			{
				redirecciones: [
					{ desde: '/categorias/remeras/', hacia: '/catalogo' },
					{ desde: '/contacto', hacia: '//otro.com' }
				]
			}
		]
	});

describe('importación en la demo', () => {
	it('avanza de a un paso por pedido, sin duplicar, y une los clientes por email', async () => {
		const { db, sb, datos } = await nueva();
		expect(await datos.crearImportacion({ plataforma: 'shopify' })).toBe('sin-fuente');
		FUENTES.tiendanube = () => FUENTE();
		expect(await datos.crearImportacion({ plataforma: 'tiendanube', url: 'javascript:x' })).toBe(
			'invalido'
		);
		const antes = (await datos.listar('clientes', { cuantos: 1 })).total;
		const pedidosAntes = (await datos.listarPedidos({})).total;
		const imp = await datos.crearImportacion({ plataforma: 'tiendanube' });
		expect(typeof imp === 'object' && imp.estado).toBe('pendiente');
		expect(await datos.crearImportacion({ plataforma: 'tiendanube' })).toBe('activa');

		let pasos = 0;
		let a = await datos.importacionActual();
		while (a && (a.estado === 'pendiente' || a.estado === 'corriendo') && pasos < 20) {
			a = await datos.avanzarImportacion();
			pasos++;
		}
		// productos (2 páginas) + clientes + pedidos + redirecciones.
		expect(pasos).toBe(5);
		expect(a?.estado).toBe('lista');
		expect(a?.terminada).not.toBeNull();
		expect(a?.progreso).toEqual({
			productos: 1,
			clientes: 3,
			pedidos: 2,
			redirecciones: 2,
			repetidos: 2,
			recortados: 0
		});
		expect(a?.errores.map((e) => e.externo)).toEqual(['p2', 'c4', '/contacto']);
		// Terminada, avanzar no hace nada.
		expect((await datos.avanzarImportacion())?.progreso.productos).toBe(1);

		// El producto entró con la función del editor y su URL vieja redirige a la ficha nueva.
		const remera = (await datos.listar('productos', { q: 'Remera lisa', cuantos: 10 }))
			.filas[0] as unknown as { id: string; sku: string };
		expect(remera.sku).toBe('RL-1');
		expect(await redireccionDe(db, sb.id, '/productos/Remera-Lisa?ref=ig')).toBe(
			`/producto/${remera.id}`
		);
		expect(await redireccionDe(db, sb.id, '/categorias/remeras')).toBe('/catalogo');
		expect(await redireccionDe(db, sb.id, '/contacto')).toBeNull();

		// Clientes: ficha sin compras, sin email y del historial; los pedidos viejos no son Pedidos.
		// La ventana de la lista es de 200: se busca cada uno por nombre.
		const buscar = async (q: string) =>
			(await datos.listar('clientes', { q, cuantos: 50 })).filas as unknown as RegCliente[];
		expect((await datos.listar('clientes', { cuantos: 1 })).total).toBe(antes + 4);
		const lucia = (await buscar('Lucía Fernández')).find((r) => r.id === 'lucia@example.com')!;
		expect(lucia).toMatchObject({ pedidos: 1, gastado: 50000, telefono: '341 555-0101', wa: true });
		expect(lucia.ultima).toBeGreaterThan(39);
		expect((await buscar('Sin Compras'))[0]).toMatchObject({ id: 'nadie@example.com', pedidos: 0 });
		expect((await buscar('Otra Persona'))[0]).toMatchObject({
			id: 'otra@example.com',
			pedidos: 1,
			seg: 'new'
		});
		expect(await buscar('Don Sin Email')).toHaveLength(1);
		expect((await datos.listarPedidos({})).total).toBe(pedidosAntes);
		const h = await datos.historialCliente('lucia@example.com');
		expect(h).toHaveLength(1);
		expect(h[0]).toMatchObject({ plataforma: 'tiendanube', numero: '#101', total: 50000 });

		// La segunda vez, todo repetido.
		await datos.crearImportacion({ plataforma: 'tiendanube' });
		for (let i = 0; i < 6; i++) await datos.avanzarImportacion();
		const b = await datos.importacionActual();
		expect(b?.progreso).toMatchObject({ productos: 0, clientes: 0, pedidos: 0 });
	});

	it('el historial se une a un cliente de ejemplo con el mismo email', async () => {
		const { datos } = await nueva();
		// Un cliente de ejemplo con email y compras (los datos de ejemplo salen del código).
		const d = await (datos as unknown as { base(): Promise<Demo> }).base();
		const regs = (await datos.listar('clientes', { cuantos: 200 }))
			.filas as unknown as RegCliente[];
		const ejemplo = regs.find(
			(r) => r.pedidos > 0 && d.clientes.find((c) => c.id === r.id)?.email
		)!;
		const email = d.clientes.find((c) => c.id === ejemplo.id)!.email!;
		FUENTES.tiendanube = () =>
			falsa({
				pedidos: [
					{
						pedidos: [
							{
								externo: 'v1',
								numero: '9',
								fecha: Date.now() - DIA,
								email: email.toUpperCase(),
								total: 7000,
								estado: '',
								items: []
							}
						]
					}
				]
			});
		await datos.crearImportacion({ plataforma: 'tiendanube' });
		for (let i = 0; i < 5; i++) await datos.avanzarImportacion();
		const despues = (
			(await datos.listar('clientes', { q: ejemplo.nombre, cuantos: 50 }))
				.filas as unknown as RegCliente[]
		).find((r) => r.id === ejemplo.id)!;
		expect(despues.pedidos).toBe(ejemplo.pedidos + 1);
		expect(despues.gastado).toBe(ejemplo.gastado + 7000);
		expect(await datos.historialCliente(ejemplo.id)).toHaveLength(1);
	});

	it('cancelar frena el paso y un error de la fuente la deja fallida', async () => {
		const { datos } = await nueva();
		const f = FUENTE();
		FUENTES.tiendanube = () => f;
		const imp = await datos.crearImportacion({ plataforma: 'tiendanube' });
		if (typeof imp === 'string') throw new Error(imp);
		await datos.avanzarImportacion();
		const c = await datos.cancelarImportacion(imp.id);
		expect(typeof c === 'object' && c.estado).toBe('cancelada');
		expect(await datos.cancelarImportacion(imp.id)).toBe('terminada');
		expect(await datos.cancelarImportacion('otra')).toBe('no-existe');
		const pedidas = f.pedidas;
		await datos.avanzarImportacion();
		expect(f.pedidas).toBe(pedidas);

		FUENTES.tiendanube = () =>
			falsa({}, new ErrorDefinitivo('La clave de Tiendanube ya no sirve.'));
		await datos.crearImportacion({ plataforma: 'tiendanube' });
		const a = await datos.avanzarImportacion();
		expect(a?.estado).toBe('fallida');
		expect(a?.errores.at(-1)?.detalle).toBe('La clave de Tiendanube ya no sirve.');
	});

	it('la purga se lleva todo lo importado con el sandbox', async () => {
		const { db, sb, datos } = await nueva();
		FUENTES.tiendanube = () => FUENTE();
		await datos.crearImportacion({ plataforma: 'tiendanube' });
		for (let i = 0; i < 5; i++) await datos.avanzarImportacion();
		expect(await purgar(db, sb.venceEn + VIDA_MS)).toBe(1);
		for (const t of ['clientes', 'historial_pedidos', 'redirecciones', 'importaciones'])
			expect(await d1!.prepare(`select count(*) as n from ${t}`).first<number>('n')).toBe(0);
	});
});
