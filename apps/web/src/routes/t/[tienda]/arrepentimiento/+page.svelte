<script lang="ts">
	import { enhance } from '$app/forms';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import { imagenTienda } from '#lib/tienda/compartir.ts';
	import * as I from '#lib/icons.ts';
	import { plantilla as buscarPlantilla } from '#lib/plantillas.ts';
	import type { PageProps } from './$types';

	/**
	 * Botón de arrepentimiento de la tienda. No hay diseño propio: las piezas del formulario de
	 * Legales (Revocar mi contratación), con los colores y la tipografía de la plantilla de la tienda.
	 */
	let { data, form }: PageProps = $props();
	const t = $derived(data.tienda);
	const p = $derived(buscarPlantilla(t.plantilla));
	let enviando = $state(false);
	const previo = $derived(
		((form as { previo?: Record<string, string> } | null)?.previo ?? {}) as Record<string, string>
	);
</script>

<Seo
	title="Botón de arrepentimiento · {t.nombre}"
	description="Revocá tu compra en {t.nombre} hasta 10 días después de recibirla."
	path="/t/{t.slug}/arrepentimiento"
	noindex
	imagen={imagenTienda(t)}
/>

<div
	class="pagina"
	style="--bg:{p?.bg ?? '#FBF7F2'};--tx:{p?.tx ?? '#2A140C'};--ac:{p?.ac ??
		'#C24E25'};--ff:{p?.ff ?? 'inherit'};--r:{p?.radio ?? '14px'}"
>
	<div class="ancho">
		<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- ruta de la tienda con su slug -->
		<a class="volver" href="/t/{t.slug}"><Icon icon={I.phArrowLeftBold} size={14} />{t.nombre}</a>
		<h1>Botón de arrepentimiento</h1>
		<p class="bajada">
			Podés revocar tu compra sin costo y sin dar explicaciones, hasta 10 días corridos después de
			recibirla (Ley 24.240, artículo&nbsp;34).
		</p>

		<div class="revocar">
			{#if form?.tramite}
				<div class="exito" role="status">
					<img src="/hornero/exito.webp" alt="" width="88" height="88" />
					<div>
						<b>Recibimos tu pedido</b>
						<p>
							Tu código de trámite es <b class="codigo">{form.tramite}</b>. Te lo mandamos por
							email.
							{t.nombre} se va a comunicar con vos para coordinar la devolución y el&nbsp;reintegro.
						</p>
					</div>
				</div>
			{:else}
				<form
					method="POST"
					use:enhance={() => {
						enviando = true;
						return async ({ update }) => {
							await update({ reset: false });
							enviando = false;
						};
					}}
					novalidate
				>
					<div class="rv-cab">
						<span class="rv-ico"><Icon icon={I.phArrowCounterClockwiseBold} size={19} /></span>
						<b>Revocar mi compra</b>
					</div>
					{#if form?.error}<div class="err-form" role="alert">{form.error}</div>{/if}
					<div class="campos">
						<div class="campo">
							<label for="ar-pedido">Código o número del pedido</label>
							<input
								id="ar-pedido"
								name="pedido"
								autocomplete="off"
								placeholder="Está en el mail de la compra"
								value={previo.pedido ?? data.codigo}
							/>
						</div>
						<div class="campo">
							<label for="ar-email">Email de la compra</label>
							<input
								id="ar-email"
								name="email"
								type="email"
								autocomplete="email"
								value={previo.email ?? ''}
							/>
						</div>
						<div class="campo">
							<label for="ar-nombre">Nombre y apellido</label>
							<input id="ar-nombre" name="nombre" autocomplete="name" value={previo.nombre ?? ''} />
						</div>
					</div>
					<div class="campo">
						<label for="ar-motivo">Motivo (opcional)</label>
						<textarea
							id="ar-motivo"
							name="motivo"
							rows="3"
							placeholder="Contanos si querés, no es obligatorio">{previo.motivo ?? ''}</textarea
						>
					</div>
					<input class="trampa" name="web" tabindex="-1" autocomplete="off" aria-hidden="true" />
					<button class="enviar" type="submit" disabled={enviando}
						>{enviando ? 'Enviando…' : 'Enviar pedido de arrepentimiento'}</button
					>
					<div class="rv-pie">
						Sin costo y sin dar explicaciones. Te llega el código por&nbsp;email.
					</div>
				</form>
			{/if}
		</div>
	</div>
</div>

<style>
	.pagina {
		min-height: 100vh;
		background: var(--bg);
		color: var(--tx);
		font-family: var(--ff);
	}
	.ancho {
		max-width: 680px;
		margin: 0 auto;
		padding: 28px 16px 64px;
	}
	.volver {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		color: inherit;
		font-size: 14px;
		font-weight: 600;
		text-decoration: none;
	}
	h1 {
		margin: 24px 0 0;
		font-size: clamp(30px, 6vw, 44px);
		line-height: 1.05;
		letter-spacing: -0.03em;
		text-wrap: balance;
	}
	.bajada {
		margin: 12px 0 0;
		font-size: 16px;
		line-height: 1.5;
		opacity: 0.8;
		text-wrap: pretty;
	}
	.revocar {
		margin-top: 28px;
		padding: clamp(18px, 4vw, 28px);
		border-radius: var(--r);
		background: color-mix(in srgb, var(--bg) 60%, #fff);
		box-shadow: 0 0 0 1px color-mix(in srgb, var(--tx) 10%, transparent);
	}
	.rv-cab {
		display: flex;
		align-items: center;
		gap: 10px;
		font-size: 18px;
	}
	.rv-ico {
		width: 36px;
		height: 36px;
		border-radius: 50%;
		display: grid;
		place-items: center;
		background: var(--ac);
		color: #fff;
	}
	.campos {
		margin-top: 18px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 240px), 1fr));
		gap: 14px;
	}
	.campo {
		display: grid;
		gap: 6px;
		margin-top: 14px;
	}
	.campos .campo {
		margin-top: 0;
	}
	label {
		font-size: 13.5px;
		font-weight: 600;
	}
	input:not(.trampa),
	textarea {
		width: 100%;
		box-sizing: border-box;
		padding: 12px 14px;
		border: 1px solid color-mix(in srgb, var(--tx) 18%, transparent);
		border-radius: min(var(--r), 12px);
		background: #fff;
		color: var(--tx);
		font: inherit;
		font-size: 15px;
	}
	.trampa {
		position: absolute;
		left: -9999px;
		width: 1px;
		height: 1px;
	}
	.enviar {
		margin-top: 18px;
		width: 100%;
		min-height: 48px;
		border: 0;
		border-radius: min(var(--r), 999px);
		background: var(--tx);
		color: var(--bg);
		font: inherit;
		font-size: 15.5px;
		font-weight: 600;
		cursor: pointer;
	}
	.enviar:disabled {
		opacity: 0.6;
	}
	.rv-pie {
		margin-top: 10px;
		font-size: 13px;
		opacity: 0.7;
		text-align: center;
	}
	.err-form {
		margin-top: 14px;
		padding: 10px 12px;
		border-radius: 10px;
		background: #fbe9e6;
		color: #8c2a1f;
		font-size: 14px;
	}
	.exito {
		display: flex;
		gap: 16px;
		align-items: center;
	}
	.exito p {
		margin: 6px 0 0;
		line-height: 1.5;
		text-wrap: pretty;
	}
	.codigo {
		font-family: ui-monospace, monospace;
	}
	@media (max-width: 480px) {
		.exito {
			flex-direction: column;
			align-items: flex-start;
		}
	}
</style>
