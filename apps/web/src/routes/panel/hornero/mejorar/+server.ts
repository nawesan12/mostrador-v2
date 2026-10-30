import { env } from 'cloudflare:workers';
import { error, json } from '@sveltejs/kit';
import * as v from 'valibot';
import { datosDe, usaSandbox } from '#lib/server/datos/index.ts';
import { cupoDemo, diaAr, POR_VISITANTE } from '#lib/server/hornero/cupo.ts';
import { MODELO } from '#lib/server/hornero/modelo.ts';
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
/** Con una tienda real, si el modelo no está o falla, se avisa y la descripción queda como estaba. */
const sinTincho = () =>
	json({ error: 'Tincho no pudo escribir ahora. Probá de nuevo en un rato.' }, { status: 503 });

/** "✦ Mejorar con Tincho": reescribe la descripción de un producto (cuenta en el mismo cupo que la charla). */
export const POST: RequestHandler = async ({ locals, request }) => {
	const r = v.safeParse(Pedido, await request.json().catch(() => null));
	if (!r.success) error(400, 'Falta el nombre del producto.');
	const { nombre, descripcion } = r.output;
	const tienda = await datosDe(locals).tienda();
	let conModelo = env.HORNERO === 'workers-ai' && !!env.AI;
	if (usaSandbox() && locals.sandbox) {
		const n = await sumarUsoHornero(dbSandbox(env.SANDBOX), locals.sandbox.id, diaAr());
		if (n > POR_VISITANTE || (conModelo && !(await cupoDemo()))) conModelo = false;
	}
	if (!conModelo || !env.AI)
		return usaSandbox() ? json({ texto: simulado(nombre, descripcion) }) : sinTincho();
	try {
		const res = (await env.AI.run(
			MODELO as Parameters<Ai['run']>[0],
			{
				messages: [
					{
						role: 'system',
						content: `Escribís descripciones de productos para «${tienda.nombre}», una tienda online argentina. Español rioplatense, claro y vendedor sin exagerar. Entre 2 y 4 oraciones, sin títulos, sin listas, sin emojis. No inventes medidas, materiales ni datos que no estén en lo que te pasan.`
					},
					{
						role: 'user',
						content: `Producto: ${nombre}\nDescripción actual: ${descripcion || '(vacía)'}`
					}
				],
				max_tokens: 220,
				temperature: 0.5,
				chat_template_kwargs: { enable_thinking: false }
			} as never
		)) as { response?: string; choices?: { message?: { content?: string } }[] };
		const texto = (res.response ?? res.choices?.[0]?.message?.content ?? '').trim();
		return texto ? json({ texto }) : sinTincho();
	} catch (e) {
		console.error('mejorar', e);
		return sinTincho();
	}
};
