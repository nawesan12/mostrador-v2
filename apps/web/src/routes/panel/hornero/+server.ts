import { env } from 'cloudflare:workers';
import { error, json } from '@sveltejs/kit';
import * as v from 'valibot';
import { datosDe, usaSandbox } from '#lib/server/datos/index.ts';
import { sistema, type DatosHornero } from '#lib/server/hornero/contexto.ts';
import { cupoDemo, diaAr, POR_VISITANTE } from '#lib/server/hornero/cupo.ts';
import { ideasDe } from '#lib/server/hornero/ideas.ts';
import { responderConIa, type Mensaje } from '#lib/server/hornero/modelo.ts';
import { responderSimulado } from '#lib/server/hornero/simulado.ts';
import { ayuda, responderSinIa } from '#lib/server/hornero/motor.ts';
import { dbSandbox, sumarUsoHornero } from '#lib/server/sandbox/sandbox.ts';
import type { RequestHandler } from './$types';

/** Lo que muestra Tincho al abrirse: cómo viene el día y sus ideas. */
export const GET: RequestHandler = async ({ locals }) => {
	const t = await datosDe(locals).tablero('hoy');
	return json({
		hola: t.hola,
		ventasHoy: t.ventas.actual,
		ventasAyer: t.ventas.anterior,
		ideas: ideasDe(t, !usaSandbox())
	});
};

const Pedido = v.object({
	mensajes: v.pipe(
		v.array(
			v.object({
				rol: v.picklist(['persona', 'hornero']),
				texto: v.pipe(v.string(), v.trim(), v.minLength(1), v.maxLength(1000))
			})
		),
		v.minLength(1),
		v.maxLength(12),
		v.check((m) => m[m.length - 1].rol === 'persona', 'El último mensaje es de la persona.')
	)
});

const texto = (t: string) =>
	new Response(t, {
		headers: { 'content-type': 'text/plain; charset=utf-8', 'cache-control': 'no-store' }
	});

/** Una pregunta a Tincho: responde en texto, a medida que el modelo escribe. */
export const POST: RequestHandler = async ({ locals, request }) => {
	const cuerpo = v.safeParse(Pedido, await request.json().catch(() => null));
	if (!cuerpo.success) error(400, 'Escribile algo a Tincho.');
	const datos = datosDe(locals);

	// En la demo pública hay topes: por visitante y para todos (ver cupo.ts).
	let conModelo = env.HORNERO === 'workers-ai' && !!env.AI;
	if (usaSandbox() && locals.sandbox) {
		const n = await sumarUsoHornero(dbSandbox(env.SANDBOX), locals.sandbox.id, diaAr());
		if (n > POR_VISITANTE)
			return texto(
				'Por hoy ya charlamos un montón en la demo. Mañana sigo respondiendo; mientras, recorré el panel con los datos de ejemplo.'
			);
		if (conModelo && !(await cupoDemo())) conModelo = false;
	}

	const [tienda, hoy, semana, contexto] = await Promise.all([
		datos.tienda(),
		datos.tablero('hoy'),
		datos.tablero('7d'),
		datos.contextoHornero()
	]);
	const d: DatosHornero = { tienda, hola: hoy.hola, hoy, semana, contexto };
	const mensajes = cuerpo.output.mensajes;
	const pregunta = mensajes[mensajes.length - 1].texto;

	// Primero el motor propio: las preguntas de todos los días salen de los datos, sin IA.
	const directa = responderSinIa(pregunta, d);
	if (directa) return texto(directa);
	if (!conModelo || !env.AI)
		return texto(env.HORNERO === 'simulado' ? responderSimulado(pregunta, d) : ayuda());

	const conversacion: Mensaje[] = [
		{ role: 'system', content: sistema(d) },
		...mensajes.map((m): Mensaje => ({
			role: m.rol === 'persona' ? 'user' : 'assistant',
			content: m.texto
		}))
	];
	try {
		return new Response(await responderConIa(env.AI, conversacion), {
			headers: { 'content-type': 'text/plain; charset=utf-8', 'cache-control': 'no-store' }
		});
	} catch (e) {
		console.error('hornero', e);
		return texto('Uy, se me cortó la conexión. ¿Me lo preguntás de nuevo?');
	}
};
