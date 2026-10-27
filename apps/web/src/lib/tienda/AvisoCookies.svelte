<script lang="ts">
	import { onMount } from 'svelte';
	import { guardarConsentimiento, leerConsentimiento, pideConsentimiento } from './medicion.ts';
	import type { TiendaPublica } from './tipos.ts';

	/**
	 * El aviso de cookies de la tienda (armado con las piezas del CookieBanner de la landing). Sólo
	 * aparece si la tienda cargó GA4 o el píxel de Meta y quien compra todavía no eligió. En el
	 * celular va abajo, a todo el ancho, y le suma ese alto al final de la página para no tapar nada.
	 */
	let { tienda }: { tienda: TiendaPublica } = $props();

	let visible = $state(false);
	let caja = $state<HTMLElement | null>(null);

	const que = $derived(
		[tienda.medicion.ga4 && 'Google Analytics', tienda.medicion.pixel && 'el píxel de Meta']
			.filter(Boolean)
			.join(' y ')
	);

	onMount(() => {
		visible = pideConsentimiento(tienda) && leerConsentimiento(tienda.slug) === null;
	});

	// Mientras está a la vista, la página se alarga lo que mide el aviso (nada queda tapado).
	$effect(() => {
		if (!visible || !caja) return;
		const el = caja;
		const ajustar = () =>
			document.body.style.setProperty('padding-bottom', `${el.offsetHeight + 24}px`);
		ajustar();
		const ro = new ResizeObserver(ajustar);
		ro.observe(el);
		return () => {
			ro.disconnect();
			document.body.style.removeProperty('padding-bottom');
		};
	});

	function elegir(si: boolean) {
		guardarConsentimiento(tienda.slug, si);
		visible = false;
	}
</script>

{#if visible}
	<div class="ck" role="region" aria-label="Cookies de la tienda" bind:this={caja}>
		<div class="ck-txt">
			<b>Cookies de medición</b>
			<p>
				{tienda.nombre} usa {que} para saber qué funciona y mejorar sus anuncios. Sólo si vos&nbsp;aceptás.
			</p>
		</div>
		<div class="ck-acc">
			<button type="button" class="sec" onclick={() => elegir(false)}>Sólo necesarias</button>
			<button type="button" class="pri" onclick={() => elegir(true)}>Aceptar</button>
		</div>
	</div>
{/if}

<style>
	.ck {
		position: fixed;
		left: 16px;
		bottom: 16px;
		z-index: 80;
		width: min(400px, calc(100vw - 32px));
		padding: 16px;
		border-radius: 20px;
		background: #fff;
		box-shadow:
			0 2px 4px rgba(20, 16, 12, 0.08),
			0 24px 48px -20px rgba(20, 16, 12, 0.45);
		color: #1c1a17;
		font-family: 'Instrument Sans', system-ui, sans-serif;
	}
	.ck-txt b {
		font-size: 15px;
	}
	.ck-txt p {
		margin: 4px 0 0;
		font-size: 13.5px;
		line-height: 1.5;
		color: #57514a;
	}
	.ck-acc {
		margin-top: 12px;
		display: flex;
		gap: 6px;
	}
	.ck-acc button {
		flex: 1 1 0;
		min-height: 44px;
		padding: 10px 14px;
		border-radius: 999px;
		font: inherit;
		font-size: 13.5px;
		font-weight: 600;
		white-space: nowrap;
		cursor: pointer;
	}
	.sec {
		border: 1.5px solid rgba(20, 16, 12, 0.16);
		background: #fff;
		color: #1c1a17;
	}
	.pri {
		border: 0;
		background: #1c1a17;
		color: #fff;
	}
	@media (max-width: 760px) {
		.ck {
			left: 0;
			bottom: 0;
			width: 100%;
			border-radius: 20px 20px 0 0;
			padding: 14px 16px calc(14px + env(safe-area-inset-bottom));
		}
		.ck-txt p {
			font-size: 13px;
		}
	}
</style>
