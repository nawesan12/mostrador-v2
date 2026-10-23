import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { afterEach, describe, expect, it } from 'vitest';
import { codigoPedido } from '../tienda/pedido.ts';
import { d1Local } from './d1-local.ts';
import { crear, DatosSandbox, dbSandbox } from './sandbox.ts';

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

async function tienda() {
	d1 = d1Local(migraciones);
	const db = dbSandbox(d1);
	const sb = (await crear(db, 'moda'))!;
	const datos = new DatosSandbox(db, sb);
	const producto = (await datos.tiendaPublica()).productos.find((p) => p.stock > 2)!;
	return { sb, datos, producto };
}

/** Una compra en la tienda online, como el checkout (Mercado Pago queda cobrado; transferencia, no). */
async function comprar(datos: DatosSandbox, productoId: string, pagado: boolean) {
	const r = await datos.crearPedido({
		items: [{ productoId, cantidad: 1 }],
		pago: pagado ? 'mp' : 'transf',
		canal: 'web',
		envio: 'retiro',
		cliente: 'Lucía Pérez',
		pagado
	});
	if (typeof r === 'string') throw new Error(r);
	return r.numero;
}

describe('reseñas y preguntas de la tienda online (demo)', () => {
	it('con el código de un pedido cobrado queda con compra, una sola vez', async () => {
		const { sb, datos, producto } = await tienda();
		const numero = await comprar(datos, producto.id, true);
		const codigo = await codigoPedido(sb.id, numero);
		const r = { codigo, autor: 'Lucía', estrellas: 5, texto: 'Llegó perfecto.' };
		expect(await datos.dejarResena(r)).toBe('ok');
		expect(await datos.dejarResena({ ...r, estrellas: 1 })).toBe('ya-reseno');

		const panel = await datos.resenas();
		const mia = panel.resenas.find((x) => x.autor === 'Lucía')!;
		expect(mia).toMatchObject({
			conCompra: true,
			oculta: false,
			respuesta: null,
			origen: 'Tienda'
		});
		// Primera: es la más nueva.
		expect(panel.resenas[0].id).toBe(mia.id);
		const publica = await datos.tiendaPublica();
		expect(publica.resenas[0]).toMatchObject({ autor: 'Lucía', conCompra: true, producto: '' });
	});

	it('sin código, con uno inventado o de un pedido sin cobrar: se guarda sin compra', async () => {
		const { sb, datos, producto } = await tienda();
		const sinCobrar = await comprar(datos, producto.id, false);
		const codigo = await codigoPedido(sb.id, sinCobrar);
		expect(await datos.dejarResena({ codigo, autor: 'Ana', estrellas: 4 })).toBe('ok');
		expect(await datos.dejarResena({ codigo, autor: 'Ana', estrellas: 4 })).toBe('ok');
		expect(
			await datos.dejarResena({ codigo: `${sinCobrar}-trucho`, autor: 'Beto', estrellas: 1 })
		).toBe('ok');
		expect(await datos.dejarResena({ productoId: producto.id, autor: 'Caro', estrellas: 3 })).toBe(
			'ok'
		);
		const panel = await datos.resenas();
		const nuevas = panel.resenas.filter((x) => ['Ana', 'Beto', 'Caro'].includes(x.autor));
		expect(nuevas).toHaveLength(4);
		expect(nuevas.every((x) => !x.conCompra)).toBe(true);
		expect(nuevas.find((x) => x.autor === 'Caro')!.producto).toBe(producto.nombre);
	});

	it('valida como Go: nombre, estrellas de 1 a 5 y texto corto', async () => {
		const { datos } = await tienda();
		expect(await datos.dejarResena({ autor: 'X', estrellas: 5 })).toBe('invalido');
		expect(await datos.dejarResena({ autor: 'Xi', estrellas: 6 })).toBe('invalido');
		expect(await datos.dejarResena({ autor: 'Xi', estrellas: 2.5 })).toBe('invalido');
		expect(await datos.dejarResena({ autor: 'Xi', estrellas: 3, texto: 'a'.repeat(1001) })).toBe(
			'invalido'
		);
	});

	it('ocultar no borra: sale de la tienda y del promedio, y se puede mostrar', async () => {
		const { datos } = await tienda();
		expect(await datos.dejarResena({ autor: 'Troll', estrellas: 1, texto: 'Malísimo' })).toBe('ok');
		const antes = await datos.resenas();
		const troll = antes.resenas.find((x) => x.autor === 'Troll')!;
		expect(antes.porEstrellas[4]).toBeGreaterThan(0);

		expect(await datos.accionResena(troll.id, 'ocultar')).toBe('ok');
		const oculta = await datos.resenas();
		expect(oculta.resenas.find((x) => x.id === troll.id)!.oculta).toBe(true);
		expect(oculta.total).toBe(antes.total - 1);
		expect(oculta.porEstrellas[4]).toBe(antes.porEstrellas[4] - 1);
		expect((await datos.tiendaPublica()).resenas.some((x) => x.autor === 'Troll')).toBe(false);

		expect(await datos.accionResena(troll.id, 'mostrar')).toBe('ok');
		expect((await datos.tiendaPublica()).resenas.some((x) => x.autor === 'Troll')).toBe(true);
		expect(await datos.accionResena('no-existe', 'ocultar')).toBe('no-existe');
	});

	it('las de ejemplo de la tienda online cuentan con compra; se responden y se ocultan', async () => {
		const { datos } = await tienda();
		const panel = await datos.resenas();
		const tiendaOnline = panel.resenas.filter((x) => x.origen === 'Tienda');
		expect(tiendaOnline.length).toBeGreaterThan(0);
		expect(tiendaOnline.every((x) => x.conCompra)).toBe(true);
		expect(panel.resenas.filter((x) => x.origen !== 'Tienda').every((x) => !x.conCompra)).toBe(
			true
		);
		const una = (await datos.tiendaPublica()).resenas[0];
		const id = panel.resenas.find((x) => x.texto === una.texto)!.id;
		expect(await datos.accionResena(id, 'responder', '¡Gracias por elegirnos!')).toBe('ok');
		expect((await datos.tiendaPublica()).resenas[0].respuesta).toBe('¡Gracias por elegirnos!');
		expect(await datos.accionResena(id, 'responder', '')).toBe('ok');
		expect((await datos.tiendaPublica()).resenas[0].respuesta).toBeNull();
		expect(await datos.accionResena(id, 'responder', 'a'.repeat(1001))).toBe('no-se-puede');
	});

	it('las preguntas de la ficha van al panel y, respondidas, a la tienda', async () => {
		const { datos, producto } = await tienda();
		const q = { productoId: producto.id, nombre: 'Lucía', pregunta: '¿Hay en azul?' };
		expect(await datos.preguntar({ ...q, productoId: 'no-existe' })).toBe('invalido');
		expect(await datos.preguntar({ ...q, pregunta: 'ok' })).toBe('invalido');
		expect(await datos.preguntar(q)).toBe('ok');
		expect((await datos.tiendaPublica()).preguntas).toEqual([]);

		const panel = await datos.resenas();
		const mia = panel.preguntas[0];
		expect(mia).toMatchObject({ producto: producto.nombre, respuesta: null, sugerida: '' });
		expect(mia.id.startsWith('t-')).toBe(true);
		expect(await datos.responderPregunta(mia.id, 'Sí, llega la semana que viene.')).toBe('ok');
		expect((await datos.tiendaPublica()).preguntas).toEqual([
			{
				productoId: producto.id,
				pregunta: '¿Hay en azul?',
				respuesta: 'Sí, llega la semana que viene.'
			}
		]);
		// Respondida, baja detrás de las que esperan respuesta.
		expect(await datos.preguntar({ ...q, pregunta: '¿Y en rojo?' })).toBe('ok');
		expect((await datos.resenas()).preguntas[0].pregunta).toBe('¿Y en rojo?');
	});

	it('la vitrina (otro sandbox) no ve lo del visitante', async () => {
		const { datos, sb } = await tienda();
		await datos.dejarResena({ autor: 'Lucía', estrellas: 5 });
		const vitrina = new DatosSandbox(dbSandbox(d1!), { ...sb, id: `vitrina-${sb.plantilla}` });
		expect((await vitrina.tiendaPublica(true)).resenas.some((x) => x.autor === 'Lucía')).toBe(
			false
		);
	});
});
