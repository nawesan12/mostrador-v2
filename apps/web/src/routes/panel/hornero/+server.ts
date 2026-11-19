import { env } from 'cloudflare:workers';
import { error, isHttpError, json } from '@sveltejs/kit';
import * as v from 'valibot';
import { datosDe, usaSandbox } from '#lib/server/datos/index.ts';
import type { DatosHornero } from '#lib/server/hornero/contexto.ts';
import { diaAr, POR_VISITANTE } from '#lib/server/hornero/cupo.ts';
import { ideasDe } from '#lib/server/hornero/ideas.ts';
import { conIa, NO_DISPONIBLE } from '#lib/server/hornero/tincho.ts';
import { responderSimulado } from '#lib/server/hornero/simulado.ts';
import { ayudaConBotones, responder } from '#lib/server/hornero/motor.ts';
import { dbSandbox, sumarUsoHornero } from '#lib/server/sandbox/sandbox.ts';
import type { RequestHandler } from './$types';

/** Lo que muestra Tincho al abrirse: cómo viene el día y sus ideas. */
export const GET: RequestHandler = async ({ locals }) => {
	const d = datosDe(locals);
	const real = !usaSandbox();
	// Con la API, los primeros pasos (como «Para hacer hoy»): si tiene productos y cómo cobrar. Si
	// algo falla, esa idea no sale (mejor callar que decir algo que no es).
	const [t, productos, config, integ] = await Promise.all([
		d.tablero('hoy'),
		real
			? d
					.listar('productos', { desde: 0, cuantos: 1 })
					.then((l) => l.total)
					.catch(() => null)
			: null,
		real ? d.configuracion().catch(() => null) : null,
		real ? d.integraciones().catch(() => null) : null
	]);
	const cobra =
		config && integ ? !!(config.negocio.alias || config.negocio.cbu || integ.conectadas.mp) : null;
	// Sin montos (Ventas, Local): Tincho saluda con los pedidos de hoy, no con lo vendido.
	return json({
		hola: t.hola,
		pedidosHoy: t.kpis.pedidos.actual,
		...(t.ventas ? { ventasHoy: t.ventas.actual, ventasAyer: t.ventas.anterior } : {}),
		vacia: productos === 0,
		conVentas: t.kpis.pedidos.actual > 0 || (t.masVendidos?.length ?? 0) > 0,
		ideas: ideasDe(t, real, { productos, cobra })
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
	),
	/**
	 * «bloques»: la respuesta en JSON, con cifras, gráficos, tablas, botones, sugerencias y las
	 * propuestas de Tincho para confirmar. «texto»: sólo el texto (clientes viejos).
	 */
	formato: v.optional(v.picklist(['texto', 'bloques']), 'texto')
});

const texto = (t: string) =>
	new Response(t, {
		headers: { 'content-type': 'text/plain; charset=utf-8', 'cache-control': 'no-store' }
	});
const NO_STORE = { headers: { 'cache-control': 'no-store' } };

/**
 * Una pregunta a Tincho: primero el motor propio (sin IA, al instante); lo que no resuelve va a la
 * API Go (Claude Haiku, con los cupos del plan). En la demo no hay IA: respuestas armadas.
 */
export const POST: RequestHandler = async ({ locals, request, platform }) => {
	const cuerpo = v.safeParse(Pedido, await request.json().catch(() => null));
	if (!cuerpo.success) error(400, 'Escribile algo a Tincho.');
	const datos = datosDe(locals);
	const bloques = cuerpo.output.formato === 'bloques';
	const enviar = (r: { texto: string }) => (bloques ? json(r, NO_STORE) : texto(r.texto));

	// En la demo pública, un tope por visitante (ver cupo.ts).
	if (usaSandbox() && locals.sandbox) {
		const n = await sumarUsoHornero(dbSandbox(env.SANDBOX), locals.sandbox.id, diaAr());
		if (n > POR_VISITANTE)
			return enviar({
				texto:
					'Por hoy ya charlamos un montón en la demo. Mañana sigo respondiendo; mientras, recorré el panel con los datos de ejemplo.'
			});
	}

	const [tienda, hoy, semana, mes, contexto] = await Promise.all([
		datos.tienda(),
		datos.tablero('hoy'),
		datos.tablero('7d'),
		datos.tablero('30d').catch(() => undefined),
		datos.contextoHornero()
	]);
	const d: DatosHornero = { tienda, hola: hoy.hola, hoy, semana, mes, contexto };
	const mensajes = cuerpo.output.mensajes;
	const pregunta = mensajes[mensajes.length - 1].texto;
	const anteriores = mensajes
		.slice(0, -1)
		.filter((m) => m.rol === 'persona')
		.map((m) => m.texto);

	// Primero el motor propio: las preguntas de todos los días salen de los datos, sin IA.
	const directa = responder(pregunta, d, anteriores);
	if (directa) {
		if (!usaSandbox()) {
			// Se cuenta para «Respuestas automáticas sin IA»; no hace falta esperarlo.
			const cuenta = datos.contarMotorTincho().catch(() => {});
			if (platform?.ctx) platform.ctx.waitUntil(cuenta);
		}
		return enviar(directa);
	}
	if (usaSandbox()) return enviar({ ...ayudaConBotones(), texto: responderSimulado(pregunta, d) });

	const r = await datos.preguntarTincho(pregunta, mensajes.slice(0, -1)).catch((e) => {
		// La sesión vencida sigue siendo un 401: el panel manda a ingresar.
		if (isHttpError(e) && e.status === 401) throw e;
		console.error('tincho', e instanceof Error ? e.message : e);
		return 'no-disponible' as const;
	});
	if (r === 'no-disponible') return enviar({ ...ayudaConBotones(), texto: NO_DISPONIBLE });
	return enviar(conIa(r));
};
