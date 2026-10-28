import { error } from '@sveltejs/kit';
import * as v from 'valibot';
import { datosDe } from '#lib/server/datos/index.ts';
import { errorTarjeta, TOPES, type Tarjeta } from '#lib/panel/tarjeta.ts';
import type { RequestHandler } from './$types';

const Esquema = v.object({
	modelo: v.picklist(['arco', 'sello', 'carta']),
	titulo: v.pipe(v.string(), v.maxLength(TOPES.titulo)),
	mensaje: v.pipe(v.string(), v.maxLength(TOPES.mensaje)),
	dorsoTitulo: v.pipe(v.string(), v.maxLength(TOPES.dorsoTitulo)),
	dorsoTexto: v.pipe(v.string(), v.maxLength(TOPES.dorsoTexto)),
	qr: v.picklist(['resena', 'cupon', 'instagram', 'whatsapp']),
	cupon: v.pipe(v.string(), v.maxLength(20)),
	instagram: v.pipe(v.string(), v.trim(), v.maxLength(30))
});

/** Guarda la tarjeta A6. */
export const PUT: RequestHandler = async ({ locals, request }) => {
	const r = v.safeParse(Esquema, await request.json().catch(() => null));
	if (!r.success) error(400, 'Revisá la tarjeta.');
	const t: Tarjeta = { ...r.output, instagram: r.output.instagram.replace(/^@/, '') };
	const mal = errorTarjeta(t);
	if (mal) error(400, mal);
	await datosDe(locals).guardarTarjeta(t);
	return new Response(null, { status: 204 });
};
