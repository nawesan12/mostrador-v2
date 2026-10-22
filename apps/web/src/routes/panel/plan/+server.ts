import { error } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

const PLANES = ['gratis', 'emprende', 'pro', 'negocio'] as const;

/** Cambiar de plan (y de ciclo). */
export const PUT: RequestHandler = async ({ locals, request }) => {
	const c = (await request.json().catch(() => null)) as { plan?: unknown; ciclo?: unknown } | null;
	const plan = PLANES.find((p) => p === c?.plan);
	if (!plan) error(400, 'Elegí un plan.');
	const ciclo = c?.ciclo === 'mes' || c?.ciclo === 'anio' ? c.ciclo : undefined;
	await datosDe(locals).cambiarPlan(plan, ciclo);
	return new Response(null, { status: 204 });
};
