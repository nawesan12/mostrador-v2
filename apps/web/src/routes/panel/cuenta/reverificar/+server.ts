import { error, json } from '@sveltejs/kit';
import * as v from 'valibot';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** Con qué puede volver a verificarse (contraseña, app autenticadora, código por mail). */
export const GET: RequestHandler = async ({ locals }) => json(await datosDe(locals).metodosReverificar());

const Pedido = v.union([
	v.object({ accion: v.literal('codigo') }),
	v.object({ clave: v.pipe(v.string(), v.minLength(1), v.maxLength(200)) }),
	v.object({ codigo: v.pipe(v.string(), v.minLength(6), v.maxLength(12)) })
]);

/** Pide el código por mail, o verifica la contraseña o el código y devuelve el permiso (5 min). */
export const POST: RequestHandler = async ({ locals, request }) => {
	const r = v.safeParse(Pedido, await request.json().catch(() => null));
	if (!r.success) error(400, 'Revisá lo que escribiste.');
	const d = datosDe(locals);
	if ('accion' in r.output) {
		await d.pedirCodigoReverificar();
		return new Response(null, { status: 204 });
	}
	const res = await d.reverificar(r.output);
	if (!res.ok) error(400, res.mensaje);
	return json({ token: res.token });
};
