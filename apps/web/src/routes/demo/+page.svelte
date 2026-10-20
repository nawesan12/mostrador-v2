<script lang="ts">
	import { enhance } from '$app/forms';
	import { resolve } from '$app/paths';
	import Alerta from '#lib/components/Alerta.svelte';
	import Grano from '#lib/components/Grano.svelte';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import { anim } from '#lib/anim.ts';
	import { phArrowRightBold, phCheckBold } from '#lib/icons.ts';
	import { plantilla } from '#lib/plantillas.ts';

	let { data, form } = $props();
	let eligiendo = $state<string | null>(null);
</script>

<Seo
	title="Probá Mostrador · Elegí tu rubro"
	description="Armamos una tienda de ejemplo de tu rubro, con productos, pedidos y clientes, para que pruebes el panel completo."
	path="/demo"
	image="general"
/>

<Grano />

<main>
	<header>
		<a href={resolve('/')} aria-label="Mostrador, volver al inicio">
			<img src="/wordmark-brown.png" alt="Mostrador" width="150" height="38" />
		</a>
	</header>

	<div class="intro" use:anim={'rise'}>
		<div class="etiqueta">Probá Mostrador</div>
		<h1>¿Qué vendés?</h1>
		<p>
			Armamos una tienda de ejemplo de tu rubro, con productos, pedidos y clientes. Es tuya por 24
			horas: tocá todo lo que quieras.
		</p>
	</div>

	{#if form?.error}
		<div class="aviso">
			<Alerta tipo="err" titulo="No pudimos armar la demo" texto={form.error} />
		</div>
	{/if}

	<form
		method="POST"
		use:enhance={({ formData }) => {
			eligiendo = String(formData.get('rubro'));
			return async ({ update }) => {
				await update();
				eligiendo = null;
			};
		}}
	>
		<ul class="rubros">
			{#each data.rubros as r, i (r.id)}
				{@const p = plantilla(r.plantilla)}
				<li use:anim={{ t: 'rise', delay: 40 * i }}>
					<button
						type="submit"
						name="rubro"
						value={r.id}
						class="rubro"
						class:actual={data.actual === r.id}
						disabled={eligiendo !== null}
						aria-describedby="tienda-{r.id}"
					>
						<span
							class="vidriera"
							style:--bg={p?.bg}
							style:--tx={p?.tx}
							style:--ac={p?.ac}
							style:--ff={p?.serif ? "Georgia, 'Times New Roman', serif" : 'var(--font)'}
						>
							<span class="tienda" id="tienda-{r.id}">{r.tienda}</span>
							<span class="lema"><i></i>{p?.lema}</span>
						</span>
						<span class="pie">
							<span class="nombre">{r.nombre}</span>
							{#if eligiendo === r.id}
								<span class="estado">Armando…</span>
							{:else if data.actual === r.id}
								<span class="estado"><Icon icon={phCheckBold} size={13} />Tu demo</span>
							{:else}
								<Icon icon={phArrowRightBold} size={15} />
							{/if}
						</span>
					</button>
				</li>
			{/each}
		</ul>
	</form>

	<p class="nota">La demo no pide datos ni tarjeta. A las 24 horas se borra sola.</p>
</main>

<style>
	main {
		max-width: 1240px;
		margin: 0 auto;
		padding: 28px 28px 96px;
	}
	header img {
		display: block;
		width: 150px;
		height: auto;
	}
	.intro {
		margin-top: 56px;
		max-width: 640px;
	}
	h1 {
		margin: 12px 0 0;
		font-size: clamp(40px, 6vw, 72px);
		font-weight: 600;
		line-height: 0.95;
		letter-spacing: -0.045em;
	}
	.intro p {
		margin: 16px 0 0;
		font-size: 17px;
		line-height: 1.5;
		color: var(--marron);
	}
	.aviso {
		margin-top: 24px;
		max-width: 560px;
	}
	.rubros {
		margin: 40px 0 0;
		padding: 0;
		list-style: none;
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(min(210px, 100%), 1fr));
		gap: 14px;
	}
	.rubro {
		width: 100%;
		height: 100%;
		display: flex;
		flex-direction: column;
		padding: 8px;
		border: 1px solid rgba(58, 29, 18, 0.09);
		border-radius: 26px;
		background: #fff;
		color: var(--tostado);
		text-align: left;
		box-shadow: var(--shd);
		transition:
			transform 0.25s var(--ease),
			box-shadow 0.25s;
	}
	.rubro:hover:not(:disabled) {
		transform: translateY(-3px);
		box-shadow:
			inset 0 1px 0 #fff,
			0 20px 40px -20px rgba(58, 29, 18, 0.35);
	}
	.rubro:disabled {
		cursor: progress;
	}
	.rubro.actual {
		border-color: var(--terracota);
	}
	/* Vidriera con el arco de la marca: la tienda de ejemplo con los colores de su plantilla. */
	.vidriera {
		position: relative;
		aspect-ratio: 4 / 3;
		display: flex;
		flex-direction: column;
		justify-content: flex-end;
		gap: 6px;
		padding: 16px;
		border-radius: 999px 999px 18px 18px;
		background: var(--bg);
		color: var(--tx);
		overflow: hidden;
	}
	.vidriera::before {
		content: '';
		position: absolute;
		inset: 0;
		background: url('/textura/arch-pattern-light.svg') center / 46px;
		opacity: 0.35;
		mix-blend-mode: multiply;
	}
	.tienda {
		position: relative;
		font-family: var(--ff);
		font-size: 21px;
		font-weight: 600;
		line-height: 1.05;
		letter-spacing: -0.02em;
	}
	.lema {
		position: relative;
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 12px;
		font-weight: 600;
		opacity: 0.85;
	}
	.lema i {
		width: 7px;
		height: 7px;
		flex: none;
		border-radius: 50%;
		background: var(--ac);
	}
	.pie {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
		min-height: 44px;
		padding: 10px 8px 4px;
	}
	.nombre {
		font-size: 15.5px;
		font-weight: 600;
	}
	.estado {
		display: inline-flex;
		align-items: center;
		gap: 4px;
		font-size: 12.5px;
		font-weight: 600;
		color: var(--terracota-oscuro);
		white-space: nowrap;
	}
	.nota {
		margin-top: 32px;
		font-size: 14px;
		color: var(--marron-2);
	}
	@media (max-width: 760px) {
		main {
			padding: 20px 16px 72px;
		}
		.intro {
			margin-top: 36px;
		}
		.rubros {
			grid-template-columns: repeat(2, minmax(0, 1fr));
			gap: 10px;
		}
		.vidriera {
			padding: 12px;
		}
		.tienda {
			font-size: 17px;
		}
		.lema {
			font-size: 11px;
		}
	}
</style>
