import { error, json } from '@sveltejs/kit';
import * as v from 'valibot';
import { datosDe } from '#lib/server/datos/index.ts';
import { llamar } from '#lib/server/mcp/herramientas.ts';
import type { RequestHandler } from './$types';

const Pedido = v.object({
	herramienta: v.picklist([
		'cambiar_estado_pedido',
		'ajustar_producto',
		'ajustar_precios',
		'crear_cupon',
		'crear_link_de_pago'
	]),
	argumentos: v.record(v.string(), v.unknown())
});

/**
 * «Confirmar» una propuesta de Tincho: la hace la persona, con su sesión. Se ejecuta con la misma
 * herramienta del MCP (que valida los argumentos) y la API aplica el rol de la persona como en
 * cualquier otra pantalla del panel. Los interruptores de Integraciones › IA son para las IA
 * conectadas, no para esto: acá confirma la persona.
 */
export const POST: RequestHandler = async ({ locals, request, url }) => {
	const r = v.safeParse(Pedido, await request.json().catch(() => null));
	if (!r.success) error(400, 'Esa propuesta no es válida.');
	const datos = datosDe(locals);
	// El id de la propuesta (Idempotency-Key): se toma antes de ejecutar y se guarda el resultado, así
	// un reintento (se cortó la red, otro dispositivo) no suma el stock dos veces ni crea otro link.
	const clave = request.headers.get('idempotency-key') ?? '';
	const sinCache = { 'cache-control': 'no-store' };
	if (clave) {
		if (!/^[A-Za-z0-9_-]{8,64}$/.test(clave)) error(400, 'Esa propuesta no es válida.');
		const t = await datos.tomarConfirmacion(clave);
		if (t.estado === 'hecha')
			return json({ ok: true, resultado: t.resultado ?? {} }, { headers: sinCache });
		if (t.estado === 'en_curso')
			return json(
				{ ok: false, incierto: true, error: 'Se está haciendo. Fijate en un momento.' },
				{ status: 409, headers: sinCache }
			);
	}
	// Si esto se corta a la mitad, la confirmación queda «en_curso» y vence sola (tincho/confirmaciones.go).
	const res = await llamar(
		{
			datos,
			permisos: ['lectura', 'acciones'],
			sitio: url.origin,
			permisosPanel: { escribir: true, plata: true }
		},
		r.output.herramienta,
		r.output.argumentos
	);
	if (res.isError) {
		if (clave) await datos.terminarConfirmacion(clave, false).catch(() => {});
		return json(
			{ ok: false, error: res.content[0]?.text ?? 'No se pudo hacer.' },
			{ status: 409, headers: sinCache }
		);
	}
	const resultado = res.structuredContent ?? {};
	if (clave) await datos.terminarConfirmacion(clave, true, resultado).catch(() => {});
	return json({ ok: true, resultado }, { headers: sinCache });
};
