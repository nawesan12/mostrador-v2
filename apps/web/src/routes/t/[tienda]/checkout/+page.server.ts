import { error, fail, redirect } from '@sveltejs/kit';
import * as v from 'valibot';
import { env } from 'cloudflare:workers';
import { codigoPedido } from '#lib/server/tienda/pedido.ts';
import { COOKIE, dbSandbox } from '#lib/server/sandbox/sandbox.ts';
import { avisarVenta } from '#lib/server/telegram.ts';
import { resolverTienda } from '#lib/server/tienda/publica.ts';
import { cotizar, cpValido } from '#lib/tienda/envio.ts';
import type { Actions } from './$types';

const Pedido = v.object({
	carrito: v.pipe(
		v.array(
			v.object({
				id: v.pipe(v.string(), v.maxLength(80)),
				cantidad: v.pipe(v.number(), v.integer(), v.minValue(1), v.maxValue(99))
			})
		),
		v.minLength(1),
		v.maxLength(50)
	),
	email: v.pipe(v.string(), v.trim(), v.email('Revisá el email.'), v.maxLength(120)),
	nombre: v.pipe(v.string(), v.trim(), v.minLength(2, 'Poné tu nombre.'), v.maxLength(80)),
	telefono: v.pipe(v.string(), v.trim(), v.regex(/^[\d\s+()-]{8,20}$/, 'Revisá el teléfono.')),
	dni: v.pipe(v.string(), v.trim(), v.maxLength(12)),
	cp: v.pipe(v.string(), v.trim(), v.maxLength(4)),
	direccion: v.pipe(v.string(), v.trim(), v.maxLength(160)),
	envio: v.picklist(['andreani', 'correo', 'moto', 'retiro'], 'Elegí cómo te llega.'),
	pago: v.picklist(['transferencia', 'mercadopago'])
});

export const actions: Actions = {
	default: async ({ request, params, cookies, platform }) => {
		const r = await resolverTienda(params.tienda, cookies.get(COOKIE));
		if (!r) error(404, 'Esta tienda no existe.');
		if (r.tienda.vitrina || !r.datos || !r.sandboxId)
			return fail(409, { error: 'Es una tienda de ejemplo: no se compra de verdad.' });
		const f = await request.formData();
		let carrito: unknown = null;
		try {
			carrito = JSON.parse(String(f.get('carrito') ?? '[]'));
		} catch {
			// queda null y falla la validación
		}
		const b = v.safeParse(Pedido, { ...Object.fromEntries(f), carrito });
		if (!b.success) return fail(400, { error: b.issues[0].message });
		const p = b.output;
		if (p.envio !== 'retiro' && (!cpValido(p.cp) || p.direccion.length < 4))
			return fail(400, { error: 'Poné el código postal y la dirección de entrega.' });
		// Los precios y el envío salen del servidor, nunca del navegador.
		const porId = new Map(r.tienda.productos.map((x) => [x.id, x]));
		let subtotal = 0;
		for (const l of p.carrito) {
			const x = porId.get(l.id);
			if (!x)
				return fail(409, { error: 'Algún producto ya no está publicado. Revisá el carrito.' });
			subtotal += x.precio * l.cantidad;
		}
		const opcion = cotizar(p.envio === 'retiro' ? '0000' : p.cp, subtotal, r.tienda).find(
			(o) => o.id === p.envio
		);
		if (!opcion) return fail(400, { error: 'Esa forma de entrega no llega a tu código postal.' });
		const descuento =
			p.pago === 'transferencia' ? Math.round((subtotal * r.tienda.cobros.transferencia) / 100) : 0;
		const total = subtotal + opcion.precio - descuento;
		const res = await r.datos.crearPedido({
			items: p.carrito.map((l) => ({ productoId: l.id, cantidad: l.cantidad })),
			pago: p.pago === 'transferencia' ? 'transf' : 'mp',
			canal: 'web',
			envio: p.envio,
			cliente: p.nombre,
			telefono: p.telefono,
			nota: [
				`Email: ${p.email}`,
				p.dni && `DNI: ${p.dni}`,
				p.envio !== 'retiro' && `Entrega: ${p.direccion} (${p.cp})`,
				`Total cobrado: $ ${total.toLocaleString('es-AR')}`
			]
				.filter(Boolean)
				.join(' · '),
			// En la demo, Mercado Pago se aprueba al toque; la transferencia queda esperando el comprobante.
			pagado: p.pago === 'mercadopago'
		});
		if (res === 'sin-stock')
			return fail(409, { error: 'Algún producto se quedó sin stock. Revisá el carrito.' });
		if (res === 'invalido') return fail(400, { error: 'Revisá el carrito.' });
		const aviso = avisarVenta(
			dbSandbox(env.SANDBOX),
			r.sandboxId,
			`🛒 Nueva venta en la tienda: pedido #${res.numero} de ${p.nombre}.`
		).catch(() => {});
		if (platform?.ctx) platform.ctx.waitUntil(aviso);
		const q = new URLSearchParams({ pago: p.pago, total: String(total) });
		redirect(303, `/t/${r.tienda.slug}/pedido/${await codigoPedido(r.sandboxId, res.numero)}?${q}`);
	}
};
