<script lang="ts">
	import { resolve } from '$app/paths';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { centavos, fechaCorta } from '#lib/staff/estado.ts';

	let { data } = $props();
	const k = $derived(data.inicio.kpis);
	const KPIS = $derived([
		['Tiendas activas', String(k.tiendasActivas)],
		['Nuevas en la semana', String(k.nuevasSemana)],
		['Pruebas que vencen en 3 días', String(k.pruebasPorVencer)],
		['Ingresos del mes (MRR)', centavos(k.mrr)],
		['Bajas del mes', String(k.bajasMes)],
		['Consultas sin responder', String(k.consultasSinResponder)],
		['Postulaciones de partners', String(k.postulacionesNuevas)]
	]);
	const ALERTA = {
		cobro_rechazado: I.phCreditCardFill,
		integracion: I.phPlugsConnectedFill,
		arrepentimiento: I.phArrowUUpLeftFill
	} as const;
	const ficha = (id: string) => resolve('/staff/(equipo)/tiendas/[id]', { id });
</script>

<Seo title="Inicio · Equipo" description="Backoffice del equipo" path="/staff" noindex />

<div class="cab"><h1>Hola, {data.sesion.miembro.nombre.split(' ')[0]}</h1></div>

<div class="kpis">
	{#each KPIS as [t, n] (t)}
		<div class="kpi">
			<div class="k-t">{t}</div>
			<div class="k-n">{n}</div>
		</div>
	{/each}
</div>

<div class="dos">
	<section class="tarjeta">
		<div class="t-t">Alertas</div>
		{#each data.inicio.alertas as a, i (i)}
			<div class="fila">
				<Icon icon={ALERTA[a.tipo]} size={18} style="color:#b0441d;flex:none" />
				<span class="f-txt"><b>{a.titulo}</b><span class="f-s">{a.detalle}</span></span>
				{#if a.tiendaId}<a class="borde" href={ficha(a.tiendaId)}>Ver</a>{:else if a.consulta}<a
						class="borde"
						href={resolve('/staff/(equipo)/consultas')}>Ver</a
					>{/if}
			</div>
		{:else}
			<div class="vacio">Sin alertas: todo en orden.</div>
		{/each}
	</section>

	<section class="tarjeta">
		<div class="t-t">Pruebas que vencen en 3 días</div>
		{#each data.inicio.pruebas as t (t.id)}
			<a class="fila link" href={ficha(t.id)}
				><span class="f-txt"
					><b>{t.nombre}</b><span class="f-s">Vence el {fechaCorta(t.fecha)}</span></span
				><Icon icon={I.phCaretRightBold} size={13} /></a
			>
		{:else}
			<div class="vacio">Ninguna vence en estos días.</div>
		{/each}
	</section>

	<section class="tarjeta">
		<div class="t-t">Bajas del mes</div>
		{#each data.inicio.bajas as t (t.id)}
			<a class="fila link" href={ficha(t.id)}
				><span class="f-txt"
					><b>{t.nombre}</b><span class="f-s"
						>{fechaCorta(t.fecha)}{t.motivo ? ` · ${t.motivo}` : ''}</span
					></span
				><Icon icon={I.phCaretRightBold} size={13} /></a
			>
		{:else}
			<div class="vacio">Ninguna baja este mes.</div>
		{/each}
	</section>
</div>

<style>
	.dos {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 320px), 1fr));
		gap: 14px;
		align-items: start;
	}
	.link {
		color: inherit;
		text-decoration: none;
	}
</style>
