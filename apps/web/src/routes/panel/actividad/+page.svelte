<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { unsplash } from '#lib/landing/formato.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/** Actividad (design/PanelCrecer.dc.html, sAct): quién hizo qué y desde dónde, con deshacer. */
	let { data } = $props();
	const ui = getUi();

	const FILTROS = [
		['todo', 'Todo'],
		['equipo', 'Equipo'],
		['hornero', 'Hornero'],
		['sistema', 'Sistema']
	] as const;
	let filtro = $state<(typeof FILTROS)[number][0]>('todo');
	const items = $derived(data.items.filter((a) => filtro === 'todo' || a.tipo === filtro));

	const MIN = 60_000;
	function cuando(iso: string) {
		const t = Date.parse(iso);
		const min = Math.round((Date.now() - t) / MIN);
		if (min < 1) return 'recién';
		if (min < 60) return `hace ${min} min`;
		if (min < 6 * 60) return `hace ${Math.round(min / 60)} h`;
		const hora = new Date(t).toLocaleTimeString('es-AR', {
			hour: '2-digit',
			minute: '2-digit',
			timeZone: 'America/Argentina/Buenos_Aires'
		});
		const dia = (x: number) => new Date(x - 3 * 3_600_000).toISOString().slice(0, 10);
		if (dia(t) === dia(Date.now())) return `hoy ${hora}`;
		if (dia(t) === dia(Date.now() - 86_400_000)) return `ayer ${hora}`;
		return new Date(t).toLocaleDateString('es-AR', {
			day: 'numeric',
			month: 'short',
			timeZone: 'America/Argentina/Buenos_Aires'
		});
	}

	async function deshacer(id: string) {
		const r = await fetch(`/panel/actividad/${encodeURIComponent(id)}/deshacer`, {
			method: 'POST'
		});
		if (!r.ok) return ui.toast({ k: 'err', title: 'No se pudo deshacer', desc: 'Probá de nuevo.' });
		await invalidateAll();
		ui.toast({ k: 'ok', title: 'Cambio deshecho', desc: 'Volvimos todo a como estaba' });
	}
	const ICONO = {
		hornero: [I.phSparkleFill, 'rgba(210,88,47,.12)', '#C24E25'],
		sistema: [I.phGearSixFill, 'var(--chip)', 'var(--mu)'],
		equipo: [I.phUserFill, 'rgba(210,88,47,.12)', '#C24E25']
	} as const;
</script>

<Seo
	title="Actividad · Panel"
	description="Actividad de la cuenta"
	path="/panel/actividad"
	noindex
/>

<div class="pagina">
	<div class="cab" use:anim={{ t: 'rise' }}>
		<div class="c-txt">
			<h1>Actividad</h1>
			<div class="c-s">Todo lo que pasó en tu cuenta, quién lo hizo y desde dónde</div>
		</div>
		<div class="tabs" role="group" aria-label="Ver">
			{#each FILTROS as [id, t] (id)}
				<button type="button" aria-pressed={filtro === id} onclick={() => (filtro = id)}>{t}</button
				>
			{/each}
		</div>
	</div>

	<div class="lista" use:anim={{ t: 'rise', delay: 60 }}>
		{#each items as a (a.id)}
			{@const [icono, bg, fg] = ICONO[a.tipo]}
			<div class="evento" class:deshecho={a.deshecho}>
				{#if a.foto}
					<img src={unsplash(a.foto, 72, '&h=72&crop=faces')} alt="" />
				{:else if a.quien === 'Vos'}
					<span class="av vos">V</span>
				{:else}
					<span class="av" style:background={bg} style:color={fg}
						><Icon icon={icono} size={17} /></span
					>
				{/if}
				<span class="txt">
					<span class="t"><b>{a.quien}</b> {a.texto}</span>
					<span class="s">{cuando(a.cuando)} · {a.donde}</span>
				</span>
				{#if a.deshacer}
					<button type="button" class="deshacer" onclick={() => deshacer(a.id)}>Deshacer</button>
				{/if}
			</div>
		{:else}
			<div class="vacio">Nada por acá todavía.</div>
		{/each}
	</div>
</div>

<style>
	.pagina {
		display: flex;
		flex-direction: column;
		gap: 18px;
	}
	.cab {
		display: flex;
		align-items: end;
		gap: 12px;
		flex-wrap: wrap;
	}
	.c-txt {
		flex: 1;
		min-width: 220px;
	}
	h1 {
		margin: 0;
		font-size: 30px;
		letter-spacing: -0.035em;
		font-weight: 600;
	}
	.c-s {
		margin-top: 2px;
		font-size: 15px;
		color: var(--mu);
	}
	.tabs {
		display: flex;
		padding: 3px;
		border-radius: 999px;
		background: var(--chip);
	}
	.tabs button {
		border: 0;
		font:
			600 13px 'Instrument Sans',
			sans-serif;
		padding: 8px 13px;
		border-radius: 999px;
		background: transparent;
		color: var(--tx);
		white-space: nowrap;
	}
	.tabs button[aria-pressed='true'] {
		background: var(--sf);
		box-shadow: var(--sh);
	}
	.lista {
		padding: 8px 20px;
		border-radius: 24px;
		background: var(--sf);
		box-shadow: var(--shd);
	}
	.evento {
		display: flex;
		gap: 14px;
		padding: 14px 0;
		border-bottom: 1px solid var(--ln);
	}
	.evento:last-child {
		border-bottom: 0;
	}
	.evento img,
	.av {
		width: 36px;
		height: 36px;
		border-radius: 50%;
		flex: none;
	}
	.evento img {
		object-fit: cover;
	}
	.av {
		display: grid;
		place-items: center;
	}
	.av.vos {
		background: #2a140c;
		color: #fbf7f2;
		font-size: 14px;
		font-weight: 700;
	}
	.txt {
		flex: 1;
		min-width: 0;
	}
	.t {
		display: block;
		font-size: 14px;
		line-height: 1.4;
	}
	.deshecho .t {
		color: var(--mu);
	}
	.s {
		display: block;
		font-size: 12.5px;
		color: var(--fa);
		margin-top: 2px;
	}
	.deshacer {
		align-self: center;
		padding: 7px 12px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: transparent;
		color: var(--tx);
		font:
			600 12.5px 'Instrument Sans',
			sans-serif;
	}
	.vacio {
		padding: 24px 0;
		text-align: center;
		color: var(--mu);
	}
</style>
