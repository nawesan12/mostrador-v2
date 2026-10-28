import { error, json } from '@sveltejs/kit';
import * as v from 'valibot';
import { env } from 'cloudflare:workers';
import { datosDe } from '#lib/server/datos/index.ts';
import { dbSandbox } from '#lib/server/sandbox/sandbox.ts';
import { avisarVenta } from '#lib/server/telegram.ts';
import type { RequestHandler } from './$types';

const Nuevo = v.object({
	items: v.pipe(
		v.array(
			v.object({
				productoId: v.pipe(v.string(), v.maxLength(80)),
				varianteId: v.optional(v.nullable(v.pipe(v.string(), v.maxLength(120))), null),
				cantidad: v.pipe(v.number(), v.integer(), v.minValue(1), v.maxValue(10_000))
			})
		),
		v.minLength(1, 'Sumá al menos un producto.'),
		v.maxLength(100)
	),
	pago: v.picklist(['mp', 'tarjeta', 'payway', 'modo', 'uala', 'transf', 'efectivo']),
	canal: v.picklist(['pos', 'wa', 'web', 'ig', 'ml', 'tg']),
	envio: v.optional(v.picklist(['retiro', 'andreani', 'oca', 'correo', 'moto'])),
	cliente: v.optional(v.nullable(v.pipe(v.string(), v.maxLength(80)))),
	clienteId: v.optional(v.nullable(v.pipe(v.string(), v.maxLength(80)))),
	telefono: v.optional(v.nullable(v.pipe(v.string(), v.maxLength(30)))),
	nota: v.optional(v.nullable(v.pipe(v.string(), v.maxLength(300)))),
	pagado: v.optional(v.boolean(), true)
});

/** Un pedido nuevo: venta en el local o cargado a mano. */
export const POST: RequestHandler = async ({ locals, request, platform }) => {
	const r = v.safeParse(Nuevo, await request.json().catch(() => null));
	if (!r.success) error(400, r.issues[0].message);
	const res = await datosDe(locals).crearPedido({ ...r.output, envio: r.output.envio ?? 'retiro' });
	if (res === 'sin-stock') error(409, 'No alcanza el stock de algún producto.');
	if (res === 'invalido')
		error(400, 'Hay un producto que no existe o al que le falta elegir la variante.');
	// Aviso de la venta por Telegram (si la tienda vinculó su chat). No demora la respuesta.
	if (locals.sandbox) {
		const total = '$ ' + Math.round(res.total).toLocaleString('es-AR');
		const canal = {
			pos: 'en el local',
			wa: 'por WhatsApp',
			web: 'en la tienda',
			ig: 'por Instagram',
			ml: 'en Mercado Libre',
			tg: 'por Telegram'
		}[r.output.canal];
		const aviso = avisarVenta(
			dbSandbox(env.SANDBOX),
			locals.sandbox.id,
			`🛒 Nueva venta ${canal}: pedido #${res.numero} por ${total}.`
		).catch(() => {});
		if (platform?.ctx) platform.ctx.waitUntil(aviso);
	}
	return json(res, { status: 201 });
};
