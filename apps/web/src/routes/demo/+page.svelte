<script lang="ts">
	import { enhance } from '$app/forms';
	import { page } from '$app/state';
	import Alerta from '#lib/components/Alerta.svelte';
	import Grano from '#lib/components/Grano.svelte';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import TplMini from '#lib/components/TplMini.svelte';
	import { phCheckBold, phMagnifyingGlass } from '#lib/icons.ts';
	import { CON_PORTADA, LAYOUT, PLANTILLAS, RUBROS_PLANTILLAS } from '#lib/plantillas.ts';

	/*
	 * El handoff no tiene una pantalla de "entrar a la demo": se arma con la galería de plantillas del
	 * onboarding (design/Onboarding Mostrador.dc.html, `gallery`), tal cual, como página.
	 */
	let { data, form } = $props();
	// Viene del consentimiento de una IA: después de elegir, vuelve ahí.
	const volver = $derived(page.url.searchParams.get('volver') ?? '');

	let rubro = $state<string>('Todos');
	let q = $state('');
	let eligiendo = $state<string | null>(null);

	// Chips: "Todos" y los rubros que tienen alguna plantilla (handoff: rubros filtrados).
	const chips = RUBROS_PLANTILLAS.filter(
		(r) => r === 'Todos' || PLANTILLAS.some((t) => t.rubro === r)
	);
	const lista = $derived.by(() => {
		const b = q.trim().toLowerCase();
		return PLANTILLAS.filter(
			(t) =>
				(rubro === 'Todos' || t.rubro === rubro) &&
				(!b || `${t.nombre} ${t.rubro} ${LAYOUT[t.layout]}`.toLowerCase().includes(b))
		);
	});
</script>

<Seo
	title="Probá Mostrador · Elegí una tienda"
	description="Elegí una tienda de ejemplo y recorré el panel completo con sus productos, pedidos y clientes."
	path="/demo"
	image="plantillas"
/>

<Grano />

<main class="gal">
	<div class="cab">
		<div class="tit">
			<h1>Probá Mostrador</h1>
			<p>Elegí una tienda de ejemplo. La demo es tuya por 24 horas.</p>
		</div>
		<label class="buscar">
			<Icon icon={phMagnifyingGlass} size={15} style="color:#7A6252" />
			<input
				bind:value={q}
				placeholder="Buscar estilo o rubro"
				aria-label="Buscar estilo o rubro"
			/>
		</label>
	</div>

	<div class="chips" role="group" aria-label="Rubro">
		{#each chips as r (r)}
			<button
				type="button"
				class:on={rubro === r}
				aria-pressed={rubro === r}
				onclick={() => (rubro = r)}>{r}</button
			>
		{/each}
	</div>

	{#if form?.error}
		<div class="aviso">
			<Alerta tipo="err" titulo="No pudimos armar la demo" texto={form.error} />
		</div>
	{/if}

	<form
		method="POST"
		class="cuerpo"
		use:enhance={({ formData }) => {
			eligiendo = String(formData.get('plantilla'));
			return async ({ update }) => {
				await update();
				eligiendo = null;
			};
		}}
	>
		{#if volver}<input type="hidden" name="volver" value={volver} />{/if}
		<div class="grilla">
			{#each lista as t (t.id)}
				{@const sel = data.actual === t.id}
				<button
					type="submit"
					name="plantilla"
					value={t.id}
					class="tpl"
					disabled={eligiendo !== null}
				>
					<div
						class="marco"
						aria-hidden="true"
						style:box-shadow="0 0 0 2px {sel ? '#D2582F' : 'transparent'},0 20px 40px -24px
						rgba(58,29,18,.55)"
					>
						{#if CON_PORTADA.has(t.id)}
							<img
								class="portada"
								src="/plantillas/{t.id}.jpg"
								alt=""
								loading="lazy"
								width="560"
								height="611"
							/>
						{:else}
							<TplMini {t} />
						{/if}
						{#if sel}
							<span class="elegida"><Icon icon={phCheckBold} size={12} />Elegida</span>
						{:else if eligiendo === t.id}
							<span class="elegida">Armando…</span>
						{/if}
					</div>
					<div class="pie">
						<span class="nom">{t.nombre}</span>
						<span class="sub">{t.rubro} · {LAYOUT[t.layout]}</span>
					</div>
				</button>
			{/each}
		</div>
		{#if lista.length === 0}
			<div class="vacia">No hay plantillas con «{q}».</div>
		{/if}
	</form>
</main>

<style>
	.gal {
		min-height: 100svh;
		background: #fbf7f2;
		color: #2a140c;
		display: flex;
		flex-direction: column;
	}
	.cab {
		display: flex;
		align-items: center;
		gap: 14px;
		padding: 20px 24px;
		border-bottom: 1px solid rgba(58, 29, 18, 0.08);
		flex-wrap: wrap;
	}
	.tit {
		flex: 1;
		min-width: 240px;
	}
	h1 {
		margin: 0;
		font-size: 24px;
		font-weight: 600;
		letter-spacing: -0.03em;
	}
	.tit p {
		margin: 0;
		font-size: 13.5px;
		color: #7a6252;
	}
	.buscar {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 9px 14px;
		border-radius: 999px;
		background: #fff;
		border: 1px solid rgba(58, 29, 18, 0.12);
		min-width: 220px;
	}
	.buscar input {
		flex: 1;
		min-width: 0;
		border: 0;
		outline: 0;
		background: transparent;
		font:
			500 14px 'Instrument Sans Variable',
			'Instrument Sans',
			sans-serif;
		color: #2a140c;
	}
	.chips {
		display: flex;
		gap: 6px;
		padding: 14px 24px;
		overflow-x: auto;
		flex: none;
	}
	.chips button {
		border: 1px solid rgba(58, 29, 18, 0.14);
		font-size: 13px;
		font-weight: 600;
		padding: 8px 14px;
		border-radius: 999px;
		background: #fff;
		color: #2a140c;
		white-space: nowrap;
	}
	.chips button.on {
		border-color: #2a140c;
		background: #2a140c;
		color: #fff;
	}
	.aviso {
		padding: 0 24px 12px;
		max-width: 620px;
	}
	.cuerpo {
		flex: 1;
		padding: 4px 24px 24px;
	}
	.grilla {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(230px, 1fr));
		gap: 16px;
	}
	.tpl {
		padding: 0;
		border: 0;
		background: transparent;
		font-family: inherit;
		text-align: left;
		color: #2a140c;
		transition: transform 0.3s var(--ease);
	}
	.tpl:hover:not(:disabled) {
		transform: translateY(-4px);
	}
	.tpl:disabled {
		cursor: progress;
	}
	.marco {
		position: relative;
		height: 290px;
		border-radius: 16px;
		overflow: hidden;
		background: #fff;
	}
	/* Portada real de la plantilla trabajada, anclada arriba como la vista de la tienda. */
	.portada {
		display: block;
		width: 100%;
		height: 100%;
		object-fit: cover;
		object-position: top center;
	}
	.elegida {
		position: absolute;
		top: 10px;
		right: 10px;
		display: flex;
		align-items: center;
		gap: 5px;
		font-size: 12px;
		font-weight: 700;
		color: #fff;
		background: #d2582f;
		padding: 5px 10px;
		border-radius: 999px;
	}
	.pie {
		margin-top: 9px;
		display: grid;
	}
	.nom {
		font-size: 15px;
		font-weight: 600;
	}
	.sub {
		font-size: 12.5px;
		color: #7a6252;
	}
	.vacia {
		padding: 60px;
		text-align: center;
		color: #7a6252;
	}
	@media (max-width: 760px) {
		.cab,
		.chips,
		.cuerpo {
			padding-left: 16px;
			padding-right: 16px;
		}
		/* Como en el handoff: el buscador baja a su propia fila. */
		.buscar {
			flex: 1 1 100%;
			min-width: 0;
		}
		.chips button {
			min-height: 40px;
		}
	}
</style>
