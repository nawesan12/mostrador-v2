import { error, json } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** Cancelar el plan (o aceptar la oferta para quedarse); DELETE lo reactiva. */
export const POST: RequestHandler = async ({ locals, request }) => {
	const c = (await request.json().catch(() => null)) as {
		motivo?: unknown;
		accion?: unknown;
	} | null;
	const motivo = Number(c?.motivo);
	if (!Number.isInteger(motivo) || motivo < 0 || motivo > 4) error(400, 'Falta el motivo.');
	if (c?.accion !== 'cancelar' && c?.accion !== 'oferta') error(400, 'Falta qué hacer.');
	return json(await datosDe(locals).cancelarPlan(motivo, c.accion));
};
export const DELETE: RequestHandler = async ({ locals }) => {
	await datosDe(locals).reactivarPlan();
	return new Response(null, { status: 204 });
};
