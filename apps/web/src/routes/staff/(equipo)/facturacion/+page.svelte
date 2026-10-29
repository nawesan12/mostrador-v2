<script lang="ts">
	import { resolve } from '$app/paths';
	import Seo from '#lib/components/Seo.svelte';
	import { ESTADOS_TIENDA, centavos, fechaCorta, nombrePlan } from '#lib/staff/estado.ts';

	/** Facturación: lo que cobra Tu Local Web a las tiendas. */
	let { data } = $props();
	const k = $derived(data.facturacion.kpis);
	const KPIS = $derived([
		['Ingresos del mes (MRR)', centavos(k.mrr)],
		['Cobrado este mes', centavos(k.cobrosMes)],
		['Suscripciones activas', String(k.activas)],
		['Con problemas', String(k.conProblemas)],
		['Créditos a favor', centavos(k.creditos)]
	]);
	// Primero las que tienen problemas.
	const subs = $derived(
		[...data.facturacion.suscripciones].sort(
			(a, b) => Number(b.estado === 'cobro_rechazado') - Number(a.estado === 'cobro_rechazado')
		)
	);
</script>

<Seo
	title="Facturación · Equipo"
	description="Facturación de Tu Local Web"
	path="/staff/facturacion"
	noindex
/>

<div class="cab"><h1>Facturación</h1></div>

<div class="kpis">
	{#each KPIS as [t, n] (t)}
		<div class="kpi">
			<div class="k-t">{t}</div>
			<div class="k-n">{n}</div>
		</div>
	{/each}
</div>

<section class="tarjeta">
	<div class="t-t">Suscripciones</div>
	{#each subs as s (s.tiendaId)}
		{@const e = ESTADOS_TIENDA[s.estado]}
		<a class="fila link" href={resolve('/staff/(equipo)/tiendas/[id]', { id: s.tiendaId })}>
			<span class="f-txt"
				><b>{s.nombre}</b><span class="f-s"
					>{nombrePlan(s.plan)} · {s.ciclo === 'anio' ? 'anual' : 'mensual'}{s.proximoCobro
						? ` · próximo cobro el ${fechaCorta(s.proximoCobro)}`
						: ''}{s.credito ? ` · crédito ${centavos(s.credito)}` : ''}</span
				></span
			>
			<span class="chip" style:background={e.bg} style:color={e.c}>{e.t}</span>
			<b class="monto">{centavos(s.monto)}</b>
		</a>
	{:else}
		<div class="vacio">Todavía no hay suscripciones pagas.</div>
	{/each}
</section>

<section class="tarjeta">
	<div class="t-t">Comprobantes emitidos</div>
	{#each data.facturacion.comprobantes as c (c.numero)}
		<div class="fila">
			<span class="f-txt"
				><b>{c.numero}</b><span class="f-s">{c.tienda} · {fechaCorta(c.fecha)}</span></span
			>
			<b class="monto">{centavos(c.monto)}</b>
		</div>
	{:else}
		<div class="vacio">Sin comprobantes este mes.</div>
	{/each}
</section>

<style>
	.link {
		color: inherit;
		text-decoration: none;
		flex-wrap: wrap;
	}
	.monto {
		font-variant-numeric: tabular-nums;
		white-space: nowrap;
	}
</style>
