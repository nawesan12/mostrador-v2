import { goto } from '$app/navigation';
import { resolve } from '$app/paths';
import { TITULO_TOPE, type UiState } from '#lib/ui.svelte.ts';
import { planDe } from './planes.ts';

/**
 * El tope de plan (la API responde 402 con code tope_plan y el plan que lo habilita): se avisa en el
 * momento, con el botón para pasar a ese plan. Los topes los aplica el servidor; esto sólo los cuenta.
 */
export type ProblemaPlan = { message?: string; code?: string; plan?: string } | null;

export const esTope = (
	p: ProblemaPlan
): p is { message?: string; code: 'tope_plan'; plan?: string } => p?.code === 'tope_plan';

/** El link a Plan y facturación con el plan ya elegido. */
export const hrefPlan = (plan?: string) =>
	plan
		? `${resolve('/panel/plan-y-facturacion')}?plan=${encodeURIComponent(plan)}`
		: resolve('/panel/plan-y-facturacion');

/** Si el problema es un tope de plan, lo avisa con «Pasar a X» y devuelve true. */
export function avisarTope(ui: UiState, p: ProblemaPlan): boolean {
	if (!esTope(p)) return false;
	const sube = p.plan ? planDe(p.plan) : null;
	ui.toast({
		k: 'info',
		title: TITULO_TOPE,
		desc: p.message ?? 'Pasá a un plan más grande para seguir.',
		action: sube ? `Pasar a ${sube.nombre}` : 'Ver planes',
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- hrefPlan ya pasa por resolve
		onAction: () => goto(hrefPlan(p.plan))
	});
	return true;
}

/** Lee el cuerpo de una respuesta fallida y avisa si es un tope. */
export async function avisarTopeDe(ui: UiState, r: Response): Promise<boolean> {
	if (r.status !== 402) return false;
	const p = (await r
		.clone()
		.json()
		.catch(() => null)) as ProblemaPlan;
	return avisarTope(ui, p);
}

/**
 * Todo el panel: cualquier pedido a /panel/ que vuelve con un tope de plan (crear un cupón, conectar
 * Mercado Libre, un dominio, una venta en el local…) avisa con «Pasar a X», sin que cada pantalla lo
 * tenga que saber. Devuelve cómo deshacerlo.
 */
export function escucharTopes(ui: UiState): () => void {
	const original = window.fetch;
	const envuelto: typeof fetch = async (entrada, init) => {
		const r = await original(entrada, init);
		const url =
			typeof entrada === 'string' ? entrada : entrada instanceof URL ? entrada.href : entrada.url;
		if (r.status === 402 && new URL(url, location.href).pathname.startsWith('/panel/'))
			void avisarTopeDe(ui, r);
		return r;
	};
	window.fetch = envuelto;
	return () => {
		if (window.fetch === envuelto) window.fetch = original;
	};
}
