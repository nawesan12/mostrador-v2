<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import Seo from '#lib/components/Seo.svelte';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { unsplash } from '#lib/landing/formato.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { hace } from '#lib/panel/listas/clientes.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/** Reseñas y preguntas (design/PanelVentas.dc.html): responder preguntas de Mercado Libre y ver las reseñas. */
	let { data } = $props();
	const ui = getUi();
	let tab = $state<'q' | 'r'>('q');
	const pendientes = $derived(data.preguntas.filter((p) => !p.respuesta).length);
	let editando = $state<string | null>(null);
	let texto = $state('');

	async function publicar(id: string, respuesta: string, producto: string) {
		const r = await fetch(`/panel/resenas-y-preguntas/${id}`, {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ respuesta })
		});
		if (!r.ok) return ui.toast({ k: 'err', title: 'No se publicó', desc: 'Revisá la respuesta.' });
		editando = null;
		await invalidateAll();
		ui.toast({ k: 'ml', title: 'Respuesta publicada', desc: `${producto} · Mercado Libre` });
	}
	const dias = (iso: string) => (Date.now() - Date.parse(iso)) / 86_400_000;
	const cuando = (iso: string) => {
		const h = (Date.now() - Date.parse(iso)) / 3_600_000;
		return h < 24 ? `hace ${Math.max(1, Math.round(h))} h` : hace(dias(iso));
	};
	const maxEstrellas = $derived(Math.max(1, ...data.porEstrellas));
</script>

<Seo
	title="Reseñas y preguntas · Panel"
	description="Reseñas de tu tienda"
	path="/panel/resenas-y-preguntas"
	noindex
/>

<div class="pagina">
	<div class="cab" use:anim={{ t: 'rise' }}>
		<div class="c-txt">
			<h1>Reseñas y preguntas</h1>
			<div class="c-s">De tu tienda, Google y Mercado Libre</div>
		</div>
		<div class="tabs" role="group" aria-label="Ver">
			<button type="button" aria-pressed={tab === 'q'} onclick={() => (tab = 'q')}
				>Preguntas{#if pendientes}{` · ${pendientes}`}{/if}</button
			>
			<button type="button" aria-pressed={tab === 'r'} onclick={() => (tab = 'r')}>Reseñas</button>
		</div>
	</div>

	{#if tab === 'q'}
		<div class="preguntas" use:anim={{ t: 'rise', delay: 60 }}>
			{#each data.preguntas as q (q.id)}
				<div class="tarjeta">
					<div class="q-cab">
						{#if q.foto}<img src={unsplash(q.foto, 96, '&h=96')} alt="" />{/if}
						<span class="q-txt"
							><span class="q-s"><span class="ml">ML</span>{q.producto} · {cuando(q.creada)}</span
							><span class="q-q">{q.pregunta}</span></span
						>
					</div>
					{#if q.respuesta}
						<div class="publicada">
							<Icon icon={I.phCheckCircleFill} size={15} />Publicada en Mercado Libre
						</div>
						<div class="r-txt">{q.respuesta}</div>
					{:else}
						<div class="sugerida">
							<div class="s-t">✦ Respuesta de Hornero</div>
							{#if editando === q.id}
								<textarea bind:value={texto} aria-label="Respuesta" maxlength="2000"></textarea>
							{:else}
								<div class="s-txt">{q.sugerida}</div>
							{/if}
							<div class="s-acc">
								<button
									type="button"
									class="pub"
									onclick={() => publicar(q.id, editando === q.id ? texto : q.sugerida, q.producto)}
									>Publicar</button
								>
								{#if editando !== q.id}
									<button
										type="button"
										class="ed"
										onclick={() => {
											editando = q.id;
											texto = q.sugerida;
										}}>Editar</button
									>
								{/if}
							</div>
						</div>
					{/if}
				</div>
			{:else}
				<div class="vacio">No hay preguntas sin responder.</div>
			{/each}
		</div>
	{:else}
		<div class="resenas" use:anim={{ t: 'rise', delay: 60 }}>
			<div class="tarjeta resumen">
				<div class="r-prom">
					<span class="n"
						>{(data.promedio ?? 0).toLocaleString('es-AR', { minimumFractionDigits: 1 })}</span
					><span class="estrellas">★★★★★</span>
				</div>
				<div class="r-s">{data.total} reseñas · {data.recomendaria}% recomendaría</div>
				<div class="barras">
					{#each data.porEstrellas as c, i (i)}
						<div class="barra">
							<span class="b-n">{5 - i}</span>
							<div class="b-f"><div style:width="{(c / maxEstrellas) * 100}%"></div></div>
							<span class="b-c">{c}</span>
						</div>
					{/each}
				</div>
			</div>
			<div class="lista">
				{#each data.resenas.slice(0, 20) as r, i (i)}
					<div class="tarjeta resena">
						<div class="re-cab">
							<span class="estrellas" aria-label="{r.estrellas} estrellas"
								>{'★'.repeat(r.estrellas)}{'☆'.repeat(5 - r.estrellas)}</span
							><span class="re-q">{r.autor} · {r.origen}</span><span class="re-h"
								>{hace(dias(r.creada))}</span
							>
						</div>
						<div class="re-t">{r.texto}</div>
					</div>
				{/each}
			</div>
		</div>
	{/if}
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
	.tabs {
		display: flex;
		padding: 3px;
		border-radius: 999px;
		background: var(--chip);
	}
	.tabs button {
		border: 0;
		font-size: 13px;
		font-weight: 600;
		padding: 8px 14px;
		border-radius: 999px;
		background: transparent;
		color: var(--tx);
		white-space: nowrap;
	}
	.tabs button[aria-pressed='true'] {
		background: var(--sf);
		box-shadow: var(--sh);
	}
	.preguntas,
	.lista {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	.tarjeta {
		padding: 16px;
		border-radius: 22px;
		background: var(--sf);
		box-shadow: var(--shd);
	}
	.q-cab {
		display: flex;
		gap: 12px;
		align-items: flex-start;
	}
	.q-cab img {
		width: 48px;
		height: 48px;
		border-radius: 12px;
		object-fit: cover;
		flex: none;
	}
	.q-txt {
		flex: 1;
		min-width: 0;
	}
	.q-s {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 12.5px;
		color: var(--mu);
	}
	.ml {
		font-size: 8px;
		font-weight: 800;
		padding: 2px 4px;
		border-radius: 4px;
		background: #ffe600;
		color: #2d3277;
	}
	.q-q {
		display: block;
		font-size: 15px;
		font-weight: 600;
		margin-top: 3px;
	}
	.sugerida {
		margin-top: 12px;
		padding: 12px;
		border-radius: 14px;
		background: rgba(210, 88, 47, 0.08);
		border: 1px solid rgba(210, 88, 47, 0.2);
	}
	.s-t {
		font-size: 12px;
		font-weight: 600;
		color: #c24e25;
		margin-bottom: 4px;
	}
	.s-txt {
		font-size: 14px;
		line-height: 1.45;
	}
	.sugerida textarea {
		width: 100%;
		min-height: 80px;
		padding: 10px 12px;
		border-radius: 12px;
		border: 1px solid var(--ln2);
		background: var(--sf);
		color: var(--tx);
		font-size: 14px;
		line-height: 1.45;
		resize: vertical;
	}
	.s-acc {
		margin-top: 10px;
		display: flex;
		gap: 6px;
	}
	.s-acc button {
		padding: 8px 14px;
		border-radius: 999px;
		font-size: 13px;
		font-weight: 600;
	}
	.pub {
		border: 0;
		background: var(--tx);
		color: var(--sf);
	}
	.ed {
		border: 1px solid var(--ln2);
		background: transparent;
		color: var(--tx);
	}
	.publicada {
		margin-top: 10px;
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 13px;
		font-weight: 600;
		color: #2f8f63;
	}
	.r-txt {
		margin-top: 6px;
		font-size: 13.5px;
		line-height: 1.45;
		color: var(--mu);
	}
	.resenas {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
		gap: 16px;
		align-items: start;
	}
	.resumen {
		padding: 20px;
	}
	.r-prom {
		display: flex;
		align-items: baseline;
		gap: 10px;
	}
	.r-prom .n {
		font-size: 52px;
		font-weight: 600;
		letter-spacing: -0.05em;
	}
	.estrellas {
		color: #e3a93b;
		letter-spacing: 2px;
	}
	.r-prom .estrellas {
		font-size: 18px;
	}
	.r-s {
		font-size: 13.5px;
		color: var(--mu);
	}
	.barras {
		margin-top: 14px;
		display: flex;
		flex-direction: column;
		gap: 7px;
	}
	.barra {
		display: flex;
		align-items: center;
		gap: 10px;
		font-size: 13px;
	}
	.b-n {
		width: 14px;
	}
	.b-f {
		flex: 1;
		height: 8px;
		border-radius: 6px;
		background: var(--sf3);
		overflow: hidden;
	}
	.b-f div {
		height: 100%;
		border-radius: 6px;
		background: #e3a93b;
	}
	.b-c {
		width: 36px;
		text-align: right;
		color: var(--mu);
	}
	.resena {
		padding: 14px 16px;
		border-radius: 20px;
	}
	.re-cab {
		display: flex;
		align-items: center;
		gap: 8px;
	}
	.re-cab .estrellas {
		letter-spacing: 1px;
	}
	.re-q {
		flex: 1;
		font-size: 13px;
		color: var(--mu);
	}
	.re-h {
		font-size: 12px;
		color: var(--fa);
	}
	.re-t {
		margin-top: 6px;
		font-size: 14px;
		line-height: 1.5;
	}
	.vacio {
		padding: 28px;
		text-align: center;
		color: var(--mu);
	}
</style>
