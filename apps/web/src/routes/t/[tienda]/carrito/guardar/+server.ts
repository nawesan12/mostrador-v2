import { error } from '@sveltejs/kit';
import * as v from 'valibot';
import { clienteDe } from '#lib/server/datos/index.ts';
import { COOKIE } from '#lib/server/sandbox/sandbox.ts';
import { guardarCarrito } from '#lib/server/tienda/carritos.ts';
import { resolverTienda } from '#lib/server/tienda/publica.ts';
import type { RequestHandler } from './$types';

const Guardar = v.object({
	email: v.pipe(v.string(), v.trim(), v.email(), v.maxLength(120)),
	nombre: v.optional(v.pipe(v.string(), v.trim(), v.maxLength(80)), ''),
	carrito: v.pipe(
		v.array(
			v.object({
				id: v.pipe(v.string(), v.maxLength(80)),
				varianteId: v.optional(v.nullable(v.pipe(v.string(), v.maxLength(120))), null),
				cantidad: v.pipe(v.number(), v.integer(), v.minValue(1), v.maxValue(99))
			})
		),
		v.minLength(1),
		v.maxLength(50)
	)
});

/** El checkout guarda el carrito cuando quien compra deja su email (para el recordatorio). */
export const POST: RequestHandler = async (event) => {
	const { params, cookies, request } = event;
	const texto = await request.text();
	if (texto.length > 16 * 1024) error(413, 'El carrito es demasiado grande.');
	let cuerpo: unknown = null;
	try {
		cuerpo = JSON.parse(texto);
	} catch {
		// queda null y falla la validación
	}
	const b = v.safeParse(Guardar, cuerpo);
	if (!b.success) error(400, 'Revisá el carrito.');
	const r = await resolverTienda(params.tienda, cookies.get(COOKIE));
	if (!r) error(404, 'Esta tienda no existe.');
	const ok = await guardarCarrito(
		r,
		clienteDe(event),
		b.output.email,
		b.output.nombre,
		b.output.carrito
	).catch(() => false);
	return new Response(null, { status: ok ? 204 : 202 });
};
