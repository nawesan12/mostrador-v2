<script lang="ts">
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { ESTADOS_CONSULTA, fechaCorta, habilesHasta } from '#lib/staff/estado.ts';

	/** Consultas: lo que llega por Contacto y por el botón de arrepentimiento (con su plazo legal). */
	let { data } = $props();
	const FILTROS = [
		['todas', 'Todas'],
		['nueva', 'Nuevas'],
		['en_curso', 'En curso'],
		['resuelta', 'Resueltas']
	] as const;
	const filtro = $derived(page.url.searchParams.get('estado') ?? 'todas');
	const lista = $derived(
		data.bandeja.consultas.filter((c) => filtro === 'todas' || c.estado === filtro)
	);
	const nombreDe = (id: string | null) =>
		data.bandeja.miembros.find((m) => m.id === id)?.nombre ?? 'Sin asignar';
</script>

<Seo title="Consultas · Equipo" description="Bandeja del equipo" path="/staff/consultas" noindex />

<div class="cab"><h1>Consultas</h1></div>

<div class="filtros" role="group" aria-label="Estado">
	{#each FILTROS as [k, t] (k)}
		<a
			href="{resolve('/staff/(equipo)/consultas')}?estado={k}"
			aria-current={filtro === k ? 'true' : undefined}
			>{t}<span class="n"
				>{k === 'todas'
					? data.bandeja.consultas.length
					: data.bandeja.consultas.filter((c) => c.estado === k).length}</span
			></a
		>
	{/each}
</div>

<section class="tarjeta">
	{#each lista as c (c.tipo + c.id)}
		{@const e = ESTADOS_CONSULTA[c.estado]}
		{@const dias =
			c.tipo === 'arrepentimiento' && c.plazo ? habilesHasta(c.plazo, Date.now()) : null}
		<a
			class="fila link"
			href={resolve('/staff/(equipo)/consultas/[tipo]/[id]', { tipo: c.tipo, id: c.id })}
		>
			<Icon
				icon={c.tipo === 'arrepentimiento' ? I.phArrowUUpLeftFill : I.phEnvelopeSimpleFill}
				size={18}
				style="color:#b0441d;flex:none"
			/>
			<span class="f-txt"
				><b>{c.tipo === 'arrepentimiento' ? `${c.codigo} · ` : ''}{c.asunto || c.nombre}</b><span
					class="f-s">{c.nombre} · {fechaCorta(c.creadaEn)} · {nombreDe(c.asignadoA)}</span
				></span
			>
			{#if dias !== null && c.estado !== 'resuelta'}<span
					class="chip"
					style:background={dias <= 2 ? 'rgba(184,58,46,.1)' : 'var(--chip)'}
					style:color={dias <= 2 ? '#B83A2E' : 'var(--mu)'}
					>{dias < 0
						? 'Plazo vencido'
						: dias === 0
							? 'Vence hoy'
							: `${dias} ${dias === 1 ? 'día hábil' : 'días hábiles'}`}</span
				>{/if}
			<span class="chip" style:background={e.bg} style:color={e.c}>{e.t}</span>
		</a>
	{:else}
		<div class="vacio">No hay consultas acá.</div>
	{/each}
</section>

<style>
	.filtros {
		display: flex;
		gap: 6px;
		overflow-x: auto;
	}
	.filtros a {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 8px 14px;
		border-radius: 999px;
		background: var(--chip);
		color: var(--tx);
		font-size: 13.5px;
		font-weight: 600;
		text-decoration: none;
		white-space: nowrap;
	}
	.filtros a[aria-current='true'] {
		background: var(--tx);
		color: var(--bg);
	}
	.n {
		opacity: 0.6;
		font-variant-numeric: tabular-nums;
	}
	.link {
		color: inherit;
		text-decoration: none;
		flex-wrap: wrap;
	}
</style>
