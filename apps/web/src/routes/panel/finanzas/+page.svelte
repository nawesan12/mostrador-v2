<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { marca } from '#lib/marcas.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { pesos } from '#lib/panel/formato.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/** Finanzas (design/PanelOperacion.dc.html, sFin): todo lo que entra y sale. */
	let { data } = $props();
	const ui = getUi();
	type M = (typeof data.movimientos)[number];

	const millones = (n: number) =>
		Math.abs(n) >= 1e6
			? '$ ' + (n / 1e6).toLocaleString('es-AR', { maximumFractionDigits: 1 }) + ' M'
			: pesos(n);
	const k = $derived(data.kpis);
	const KPIS = $derived([
		['A liquidar', pesos(k.aLiquidar), k.aLiquidarDetalle, 'var(--tx)'],
		[`Gastos de ${k.mes}`, pesos(k.gastos), k.gastosDetalle, '#C24E25'],
		[
			'Ganancia estimada',
			k.ganancia === null ? 'Sin dato' : millones(k.ganancia),
			k.ganancia === null || k.margen === null
				? k.gananciaDetalle || 'Cargá el costo de tus productos'
				: `margen ${k.margen} % en 30 días`,
			k.ganancia === null ? 'var(--mu)' : '#2F8F63'
		]
	] as const);
	const maxEntra = $derived(Math.max(1, ...data.flujo.map((f) => f.entra)));
	const maxSale = $derived(Math.max(1, ...data.flujo.map((f) => f.sale)));
	function icono(m: M) {
		if (m.tipo === 'gasto')
			return {
				icon: I.phArrowUpRightBold,
				col: '#C24E25',
				bg: 'rgba(210,88,47,.12)',
				sigla: undefined
			};
		if (m.tipo === 'retiro')
			return { icon: I.phBankFill, col: '#6B5244', bg: '#F1E9DF', sigla: undefined };
		return marca(m.tipo);
	}
	let enviando = $state(false);
	async function transferir() {
		if (enviando) return;
		enviando = true;
		const r = await fetch('/panel/finanzas/retiro', { method: 'POST' });
		enviando = false;
		if (!r.ok) {
			const m = ((await r.json().catch(() => null)) as { message?: string } | null)?.message;
			return ui.toast({
				k: 'info',
				title: 'No hay nada para transferir',
				desc: m ?? 'Probá más tarde.'
			});
		}
		const { monto, cuenta } = (await r.json()) as { monto: number; cuenta: string };
		await invalidateAll();
		ui.toast({
			k: 'ok',
			title: 'Transferencia en camino',
			desc: `${pesos(monto)} a tu cuenta del ${cuenta} · llega hoy`
		});
	}
</script>

<Seo title="Finanzas · Panel" description="Finanzas de tu tienda" path="/panel/finanzas" noindex />

<div class="pagina">
	<div class="cab" use:anim={{ t: 'rise' }}>
		<div class="c-txt">
			<h1>Finanzas</h1>
			<div class="c-s">Todo lo que entra y sale</div>
		</div>
		<button type="button" class="transferir" disabled={enviando} onclick={transferir}
			><Icon icon={I.phArrowUpRightBold} size={15} />Transferir a mi banco</button
		>
	</div>

	<div class="kpis" use:anim={{ t: 'rise', delay: 90 }}>
		<div class="disponible">
			<div class="d-in">
				<div class="d-t">Disponible hoy</div>
				<div class="d-v">{pesos(data.disponible)}</div>
				<div class="d-d">
					<Icon icon={I.siMercadopago} size={14} />En Mercado Pago{data.cuenta
						? ` · ${data.cuenta}`
						: ''}
				</div>
			</div>
		</div>
		{#each KPIS as [t, v, d, c] (t)}
			<div class="kpi">
				<div class="k-t">{t}</div>
				<div class="k-v" style:color={c}>{v}</div>
				<div class="k-d">{d}</div>
			</div>
		{/each}
	</div>

	<div class="tarjeta flujo" use:anim={{ t: 'rise', delay: 135 }}>
		<div class="f-cab">
			<span>Flujo de caja · próximas 2 semanas</span>
			<span class="ley"
				><span><i style:background="#2F8F63"></i>Entra</span><span
					><i style:background="#E8875F"></i>Sale</span
				></span
			>
		</div>
		<div class="barras">
			{#each data.flujo as f, i (i)}
				<div class="col" title="Entra {pesos(f.entra)} · sale {pesos(f.sale)}">
					<div class="arriba"><span style:height="{(f.entra / maxEntra) * 100}%"></span></div>
					<div class="linea"></div>
					<div class="abajo"><span style:height="{(f.sale / maxSale) * 100}%"></span></div>
					<span class="dia">{f.dia}</span>
				</div>
			{/each}
		</div>
	</div>

	<div class="tarjeta movs" use:anim={{ t: 'rise', delay: 180 }}>
		<div class="mv-cab">
			<span>Movimientos</span>
			{#if data.conciliado}<span class="ok"
					><Icon icon={I.phCheckCircleFill} size={15} />Todo conciliado</span
				>{/if}
		</div>
		{#each data.movimientos as m (m.id)}
			{@const ic = icono(m)}
			<div class="mov">
				<span class="mv-ico" style:background={ic.bg}
					>{#if ic.icon}<Icon
							icon={ic.icon}
							size={17}
							style="color:{ic.col ?? 'inherit'}"
						/>{:else}<span class="sigla" style:color={ic.col}>{ic.sigla}</span>{/if}</span
				>
				<span class="mv-txt"
					><span class="mv-t">{m.titulo}</span><span class="mv-d">{m.detalle}</span></span
				>
				<b class:entra={m.monto > 0}>{m.monto > 0 ? '+' : '−'} {pesos(Math.abs(m.monto))}</b>
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
	.transferir {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 11px 20px;
		border-radius: 999px;
		border: 0;
		background: linear-gradient(180deg, #c24e25, #b0441d);
		color: #fff;
		font:
			600 14px 'Instrument Sans',
			sans-serif;
		white-space: nowrap;
		box-shadow:
			inset 0 1px 0 rgba(255, 255, 255, 0.3),
			0 10px 24px -10px rgba(210, 88, 47, 0.8);
	}
	.kpis {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
		gap: 12px;
	}
	.disponible {
		position: relative;
		overflow: hidden;
		padding: 20px;
		border-radius: 22px;
		background: #2a140c;
		color: #fbf7f2;
	}
	.d-in {
		position: relative;
	}
	.d-t {
		font-size: 13px;
		color: #d9c8ba;
	}
	.d-v {
		font-size: 36px;
		font-weight: 600;
		letter-spacing: -0.045em;
	}
	.d-d {
		margin-top: 6px;
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 12.5px;
		color: #f5b796;
	}
	.kpi {
		padding: 18px;
		border-radius: 22px;
		background: var(--sf);
		box-shadow: var(--shd);
	}
	.k-t {
		font-size: 13px;
		color: var(--mu);
	}
	.k-v {
		font-size: 28px;
		font-weight: 600;
		letter-spacing: -0.04em;
		margin-top: 4px;
	}
	.k-d {
		font-size: 12.5px;
		color: var(--mu);
	}
	.tarjeta {
		border-radius: 24px;
		background: var(--sf);
		box-shadow: var(--shd);
	}
	.flujo {
		padding: 20px;
	}
	.f-cab {
		display: flex;
		justify-content: space-between;
		align-items: center;
		flex-wrap: wrap;
		gap: 8px;
		font-size: 16px;
		font-weight: 600;
	}
	.ley {
		display: flex;
		gap: 12px;
		font-size: 12.5px;
		color: var(--mu);
		font-weight: 400;
	}
	.ley span {
		display: flex;
		align-items: center;
		gap: 5px;
	}
	.ley i {
		width: 9px;
		height: 9px;
		border-radius: 3px;
	}
	.barras {
		margin-top: 16px;
		height: 170px;
		display: flex;
		align-items: center;
		gap: 6px;
	}
	.col {
		flex: 1;
		height: 100%;
		display: flex;
		flex-direction: column;
		align-items: center;
		min-width: 0;
	}
	.arriba,
	.abajo {
		width: 100%;
		display: flex;
		justify-content: center;
	}
	.arriba {
		flex: 1;
		align-items: flex-end;
	}
	.abajo {
		flex: 0.6;
		align-items: flex-start;
	}
	.arriba span,
	.abajo span {
		width: 70%;
		max-width: 26px;
	}
	.arriba span {
		border-radius: 6px 6px 2px 2px;
		background: #2f8f63;
	}
	.abajo span {
		border-radius: 2px 2px 6px 6px;
		background: #e8875f;
	}
	.linea {
		height: 1px;
		width: 100%;
		background: var(--ln2);
	}
	.dia {
		font-size: 10.5px;
		color: var(--fa);
	}
	.movs {
		overflow: hidden;
	}
	.mv-cab {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 16px 18px;
		border-bottom: 1px solid var(--ln);
		font-size: 16px;
		font-weight: 600;
	}
	.ok {
		font-size: 12.5px;
		font-weight: 600;
		color: #237550;
		display: flex;
		align-items: center;
		gap: 5px;
	}
	.mov {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 11px 18px;
		border-bottom: 1px solid var(--ln);
	}
	.mov:last-child {
		border-bottom: 0;
	}
	.mv-ico {
		width: 34px;
		height: 34px;
		flex: none;
		border-radius: 10px;
		display: grid;
		place-items: center;
	}
	.sigla {
		font-size: 7px;
		font-weight: 800;
	}
	.mv-txt {
		flex: 1;
		min-width: 0;
	}
	.mv-t {
		display: block;
		font-size: 14px;
		font-weight: 600;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.mv-d {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
	}
	.mov b {
		white-space: nowrap;
		color: var(--tx);
	}
	.mov b.entra {
		color: #237550;
	}
	@media (max-width: 560px) {
		.barras {
			gap: 3px;
		}
	}
</style>
