<script lang="ts" module>
	import { page } from '$app/state';
	import { planCon, planDeTienda, type Funcion } from './planes.ts';

	/** Los nombres de cada función, como los dice la API (dominio/planes › nombreFuncion). */
	export const NOMBRE_FUNCION: Record<Funcion, string> = {
		dominio: 'Dominio propio',
		arca: 'Factura electrónica con ARCA',
		correo: 'Envíos con Correo Argentino',
		cupones: 'Cupones, combos y carritos abandonados',
		roles: 'Roles y permisos del equipo',
		local: 'Venta en el local y cobro con QR',
		mercadolibre: 'Mercado Libre',
		newsletter: 'Newsletter y estadísticas',
		mayorista: 'Venta mayorista y listas de precio'
	};

	/** El plan que rige en la tienda del panel (en la prueba, Crece). Sin tienda (fuera del panel): null. */
	export function planDelPanel() {
		const t = page.data.tienda as
			{ plan: string; creadaEn: string; finPrueba?: string } | undefined;
		return t ? planDeTienda(t) : null;
	}

	/** La función no está en el plan que rige: la sección se ve igual y avisa antes de usarla. */
	export function fueraDelPlan(f: Funcion) {
		const p = planDelPanel();
		return !!p && !(p.funciones as readonly string[]).includes(f);
	}

	/** El problema de tope (como el de la API) para avisarlo al iniciar la acción. */
	export function topeDe(f: Funcion) {
		const p = planDelPanel();
		const sube = planCon(f);
		return {
			code: 'tope_plan' as const,
			plan: sube.id,
			title: `Es del plan ${sube.nombre}`,
			message: `${NOMBRE_FUNCION[f]}: desde el plan ${sube.nombre}. Tu plan es ${p?.nombre ?? 'Gratis'}.`
		};
	}
</script>

<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { hrefPlan } from './tope.ts';

	/**
	 * La franja de una sección cuya función no está en el plan que rige: dice desde qué plan está y
	 * lleva a cambiarlo. No esconde nada: la sección se sigue viendo (lo guardado queda).
	 */
	let { funcion, texto }: { funcion: Funcion; texto?: string } = $props();
	const sube = $derived(planCon(funcion));
	const fuera = $derived(fueraDelPlan(funcion));
</script>

{#if fuera}
	<div class="franja-plan" role="note">
		<Icon icon={I.phLockSimpleFill} size={20} style="color:var(--ac-tx);flex:none" />
		<p>
			<b>{NOMBRE_FUNCION[funcion]}: desde {sube.nombre}</b>
			<span
				>{texto ??
					`Tu plan es ${planDelPanel()?.nombre}. Lo que ya configuraste queda guardado.`}</span
			>
		</p>
		<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- hrefPlan ya pasa por resolve -->
		<a class="btn btn-primary" href={hrefPlan(sube.id)}>Pasar a {sube.nombre}</a>
	</div>
{/if}

<style>
	.franja-plan {
		display: flex;
		align-items: center;
		gap: 12px;
		flex-wrap: wrap;
		padding: 14px 16px;
		border-radius: 18px;
		background: linear-gradient(135deg, rgba(210, 88, 47, 0.12), rgba(210, 88, 47, 0.04));
		border: 1px solid rgba(210, 88, 47, 0.25);
	}
	p {
		flex: 1;
		min-width: 200px;
		margin: 0;
		display: flex;
		flex-direction: column;
		gap: 2px;
		font-size: 14px;
		color: var(--tx);
		text-wrap: pretty;
	}
	span {
		color: var(--mu);
		font-size: 13px;
	}
	a {
		min-height: 44px;
		white-space: nowrap;
	}
	@media (max-width: 480px) {
		a {
			width: 100%;
			justify-content: center;
		}
	}
</style>
