import { error, json } from '@sveltejs/kit';
import * as v from 'valibot';
import { datosDe } from '#lib/server/datos/index.ts';
import { llamar } from '#lib/server/mcp/herramientas.ts';
import type { RequestHandler } from './$types';

const Pedido = v.object({
	herramienta: v.picklist([
		'cambiar_estado_pedido',
		'ajustar_producto',
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
	const res = await llamar(
		{
			datos: datosDe(locals),
			permisos: ['lectura', 'acciones'],
			sitio: url.origin,
			permisosPanel: { escribir: true, plata: true }
		},
		r.output.herramienta,
		r.output.argumentos
	);
	if (res.isError)
		return json(
			{ ok: false, error: res.content[0]?.text ?? 'No se pudo hacer.' },
			{ status: 409, headers: { 'cache-control': 'no-store' } }
		);
	return json(
		{ ok: true, resultado: res.structuredContent ?? {} },
		{ headers: { 'cache-control': 'no-store' } }
	);
};
