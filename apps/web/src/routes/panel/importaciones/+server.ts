import { error, json } from '@sveltejs/kit';
import * as v from 'valibot';
import { datosDe } from '#lib/server/datos/index.ts';
import { PLATAFORMAS } from '#lib/server/importacion/fuente.ts';
import type { RequestHandler } from './$types';

/**
 * La importación desde otra plataforma. GET: cómo va. POST: empezar, avanzar o cancelar. En la demo
 * no hay outbox: cada "avanzar" hace un paso (el navegador lo pide hasta que termina); con la API,
 * avanza sola y "avanzar" sólo lee cómo va.
 */
export const GET: RequestHandler = async ({ locals }) =>
	json(await datosDe(locals).importacionActual(), { headers: { 'cache-control': 'no-store' } });

const Pedido = v.variant('accion', [
	v.object({
		accion: v.literal('crear'),
		plataforma: v.picklist(PLATAFORMAS),
		url: v.optional(v.pipe(v.string(), v.trim(), v.maxLength(300))),
		credencial: v.optional(v.pipe(v.string(), v.maxLength(4096)))
	}),
	v.object({ accion: v.literal('avanzar') }),
	v.object({ accion: v.literal('cancelar'), id: v.pipe(v.string(), v.maxLength(64)) })
]);

const TEXTOS = {
	invalido: [400, 'Revisá la plataforma y la dirección de tu tienda.'],
	'sin-fuente': [409, 'Todavía no podemos importar desde esa plataforma.'],
	activa: [409, 'Ya hay una importación en curso: esperá que termine o cancelala.'],
	'no-existe': [404, 'Esa importación no existe.'],
	terminada: [409, 'Esa importación ya terminó.']
} as const;

export const POST: RequestHandler = async ({ locals, request }) => {
	const r = v.safeParse(Pedido, await request.json().catch(() => null));
	if (!r.success) error(400, 'El pedido no es válido.');
	const datos = datosDe(locals);
	const p = r.output;
	const res =
		p.accion === 'crear'
			? await datos.crearImportacion({
					plataforma: p.plataforma,
					url: p.url,
					credencial: p.credencial
				})
			: p.accion === 'cancelar'
				? await datos.cancelarImportacion(p.id)
				: await datos.avanzarImportacion();
	if (typeof res === 'string') error(TEXTOS[res][0], TEXTOS[res][1]);
	return json(res, { headers: { 'cache-control': 'no-store' } });
};
