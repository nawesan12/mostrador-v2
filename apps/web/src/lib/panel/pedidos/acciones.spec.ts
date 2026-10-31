import { describe, expect, it } from 'vitest';
import { accionRapida, mensajeWhatsapp, primerNombre, whatsappDe } from './acciones.ts';
import { clienteDe, volverAVender } from './volver-a-vender.ts';

const ctx = {
	retiro: { direccion: 'Gorriti 4520, Palermo', horario: 'Lun a sáb de 10 a 19 h.' },
	slug: 'luna'
};
const p = {
	numero: 12,
	cliente: 'Ana Pérez',
	estado: 'Sin pagar' as const,
	envio: 'retiro',
	pago: 'transf',
	total: 25000,
	codigo: null,
	telefono: '11 5555-0000'
};

describe('acción de un toque', () => {
	it('sigue el estado y la entrega reales', () => {
		expect(accionRapida(p)).toMatchObject({ texto: 'Marcar pagado', a: 'Para preparar' });
		expect(accionRapida({ ...p, pago: 'mp' })).toBeNull();
		expect(accionRapida({ ...p, estado: 'Para preparar' })?.texto).toBe('Listo para retirar');
		expect(accionRapida({ ...p, estado: 'Para preparar', envio: 'moto' })?.texto).toBe('Preparar');
		expect(accionRapida({ ...p, estado: 'Listo para despachar' })).toMatchObject({
			texto: 'Retirado',
			a: 'Entregado'
		});
		expect(accionRapida({ ...p, estado: 'Listo para despachar', envio: 'correo' })).toMatchObject({
			texto: 'Despachado',
			a: 'En camino'
		});
		expect(accionRapida({ ...p, estado: 'En camino', envio: 'moto' })?.texto).toBe('Entregado');
		expect(accionRapida({ ...p, estado: 'Entregado' })).toBeNull();
		expect(accionRapida({ ...p, estado: 'Cancelado' })).toBeNull();
	});
});

describe('WhatsApp a quien compró', () => {
	it('arma el mensaje según el estado', () => {
		expect(mensajeWhatsapp({ ...p, estado: 'Listo para despachar' }, ctx)).toBe(
			'¡Hola Ana! Tu pedido #12 está listo para retirar en Gorriti 4520, Palermo (Lun a sáb de 10 a 19 h).'
		);
		expect(
			mensajeWhatsapp(
				{ ...p, estado: 'Listo para despachar' },
				{ ...ctx, retiro: { direccion: '', horario: '' } }
			)
		).toBe('¡Hola Ana! Tu pedido #12 está listo para retirar.');
		expect(mensajeWhatsapp({ ...p, estado: 'En camino', envio: 'moto' }, ctx)).toBe(
			'¡Hola Ana! Tu pedido #12 ya salió para tu casa.'
		);
		expect(
			mensajeWhatsapp({ ...p, estado: 'En camino', envio: 'correo' }, { ...ctx, seguimiento: '36' })
		).toBe('¡Hola Ana! Tu pedido #12 ya está en camino. Número de seguimiento: 36.');
		expect(mensajeWhatsapp({ ...p, estado: 'Para preparar' }, ctx)).toBe(
			'¡Hola Ana! Recibimos tu pago del pedido #12. Ya lo estamos preparando.'
		);
		expect(mensajeWhatsapp(p, ctx)).toContain('todavía no nos llegó el pago');
	});

	it('al entregarse pide la reseña sólo si hay dónde dejarla', () => {
		expect(mensajeWhatsapp({ ...p, estado: 'Entregado', codigo: 'abc123' }, ctx)).toBe(
			'¡Hola Ana! ¿Te pudimos ayudar? Si querés, dejanos una reseña: https://tulocalweb.com.ar/t/luna/pedido/abc123'
		);
		expect(mensajeWhatsapp({ ...p, estado: 'Entregado' }, ctx)).not.toContain('reseña');
	});

	it('wa.me con el número normalizado; sin teléfono no se ofrece', () => {
		expect(whatsappDe(p, ctx)).toMatch(/^https:\/\/wa\.me\/5491155550000\?text=/);
		expect(whatsappDe({ ...p, telefono: null }, ctx)).toBeNull();
		expect(whatsappDe({ ...p, telefono: '123' }, ctx)).toBeNull();
	});

	it('sin nombre de verdad saluda sin nombre', () => {
		expect(primerNombre('Consumidor final')).toBe('');
		expect(mensajeWhatsapp({ ...p, cliente: 'Consumidor final' }, ctx)).toMatch(/^¡Hola! /);
	});
});

describe('volver a vender', () => {
	const productos = [
		{ id: 'mate', nombre: 'Mate', sku: 'M1', precio: 12000, stock: 5 },
		{ id: 'termo', nombre: 'Termo', sku: 'T1', precio: 30000, stock: 1 },
		{ id: 'yerba', nombre: 'Yerba', sku: 'Y1', precio: 5000, stock: 0 },
		{
			id: 'remera',
			nombre: 'Remera',
			sku: 'R',
			precio: 20000,
			stock: 3,
			variantes: [
				{ id: 'm', texto: 'Talle M', sku: '', precio: null, stock: 3, activa: true },
				{ id: 'l', texto: 'Talle L', sku: '', precio: null, stock: 0, activa: false }
			]
		}
	];
	const it0 = { varianteId: null, precio: 12000, cantidad: 2 };

	it('usa el precio y el stock de hoy y avisa lo que cambió', () => {
		const r = volverAVender(
			[
				{ ...it0, productoId: 'mate', nombre: 'Mate' },
				{ productoId: 'termo', varianteId: null, nombre: 'Termo', cantidad: 3, precio: 28000 },
				{ ...it0, productoId: 'yerba', nombre: 'Yerba', precio: 5000 },
				{ ...it0, productoId: 'remera', varianteId: 'm', nombre: 'Remera', precio: 20000 },
				{ ...it0, productoId: 'remera', varianteId: 'l', nombre: 'Remera', variante: 'Talle L' },
				{ ...it0, productoId: 'borrado', nombre: 'Bombilla' }
			],
			productos
		);
		expect(r.renglones).toEqual([
			{ clave: 'mate', cantidad: 2 },
			{ clave: 'termo', cantidad: 1 },
			{ clave: 'remera:m', cantidad: 2 }
		]);
		expect(r.avisos).toEqual([
			'De «Termo» queda 1: sumamos 1.',
			'«Termo» cambió de precio: $\u00a028.000 → $\u00a030.000.',
			'«Yerba» no tiene stock.',
			'«Remera · Talle L» ya no está a la venta.',
			'«Bombilla» ya no está a la venta.'
		]);
	});

	it('los pedidos de antes se buscan por el nombre', () => {
		const r = volverAVender([{ ...it0, productoId: null, nombre: 'mate ' }], productos);
		expect(r).toEqual({ renglones: [{ clave: 'mate', cantidad: 2 }], avisos: [] });
	});

	it('encuentra al cliente por WhatsApp o por nombre', () => {
		const clientes = [
			{ id: 'a', nombre: 'Ana Pérez', telefono: '+54 9 11 5555 0000' },
			{ id: 'b', nombre: 'Beto', telefono: null }
		];
		expect(clienteDe({ nombre: 'Ana', telefono: '11 5555-0000' }, clientes)?.id).toBe('a');
		expect(clienteDe({ nombre: 'beto', telefono: null }, clientes)?.id).toBe('b');
		expect(clienteDe({ nombre: 'Carla', telefono: null }, clientes)).toBeNull();
	});
});
