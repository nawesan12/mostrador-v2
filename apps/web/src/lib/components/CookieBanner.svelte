<script lang="ts">
	import { onMount } from 'svelte';
	import { banner, guardarEleccion, leerEleccion } from '#lib/consentimiento.svelte.ts';

	/**
	 * Banner de cookies del handoff (design/CookieBanner.dc.html). Abajo a la izquierda. Viene en el
	 * HTML del servidor (así se pinta con la página, no al hidratar: era el LCP de la landing); a quien
	 * ya eligió lo oculta el script de app.html antes de pintar, y al montar se saca.
	 */
	let visible = $state(true);
	let cfg = $state(false);
	let an = $state(true);
	let mk = $state(false);

	$effect(() => {
		banner.abierto = visible;
	});

	onMount(() => {
		visible = leerEleccion() === null;
		// "Reconfigurar" desde la página de Cookies vuelve a abrir el panel.
		const abrir = () => {
			delete document.documentElement.dataset.cookies;
			cfg = true;
			visible = true;
		};
		window.addEventListener('mostrador:configurar-cookies', abrir);
		return () => window.removeEventListener('mostrador:configurar-cookies', abrir);
	});

	function guardar(v: Parameters<typeof guardarEleccion>[0]) {
		guardarEleccion(v);
		visible = false;
		cfg = false;
	}
</script>

{#if visible}
	<div class="ck" role="dialog" aria-label="Cookies" aria-live="polite">
		<div class="ck-top">
			<img src="/hornero/asomado.webp" alt="" width="54" height="54" />
			<div class="ck-txt">
				<b>Usamos cookies</b>
				<p>
					Las necesarias hacen funcionar el sitio. Con tu permiso sumamos analíticas y de marketing
					para mejorar Mostrador. <!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- /legales llega con las páginas públicas -->
					<a href="/legales/cookies">Ver política</a>
				</p>
			</div>
		</div>
		{#if cfg}
			<div class="ck-cfg">
				<div class="ck-fila">
					<div class="ck-f">
						<b>Necesarias</b>
						<div>Sesión, seguridad y tu elección de cookies</div>
					</div>
					<span class="ck-siempre">Siempre activas</span>
				</div>
				<div class="ck-fila">
					<div class="ck-f">
						<b>Analíticas</b>
						<div>Visitas agregadas y uso del panel</div>
					</div>
					<button
						type="button"
						role="switch"
						aria-checked={an}
						aria-label="Analíticas"
						class="tg"
						class:on={an}
						onclick={() => (an = !an)}><span></span></button
					>
				</div>
				<div class="ck-fila">
					<div class="ck-f">
						<b>Marketing</b>
						<div>Medir campañas en Meta y Google</div>
					</div>
					<button
						type="button"
						role="switch"
						aria-checked={mk}
						aria-label="Marketing"
						class="tg"
						class:on={mk}
						onclick={() => (mk = !mk)}><span></span></button
					>
				</div>
			</div>
		{/if}
		<div class="ck-acc">
			{#if !cfg}
				<button type="button" class="sec" onclick={() => (cfg = true)}>Configurar</button>
				<button type="button" class="sec" onclick={() => guardar('nec')}>Solo necesarias</button>
				<button type="button" class="pri" onclick={() => guardar('all')}>Aceptar todas</button>
			{:else}
				<button type="button" class="sec" onclick={() => guardar({ an, mk })}
					>Guardar mi elección</button
				>
				<button type="button" class="pri" onclick={() => guardar('all')}>Aceptar todas</button>
			{/if}
		</div>
	</div>
{/if}

<style>
	:global(html[data-cookies]) .ck {
		display: none;
	}
	.ck {
		position: fixed;
		left: 16px;
		bottom: 16px;
		z-index: 90;
		width: min(430px, calc(100vw - 32px));
		padding: 18px;
		border-radius: 24px;
		background: #fff;
		box-shadow:
			0 2px 4px rgba(58, 29, 18, 0.06),
			0 30px 60px -20px rgba(58, 29, 18, 0.45);
		color: #2a140c;
	}
	.ck-top {
		display: flex;
		gap: 12px;
		align-items: flex-start;
	}
	.ck-top img {
		width: 54px;
		height: 54px;
		object-fit: contain;
		flex: none;
	}
	.ck-txt {
		min-width: 0;
	}
	.ck-txt b {
		font-size: 15.5px;
	}
	.ck-txt p {
		margin: 4px 0 0;
		font-size: 13.5px;
		line-height: 1.5;
		color: #5c4536;
	}
	.ck-txt a {
		font-weight: 600;
	}
	.ck-cfg {
		margin-top: 14px;
		display: flex;
		flex-direction: column;
		gap: 2px;
		border-radius: 16px;
		overflow: hidden;
		background: #f4ece2;
	}
	.ck-fila {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 12px 14px;
		background: #fbf7f2;
	}
	.ck-f {
		flex: 1;
	}
	.ck-f b {
		font-size: 14px;
	}
	.ck-f div {
		font-size: 12.5px;
		color: #7a6252;
	}
	.ck-siempre {
		font-size: 12px;
		font-weight: 600;
		color: #2f8f63;
	}
	.tg {
		position: relative;
		width: 38px;
		height: 22px;
		flex: none;
		border: 0;
		padding: 0;
		border-radius: 999px;
		background: #d9cbbe;
		transition: background 0.25s;
	}
	.tg.on {
		background: #d2582f;
	}
	.tg span {
		position: absolute;
		top: 2.5px;
		left: 2.5px;
		width: 17px;
		height: 17px;
		border-radius: 50%;
		background: #fff;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.25);
		transition: left 0.25s;
	}
	.tg.on span {
		left: 18.5px;
	}
	.ck-acc {
		margin-top: 14px;
		display: flex;
		gap: 6px;
		flex-wrap: wrap;
	}
	.ck-acc button {
		flex: 1 1 auto;
		padding: 11px 14px;
		border-radius: 999px;
		font-size: 13.5px;
		font-weight: 600;
		white-space: nowrap;
	}
	.sec {
		border: 1.5px solid rgba(58, 29, 18, 0.16);
		background: #fff;
		color: #2a140c;
	}
	.pri {
		border: 0;
		background: #2a140c;
		color: #fff;
	}
	@media (max-width: 760px) {
		.ck-acc button {
			min-height: 44px;
		}
	}
</style>
