import { error, fail, redirect } from '@sveltejs/kit';
import * as v from 'valibot';
import { env } from 'cloudflare:workers';
import { codigoPedido } from '#lib/server/tienda/pedido.ts';
import { COOKIE, dbSandbox } from '#lib/server/sandbox/sandbox.ts';
import { avisarVenta } from '#lib/server/telegram.ts';
import { evaluarCupon } from '#lib/server/tienda/cupon.ts';
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
	pago: v.picklist(['transferencia', 'mercadopago']),
	cupon: v.optional(v.pipe(v.string(), v.trim(), v.maxLength(30)), '')
});

export const actions: Actions = {
	default: async ({ request, params, cookies, platform }) => {
		const r = await resolverTienda(params.tienda, cookies.get(COOKIE));
		if (!r) error(404, 'Esta tienda no existe.');
		if (r.tienda.vitrina)
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
		// Con la API real, el pedido lo arma Go (precios, envío y Mercado Pago salen de ahí).
		if (!r.datos || !r.sandboxId) {
			const res = await fetch(
				`${String(env.API_URL)}/v1/publico/tiendas/${r.tienda.slug}/pedidos`,
				{
					method: 'POST',
					headers: { 'content-type': 'application/json' },
					body: JSON.stringify(p)
				}
			);
			if (!res.ok) {
				const e = (await res.json().catch(() => null)) as { detail?: string } | null;
				return fail(res.status === 409 ? 409 : 400, {
					error: e?.detail ?? 'No pudimos armar el pedido.'
				});
			}
			const g = (await res.json()) as { codigo: string; url?: string };
			if (g.url && /^https:\/\/(www\.)?mercadopago\.com(\.ar)?\//.test(g.url)) redirect(303, g.url);
			redirect(303, `/t/${r.tienda.slug}/pedido/${g.codigo}`);
		}
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
		// Cupón: se vuelve a evaluar acá (el carrito sólo lo mostró).
		const cupon = p.cupon ? await evaluarCupon(r, p.cupon, p.email, p.carrito) : null;
		if (cupon && !cupon.ok) return fail(409, { error: cupon.motivo });
		const descCupon = cupon?.ok ? cupon.descuento : 0;
		const envio = cupon?.ok && cupon.envioGratis ? 0 : opcion.precio;
		const descuento =
			descCupon +
			(p.pago === 'transferencia' && (!cupon || (cupon.ok && cupon.combinable))
				? Math.round(((subtotal - descCupon) * r.tienda.cobros.transferencia) / 100)
				: 0);
		const total = subtotal + envio - descuento;
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
				cupon?.ok && `Cupón: ${p.cupon.toUpperCase()}`,
				`Total cobrado: $ ${total.toLocaleString('es-AR')}`
			]
				.filter(Boolean)
				.join(' · '),
			// En la demo, Mercado Pago se aprueba al toque; la transferencia queda esperando el comprobante.
			pagado: p.pago === 'mercadopago'
		}, {
			email: p.email,
			cupon: cupon?.ok ? p.cupon.toUpperCase() : null,
			descuentoCupon: descCupon
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
