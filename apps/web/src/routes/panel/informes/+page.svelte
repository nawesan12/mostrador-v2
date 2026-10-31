<script lang="ts">
	import { deslizable } from '#lib/panel/deslizable.ts';
	import { resolve } from '$app/paths';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { pesos, pesosCortos, plural } from '#lib/panel/formato.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/** Informes (design/PanelCrecer.dc.html, sRep): 12 meses, KPIs y lo que notó Tincho. */
	let { data } = $props();
	const ui = getUi();

	const TODAS = [
		['ventas', 'Ventas', 'Ventas por mes'],
		['productos', 'Productos', 'Unidades vendidas'],
		['clientes', 'Clientes', 'Clientes nuevos'],
		['margen', 'Margen', 'Margen bruto %'],
		// Sólo desde que se miden (medición propia, sin cookies): sin visitas, no hay pestaña.
		['visitas', 'Visitas', 'Visitas a tu tienda online']
	] as const;
	const SERIES = $derived(TODAS.filter(([id]) => id !== 'visitas' || data.series.visitas));
	let serie = $state(0);
	const s = $derived(SERIES[serie] ?? SERIES[0]);
	const valores = $derived((data.series[s[0]] ?? []).map((v) => v ?? 0));
	const max = $derived(Math.max(1, ...valores));
	/** Etiqueta de cada barra: "1,2 M", "15,2 mil" o "850" (antes, todo en millones: "0"). */
	const etiqueta = (v: number) =>
		s[0] === 'ventas'
			? v >= 1e6
				? `${(v / 1e6).toLocaleString('es-AR', { maximumFractionDigits: 1 })}\u00a0M`
				: v >= 1e3
					? `${(v / 1e3).toLocaleString('es-AR', { maximumFractionDigits: 1 })}\u00a0mil`
					: v.toLocaleString('es-AR')
			: s[0] === 'margen'
				? `${v}%`
				: v.toLocaleString('es-AR');
	const k = $derived(data.kpis);
	const signo = (n: number | null, suf = '%') =>
		n === null
			? ''
			: `${n > 0 ? '+' : ''}${n.toLocaleString('es-AR', { maximumFractionDigits: 1 })}${suf}`;
	// Sin costos cargados el margen es un guion: no se muestra. «Comisiones ahorradas» era una cuenta
	// contra otra plataforma, no un dato de la tienda: tampoco.
	const KPIS = $derived(
		(
			[
				[
					'Ventas del mes',
					pesosCortos(k.ventas),
					k.ventasVs === null ? '' : `${signo(k.ventasVs)} vs. ${k.mesAnterior}`,
					(k.ventasVs ?? 0) >= 0 ? '#2F8F63' : '#B83A2E'
				],
				[
					'Ticket promedio',
					pesos(k.ticket),
					signo(k.ticketVs),
					(k.ticketVs ?? 0) >= 0 ? '#2F8F63' : '#B83A2E'
				],
				[
					'Margen bruto',
					k.margen === null ? '—' : `${k.margen}%`,
					k.margenVs === null
						? ''
						: `${signo(k.margenVs, '')} ${Math.abs(k.margenVs) === 1 ? 'punto' : 'puntos'}`,
					'#2F8F63'
				]
			] as const
		).filter(([t]) => t !== 'Margen bruto' || k.margen !== null)
	);
	const ICONOS = [I.phTrendUpBold, I.phCoffeeFill, I.phTagBold];
</script>

<Seo title="Informes · Panel" description="Informes de tu tienda" path="/panel/informes" noindex />

<div class="pagina">
	<div class="cab" use:anim={{ t: 'rise' }}>
		<div class="c-txt">
			<h1>Informes</h1>
			<div class="c-s">
				{data.historia
					? `Con ${plural(data.historia, 'mes', 'meses')} de ventas en Tu\u00a0Local\u00a0Web`
					: 'Cuando entren ventas, acá vas a ver cómo\u00a0crecen'}
			</div>
		</div>
		<a
			class="exportar"
			href="{resolve('/panel/informes/exportar')}?serie={s[0]}"
			download
			onclick={() =>
				ui.toast({ k: 'ok', title: 'Informe exportado', desc: `${s[2]}: últimos 12 meses` })}
			><Icon icon={I.phDownloadSimpleBold} size={14} />Exportar</a
		>
	</div>

	<div
		class="tabs"
		role="group"
		aria-label="Informe"
		use:anim={{ t: 'rise', delay: 45 }}
		use:deslizable
	>
		{#each SERIES as [id, t], i (id)}
			<button type="button" aria-pressed={serie === i} onclick={() => (serie = i)}>{t}</button>
		{/each}
	</div>

	<div class="kpis" use:anim={{ t: 'rise', delay: 90 }}>
		{#each KPIS as [t, v, d, c] (t)}
			<div class="kpi">
				<div class="k-t">{t}</div>
				<div class="k-v">{v}</div>
				<div class="k-d" style:color={c}>{d}</div>
			</div>
		{/each}
	</div>

	<div class="grafico" use:anim={{ t: 'rise', delay: 135 }}>
		<div class="g-cab"><span>{s[2]}</span><span class="g-s">últimos 12 meses</span></div>
		<div class="barras">
			{#key serie}
				{#each valores as v, i (i)}
					<div class="col">
						<span class="lbl" class:ultima={i === 11}>{etiqueta(v)}</span>
						<span
							class="barra"
							class:ultima={i === 11}
							style:height="{(v / max) * 82}%"
							style:animation-delay="{i * 35}ms"
						></span>
						<span class="mes">{data.meses[i]}</span>
					</div>
				{/each}
			{/key}
		</div>
	</div>

	<div class="insights" use:anim={{ t: 'rise', delay: 180 }}>
		{#each data.insights as t, i (t)}
			<div class="insight">
				<span class="in-ico"><Icon icon={ICONOS[i % ICONOS.length]} size={19} /></span>
				<span><span class="in-t">✦ Tincho notó</span><span class="in-d">{t}</span></span>
			</div>
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
	.exportar {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 16px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: var(--sf);
		color: var(--tx);
		font:
			600 13.5px 'Instrument Sans',
			sans-serif;
		box-shadow: var(--sh);
		white-space: nowrap;
		text-decoration: none;
	}
	.tabs {
		display: flex;
		padding: 3px;
		border-radius: 999px;
		background: var(--chip);
		align-self: flex-start;
		overflow-x: auto;
		scrollbar-width: none;
		max-width: 100%;
	}
	.tabs button {
		border: 0;
		font:
			600 13px 'Instrument Sans',
			sans-serif;
		padding: 8px 15px;
		border-radius: 999px;
		background: transparent;
		color: var(--tx);
		white-space: nowrap;
	}
	.tabs button[aria-pressed='true'] {
		background: var(--sf);
		box-shadow: var(--sh);
	}
	.kpis {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(170px, 1fr));
		gap: 12px;
	}
	.kpi {
		padding: 16px 18px;
		border-radius: 20px;
		background: var(--sf);
		box-shadow: var(--shd);
	}
	.k-t {
		font-size: 13px;
		color: var(--mu);
	}
	.k-v {
		font-size: 26px;
		font-weight: 600;
		letter-spacing: -0.04em;
		margin-top: 4px;
	}
	.k-d {
		font-size: 12.5px;
		font-weight: 600;
	}
	.grafico {
		padding: 20px;
		border-radius: 24px;
		background: var(--sf);
		box-shadow: var(--shd);
	}
	.g-cab {
		display: flex;
		justify-content: space-between;
		align-items: center;
		flex-wrap: wrap;
		gap: 8px;
		font-size: 16px;
		font-weight: 600;
	}
	.g-s {
		font-size: 12.5px;
		color: var(--mu);
		font-weight: 400;
	}
	.barras {
		margin-top: 16px;
		height: 220px;
		display: flex;
		align-items: flex-end;
		gap: 10px;
	}
	.col {
		flex: 1;
		height: 100%;
		display: flex;
		flex-direction: column;
		justify-content: flex-end;
		align-items: center;
		gap: 6px;
		min-width: 0;
	}
	.lbl {
		font-size: 11px;
		color: var(--mu);
		white-space: nowrap;
	}
	.barra {
		width: 100%;
		max-width: 44px;
		border-radius: 8px 8px 3px 3px;
		background: var(--sf3);
		transform-origin: bottom;
		animation: crecer 0.7s cubic-bezier(0.2, 0.7, 0.1, 1) both;
	}
	.barra.ultima {
		background: linear-gradient(180deg, #e8875f, #d2582f);
	}
	@keyframes crecer {
		from {
			transform: scaleY(0);
		}
	}
	.mes {
		font-size: 11px;
		color: var(--fa);
	}
	.insights {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
		gap: 12px;
	}
	.insight {
		display: flex;
		gap: 12px;
		padding: 16px;
		border-radius: 20px;
		background: var(--sf);
		box-shadow: var(--shd);
	}
	.in-ico {
		width: 38px;
		height: 38px;
		flex: none;
		border-radius: 12px;
		background: rgba(210, 88, 47, 0.12);
		color: #b0441d;
		display: grid;
		place-items: center;
	}
	.in-t {
		display: block;
		font-size: 12px;
		font-weight: 600;
		color: #b0441d;
	}
	.in-d {
		display: block;
		font-size: 14px;
		line-height: 1.45;
		margin-top: 2px;
	}
	@media (max-width: 700px) {
		.barras {
			gap: 4px;
		}
		.lbl:not(.ultima) {
			display: none;
		}
	}
	@media (max-width: 560px) {
		.kpis {
			grid-template-columns: 1fr 1fr;
		}
		/* Si quedan impares, la última va a lo ancho (sin una tarjeta sola en su renglón). */
		.kpi:last-child:nth-child(odd) {
			grid-column: 1 / -1;
		}
		.k-v {
			font-size: 21px;
		}
	}
	@media (prefers-reduced-motion: reduce) {
		.barra {
			animation: none;
		}
	}
</style>
