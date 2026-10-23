<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import Icon from '#lib/components/Icon.svelte';
	import { phArrowLeftBold } from '#lib/icons.ts';
	import { FAMILIAS } from './familias.ts';

	/**
	 * /marca: sección interna con las piezas de marca listas para exportar. Fuera de (sitio): no
	 * lleva el header ni el footer públicos, sólo la barra para moverse entre familias.
	 */
	let { children } = $props();
</script>

<div class="marca">
	<nav aria-label="Marca">
		<a class="volver" href={resolve('/')}><Icon icon={phArrowLeftBold} size={15} />Mostrador</a>
		<div class="tabs">
			<a href={resolve('/marca')} aria-current={page.url.pathname === '/marca' ? 'page' : undefined}
				>Marca</a
			>
			{#each FAMILIAS as f (f.ruta)}
				<a href={f.href} aria-current={page.url.pathname === f.ruta ? 'page' : undefined}
					>{f.titulo}</a
				>
			{/each}
		</div>
	</nav>
	{@render children()}
</div>

<style>
	.marca {
		min-height: 100vh;
		background: #e7ded3;
		color: #2a140c;
		/* El handoff no usa cifras tabulares en las piezas. */
		font-variant-numeric: normal;
	}
	nav {
		display: flex;
		align-items: center;
		gap: 12px 20px;
		flex-wrap: wrap;
		padding: 16px 64px 0;
	}
	.volver {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		min-height: 44px;
		font-size: 14px;
		font-weight: 600;
		color: #6b5244;
	}
	.tabs {
		display: flex;
		gap: 8px;
		overflow-x: auto;
		max-width: 100%;
		scrollbar-width: none;
	}
	/* Chips del handoff: tostado el activo, blanco el resto. */
	.tabs a {
		flex: none;
		display: inline-flex;
		align-items: center;
		min-height: 36px;
		padding: 8px 14px;
		border-radius: 999px;
		background: #fff;
		color: #2a140c;
		font-size: 13px;
		font-weight: 600;
	}
	.tabs a[aria-current='page'] {
		background: #2a140c;
		color: #fff;
	}
	/* El lienzo de cada familia: el <section> del handoff. */
	.marca :global(.lienzo) {
		font-family: var(--font);
		color: #2a140c;
		padding: 56px 64px 120px;
		display: flex;
		flex-direction: column;
		gap: 72px;
	}
	.marca :global(.lienzo-cab) {
		display: flex;
		align-items: end;
		gap: 24px;
		flex-wrap: wrap;
		max-width: 1400px;
	}
	@media (max-width: 760px) {
		nav {
			padding: 12px 16px 0;
		}
		.tabs a {
			min-height: 44px;
		}
		.marca :global(.lienzo) {
			padding: 32px 16px 80px;
			gap: 48px;
		}
		.marca :global(.lienzo h1) {
			font-size: 40px !important;
		}
	}
</style>
