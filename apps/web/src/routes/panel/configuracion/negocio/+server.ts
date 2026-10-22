import { error } from '@sveltejs/kit';
import * as v from 'valibot';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

const txt = (max: number) => v.pipe(v.string(), v.trim(), v.maxLength(max));
const Negocio = v.object({
	nombre: v.pipe(txt(80), v.minLength(1, 'Poné el nombre del negocio.')),
	razonSocial: txt(120),
	cuit: v.pipe(txt(13), v.regex(/^$|^\d{2}-?\d{8}-?\d$/, 'El CUIT tiene 11 números.')),
	condicion: txt(60),
	direccion: txt(160),
	whatsapp: txt(30),
	email: v.pipe(txt(120), v.regex(/^$|^[^\s@]+@[^\s@]+\.[^\s@]+$/, 'Revisá el email.')),
	horario: txt(80)
});

/** Datos del negocio. */
export const PUT: RequestHandler = async ({ locals, request }) => {
	const r = v.safeParse(Negocio, await request.json().catch(() => null));
	if (!r.success) error(400, r.issues[0].message);
	await datosDe(locals).guardarNegocio(r.output);
	return new Response(null, { status: 204 });
};
