import { error } from '@sveltejs/kit';
import * as v from 'valibot';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

const id = v.pipe(v.string(), v.maxLength(40));
const Ajuste = v.variant('tipo', [
	v.object({ tipo: v.literal('conexion'), id, valor: v.boolean() }),
	v.object({ tipo: v.literal('opcion'), id, clave: id, valor: v.boolean() }),
	v.object({ tipo: v.literal('ia'), id, valor: v.boolean() }),
	v.object({ tipo: v.literal('permisoIa'), id, valor: v.boolean() })
]);

/** Conectar o desconectar una integración, sus opciones y la IA (en la demo se simula). */
export const POST: RequestHandler = async ({ locals, request }) => {
	const b = v.safeParse(Ajuste, await request.json().catch(() => null));
	if (!b.success) error(400, 'Pedido inválido.');
	const r = await datosDe(locals).ajustarIntegracion(b.output);
	if (r === 'no-existe') error(404, 'No existe.');
	if (r === 'no-se-puede') error(409, 'Todavía no se puede conectar desde acá.');
	return new Response(null, { status: 204 });
};
