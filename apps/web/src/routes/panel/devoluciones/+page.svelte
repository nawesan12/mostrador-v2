<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import Seo from '#lib/components/Seo.svelte';
	import { unsplash } from '#lib/landing/formato.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { pesos } from '#lib/panel/formato.ts';
	import { hace } from '#lib/panel/listas/clientes.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/** Devoluciones (design/PanelVentas.dc.html): aprobar u ofrecer un cambio. */
	let { data } = $props();
	const ui = getUi();

	async function resolver(id: string, accion: 'aprobar' | 'cambio' | 'deshacer') {
		await fetch(`/panel/devoluciones/${id}`, {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ accion })
		});
		await invalidateAll();
	}
	async function aprobar(d: (typeof data.devoluciones)[number]) {
		await resolver(d.id, 'aprobar');
		ui.toast({
			k: 'ok',
			title: 'Devolución aprobada',
			// Aprobar sólo cambia el estado: no hay etiqueta de transporte ni aviso automático.
			desc: `Avisale a ${d.cliente.split(' ')[0]} cómo te hace llegar el producto`,
			action: 'Deshacer',
			onAction: () => resolver(d.id, 'deshacer')
		});
	}
	async function cambio(d: (typeof data.devoluciones)[number]) {
		await resolver(d.id, 'cambio');
		ui.toast({
			k: 'ok',
			title: 'Quedó como cambio ofrecido',
			desc: `Escribile a ${d.cliente.split(' ')[0]} con las opciones que tenés`,
			action: 'Deshacer',
			onAction: () => resolver(d.id, 'deshacer')
		});
	}
	const muyBajo = $derived(data.tasa < 2);
</script>

<Seo
	title="Devoluciones · Panel"
	description="Devoluciones de tu tienda"
	path="/panel/devoluciones"
	noindex
/>

<div class="pagina">
	<div class="cab" use:anim={{ t: 'rise' }}>
		<div class="c-txt">
			<h1>Devoluciones</h1>
			<div class="c-s">Cambios, reintegros y arrepentimientos (10&nbsp;días por ley)</div>
		</div>
		<!-- Sin devoluciones, un "0 %" no dice nada (puede no haber pedidos todavía). -->
		{#if data.devoluciones.length}
			<span class="tasa" class:alta={!muyBajo}
				>{data.tasa.toLocaleString('es-AR')}% de tus pedidos · {muyBajo
					? 'muy bajo'
					: 'para mirar'}</span
			>
		{/if}
	</div>
	<div class="lista" use:anim={{ t: 'rise', delay: 60 }}>
		{#each data.devoluciones as d (d.id)}
			{@const pend = d.estado === 'Pendiente'}
			<div class="dev">
				{#if d.foto}<img src={unsplash(d.foto, 128, '&h=128')} alt="" />{/if}
				<span class="txt">
					<span class="fila"
						><span class="p">{d.producto}</span><span class="estado" class:pend>{d.estado}</span
						></span
					>
					<span class="s"
						>{d.cliente} · pedido #{d.pedido} · {hace(
							(Date.now() - Date.parse(d.creada)) / 86_400_000
						)}</span
					>
					<span class="motivo">“{d.motivo}”</span>
				</span>
				<span class="monto"
					><span class="v">{pesos(d.monto)}</span><span class="m">{d.resolucion}</span></span
				>
				{#if pend}
					<span class="acc">
						<button type="button" class="aprobar" onclick={() => aprobar(d)}>Aprobar</button>
						<button type="button" class="cambio" onclick={() => cambio(d)}>Ofrecer cambio</button>
					</span>
				{/if}
			</div>
		{:else}
			<div class="vacio">No hay devoluciones. Cuando alguien pida una, aparece acá.</div>
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
		gap: 14px;
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
	.tasa {
		font-size: 13.5px;
		font-weight: 600;
		padding: 9px 14px;
		border-radius: 999px;
		background: rgba(47, 143, 99, 0.12);
		color: #237550;
		white-space: nowrap;
	}
	.tasa.alta {
		background: rgba(227, 169, 59, 0.2);
		color: #8a5f0e;
	}
	.lista {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	.dev {
		padding: 16px;
		border-radius: 22px;
		background: var(--sf);
		box-shadow: var(--shd);
		display: flex;
		gap: 14px;
		flex-wrap: wrap;
		align-items: center;
	}
	.dev img {
		width: 64px;
		height: 64px;
		border-radius: 14px;
		object-fit: cover;
		flex: none;
	}
	.txt {
		flex: 1 1 220px;
		min-width: 0;
	}
	.fila {
		display: flex;
		align-items: center;
		gap: 8px;
		flex-wrap: wrap;
	}
	.p {
		font-size: 15px;
		font-weight: 600;
	}
	.estado {
		font-size: 12px;
		font-weight: 600;
		padding: 3px 9px;
		border-radius: 999px;
		background: rgba(47, 143, 99, 0.13);
		color: #237550;
	}
	.estado.pend {
		background: rgba(227, 169, 59, 0.2);
		color: #8a5f0e;
	}
	.s {
		display: block;
		font-size: 13px;
		color: var(--mu);
		margin-top: 3px;
	}
	.motivo {
		display: block;
		font-size: 13.5px;
		margin-top: 6px;
		line-height: 1.4;
	}
	.monto {
		text-align: right;
		min-width: 110px;
	}
	.v {
		display: block;
		font-size: 16px;
		font-weight: 600;
	}
	.m {
		display: block;
		font-size: 12px;
		color: var(--mu);
	}
	.acc {
		display: flex;
		gap: 6px;
	}
	.acc button {
		padding: 9px 14px;
		border-radius: 999px;
		font-size: 13px;
		font-weight: 600;
		white-space: nowrap;
	}
	.aprobar {
		border: 0;
		background: var(--tx);
		color: var(--sf);
	}
	.cambio {
		border: 1px solid var(--ln2);
		background: transparent;
		color: var(--tx);
	}
	.vacio {
		padding: 28px;
		text-align: center;
		color: var(--mu);
	}
</style>
