<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { pesos, plural } from '#lib/panel/formato.ts';
	import { aumentar, type Redondeo } from '#lib/panel/precios.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/** Listas de precio (design/PanelCatalogo.dc.html, sList): las listas y el aumento por inflación. */
	let { data } = $props();
	const ui = getUi();

	const ESTILO = {
		minorista: [I.phStorefrontFill, '#FCEDE5', '#C24E25'],
		mayorista: [I.phPackageFill, '#EFE9F7', '#6A4FA0'],
		ml: [I.phHandshakeFill, '#FFF6C2', '#6B5B00'],
		revendedoras: [I.phUsersFill, 'rgba(47,143,99,.13)', '#2F8F63']
	} as const;
	const regla = (a: number) => (a === 0 ? 'Base' : `${a > 0 ? '+' : '−'}${Math.abs(a)}%`);

	let lista = $state(0);
	let pct = $state(8);
	let redondeo = $state<Redondeo>('90');
	const RND: [Redondeo, string][] = [
		['no', 'Sin redondeo'],
		['90', 'Terminado en 90'],
		['500', 'Múltiplo de 500']
	];
	const elegida = $derived(data.listas[lista]);
	const muestra = $derived(
		data.muestra.map((p) => {
			const hoy = p.precio * (1 + elegida.ajuste / 100);
			return { ...p, hoy, nuevo: aumentar(hoy, pct, redondeo) };
		})
	);
	let enviando = $state(false);

	async function aplicar(programar: boolean) {
		if (enviando) return;
		enviando = true;
		const r = await fetch('/panel/listas-de-precio/aumento', {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ pct, redondeo, programar })
		});
		enviando = false;
		if (!r.ok)
			return ui.toast({
				k: 'err',
				title: 'No se actualizaron los precios',
				desc: 'Probá de nuevo.'
			});
		const { productos, para } = (await r.json()) as { productos: number; para: string | null };
		await invalidateAll();
		if (para)
			return ui.toast({
				k: 'info',
				title: 'Aumento programado',
				desc: `${cuandoEs(para)} a las 00:00 · te avisamos antes`
			});
		ui.toast({
			k: 'ok',
			title: `Precios actualizados +${pct}%`,
			desc: `${plural(productos, 'producto', 'productos')} en todos tus canales`,
			action: 'Deshacer',
			onAction: async () => {
				await fetch('/panel/listas-de-precio/aumento', { method: 'DELETE' });
				await invalidateAll();
			}
		});
	}
	async function cancelar() {
		await fetch('/panel/listas-de-precio/programado', { method: 'DELETE' });
		await invalidateAll();
		ui.toast({ k: 'ok', title: 'Aumento cancelado', desc: 'Los precios quedan como están' });
	}
	const cuandoEs = (iso: string) =>
		new Date(iso)
			.toLocaleDateString('es-AR', {
				weekday: 'long',
				day: 'numeric',
				timeZone: 'America/Argentina/Buenos_Aires'
			})
			.replace(/^\w/, (c) => c.toUpperCase());
</script>

<Seo
	title="Listas de precio · Panel"
	description="Listas de precio"
	path="/panel/listas-de-precio"
	noindex
/>

<div class="pagina">
	<div class="cab" use:anim={{ t: 'rise' }}>
		<div class="c-txt">
			<h1>Listas de precio</h1>
			<div class="c-s">Un precio base y reglas para cada canal y tipo de cliente</div>
		</div>
	</div>

	<div class="listas" use:anim={{ t: 'rise', delay: 60 }} role="group" aria-label="Listas">
		{#each data.listas as l, k (l.id)}
			{@const [icono, bg, fg] = ESTILO[l.id]}
			<button type="button" class="lista" aria-pressed={lista === k} onclick={() => (lista = k)}>
				<div class="l-cab">
					<span class="l-ico" style:background={bg} style:color={fg}
						><Icon icon={icono} size={19} /></span
					>
					<span class="l-txt"
						><span class="l-t">{l.nombre}</span><span class="l-q">{l.quien}</span></span
					>
				</div>
				<div class="l-r">{regla(l.ajuste)}</div>
				<div class="l-d">{l.detalle}</div>
			</button>
		{/each}
	</div>

	<div class="inflacion" use:anim={{ t: 'rise', delay: 120 }}>
		<div class="i-cuerpo">
			<div class="i-izq">
				<div class="i-t">Actualizar precios por inflación</div>
				<div class="i-d">
					Subí todo de una, con redondeo lindo ($&nbsp;8.990 en vez de $&nbsp;9.012).
				</div>
				<div class="i-slider">
					<input
						type="range"
						min="0"
						max="25"
						step="1"
						bind:value={pct}
						aria-label="Porcentaje de aumento"
					/>
					<span class="i-pct">+{pct}%</span>
				</div>
				<div class="i-rnd" role="group" aria-label="Redondeo">
					{#each RND as [v, t] (v)}
						<button type="button" aria-pressed={redondeo === v} onclick={() => (redondeo = v)}
							>{t}</button
						>
					{/each}
				</div>
			</div>
			<div class="tabla">
				<div class="fila cab-t">
					<span>Producto</span><span>Hoy</span><span>Nuevo · {elegida.nombre}</span>
				</div>
				{#each muestra as p (p.id)}
					<div class="fila">
						<span class="t-n">{p.nombre}</span><span class="t-a">{pesos(p.hoy)}</span><b
							>{pesos(p.nuevo)}</b
						>
					</div>
				{/each}
			</div>
		</div>
		{#if data.programado}
			<div class="prog">
				<Icon icon={I.phCalendarCheckFill} size={16} />
				<span
					>Programado: +{data.programado.pct}% el {cuandoEs(data.programado.para).toLowerCase()} a las
					00:00</span
				>
				<button type="button" onclick={cancelar}>Cancelar</button>
			</div>
		{/if}
		<div class="i-acc">
			<button type="button" class="prog-b" disabled={!pct || enviando} onclick={() => aplicar(true)}
				>Programar para el lunes</button
			>
			<button
				type="button"
				class="aplicar"
				disabled={!pct || enviando}
				onclick={() => aplicar(false)}
				>Aplicar a {plural(data.productos, 'producto', 'productos')}</button
			>
		</div>
	</div>
</div>

<style>
	.pagina {
		display: flex;
		flex-direction: column;
		gap: 18px;
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
	.listas {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
		gap: 12px;
	}
	.lista {
		text-align: left;
		padding: 16px;
		border-radius: 22px;
		border: 1.5px solid transparent;
		background: var(--sf);
		box-shadow: var(--shd);
		color: var(--tx);
	}
	.lista[aria-pressed='true'] {
		border-color: #d2582f;
	}
	.l-cab {
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.l-ico {
		width: 38px;
		height: 38px;
		border-radius: 12px;
		display: grid;
		place-items: center;
		flex: none;
	}
	.l-txt {
		flex: 1;
	}
	.l-t {
		display: block;
		font-size: 15px;
		font-weight: 600;
	}
	.l-q {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
	}
	.l-r {
		margin-top: 12px;
		font-size: 26px;
		font-weight: 600;
		letter-spacing: -0.03em;
	}
	.l-d {
		font-size: 12.5px;
		color: var(--mu);
	}
	.inflacion {
		padding: 20px;
		border-radius: 24px;
		background: #2a140c;
		color: #fbf7f2;
		position: relative;
		overflow: hidden;
	}
	.i-cuerpo {
		position: relative;
		display: flex;
		gap: 20px;
		flex-wrap: wrap;
		align-items: center;
	}
	.i-izq {
		flex: 1 1 280px;
	}
	.i-t {
		font-size: 17px;
		font-weight: 600;
	}
	.i-d {
		font-size: 13.5px;
		color: #d9c8ba;
		margin-top: 2px;
	}
	.i-slider {
		margin-top: 14px;
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.i-slider input {
		flex: 1;
		accent-color: #e8875f;
	}
	.i-pct {
		font-size: 28px;
		font-weight: 600;
		min-width: 64px;
	}
	.i-rnd {
		margin-top: 8px;
		display: flex;
		gap: 6px;
		flex-wrap: wrap;
	}
	.i-rnd button {
		padding: 7px 12px;
		border-radius: 999px;
		border: 1px solid rgba(251, 247, 242, 0.25);
		background: transparent;
		color: #fbf7f2;
		font:
			600 12.5px 'Instrument Sans',
			sans-serif;
	}
	.i-rnd button[aria-pressed='true'] {
		background: #fbf7f2;
		color: #2a140c;
	}
	.tabla {
		flex: 1 1 300px;
		border-radius: 18px;
		background: rgba(251, 247, 242, 0.06);
		border: 1px solid rgba(251, 247, 242, 0.12);
		overflow: hidden;
	}
	.fila {
		display: grid;
		grid-template-columns: minmax(0, 1.4fr) 1fr 1fr;
		gap: 8px;
		padding: 9px 14px;
		border-top: 1px solid rgba(251, 247, 242, 0.08);
		font-size: 13.5px;
	}
	.fila.cab-t {
		padding: 10px 14px;
		border-top: 0;
		font-size: 12px;
		color: #c9b4a4;
		font-weight: 600;
	}
	.t-n {
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.t-a {
		color: #c9b4a4;
	}
	.fila b {
		color: #f5b796;
	}
	.prog {
		position: relative;
		margin-top: 14px;
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 14px;
		border-radius: 14px;
		background: rgba(251, 247, 242, 0.08);
		font-size: 13.5px;
	}
	.prog span {
		flex: 1;
	}
	.prog button {
		border: 0;
		background: transparent;
		color: #f5b796;
		font:
			600 13px 'Instrument Sans',
			sans-serif;
	}
	.i-acc {
		position: relative;
		margin-top: 16px;
		display: flex;
		gap: 8px;
		justify-content: flex-end;
		flex-wrap: wrap;
	}
	.i-acc button {
		padding: 11px 16px;
		border-radius: 999px;
		font:
			600 13.5px 'Instrument Sans',
			sans-serif;
	}
	.i-acc button:disabled {
		opacity: 0.5;
	}
	.prog-b {
		border: 1px solid rgba(251, 247, 242, 0.3);
		background: transparent;
		color: #fff;
	}
	.aplicar {
		border: 0;
		background: #b0441d;
		color: #fff;
	}
</style>
