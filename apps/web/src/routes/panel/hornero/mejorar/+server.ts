import { env } from 'cloudflare:workers';
import { error, json } from '@sveltejs/kit';
import * as v from 'valibot';
import { datosDe, usaSandbox } from '#lib/server/datos/index.ts';
import { diaAr } from '#lib/server/hornero/cupo.ts';
import { dbSandbox, sumarUsoHornero } from '#lib/server/sandbox/sandbox.ts';
import type { RequestHandler } from './$types';

const Pedido = v.object({
	nombre: v.pipe(v.string(), v.trim(), v.minLength(2), v.maxLength(120)),
	descripcion: v.pipe(v.string(), v.maxLength(4000))
});

/**
 * Texto armado sin modelo, sólo para la demo y las pruebas: repite lo que hay, sin agregar promesas
 * (envíos, cambios) que la tienda no dijo.
 */
const simulado = (nombre: string, d: string) =>
	d.trim() ? `${nombre}. ${d.trim().replace(/\.?$/, '.')}` : `${nombre}.`;
/** Con una tienda real, si la IA no está o falla, se avisa y la descripción queda como estaba. */
const sinTincho = (texto = 'Tincho no pudo escribir ahora. Probá de nuevo en un rato.') =>
	json({ error: texto }, { status: 503 });

/** "✦ Mejorar con Tincho": reescribe la descripción de un producto (cuenta en el mismo cupo que la charla). */
export const POST: RequestHandler = async ({ locals, request }) => {
	const r = v.safeParse(Pedido, await request.json().catch(() => null));
	if (!r.success) error(400, 'Falta el nombre del producto.');
	const { nombre, descripcion } = r.output;
	if (usaSandbox()) {
		// En la demo cuenta en el mismo tope que la charla; pasado el tope, igual que antes: el texto armado.
		if (locals.sandbox) await sumarUsoHornero(dbSandbox(env.SANDBOX), locals.sandbox.id, diaAr());
		return json({ texto: simulado(nombre, descripcion) });
	}
	const res = await datosDe(locals).mejorarDescripcion(nombre, descripcion);
	if (res === 'no-disponible') return sinTincho();
	// Con el cupo agotado, el aviso del tope (la descripción queda como estaba).
	if (res.fuente === 'tope') return sinTincho(res.texto);
	return json({ texto: res.texto });
};
