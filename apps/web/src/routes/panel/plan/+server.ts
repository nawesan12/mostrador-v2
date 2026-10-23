import { error, json } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

const PLANES = ['gratis', 'emprende', 'pro', 'negocio'] as const;

/**
 * Cambiar de plan (y de ciclo). Con el abono por Mercado Pago, un plan pago devuelve { url }: la
 * suscripción, donde se carga la tarjeta. El plan cambia cuando Mercado Pago avisa.
 */
export const PUT: RequestHandler = async ({ locals, request }) => {
	const c = (await request.json().catch(() => null)) as { plan?: unknown; ciclo?: unknown } | null;
	const plan = PLANES.find((p) => p === c?.plan);
	if (!plan) error(400, 'Elegí un plan.');
	const ciclo = c?.ciclo === 'mes' || c?.ciclo === 'anio' ? c.ciclo : undefined;
	const r = await datosDe(locals).cambiarPlan(plan, ciclo);
	if (r?.url) {
		if (!/^https:\/\/(www\.)?mercadopago\.com(\.ar)?\//.test(r.url))
			error(502, 'Respuesta inesperada.');
		return json({ url: r.url });
	}
	return new Response(null, { status: 204 });
};
