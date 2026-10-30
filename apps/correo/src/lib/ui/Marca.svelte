<script lang="ts">
	import { inicial } from '#lib/marca.ts';

	/**
	 * El cuadradito de marca del handoff (36px, radio 8, fondo de acento). Con la marca de Grupo
	 * Vimo lleva su isotipo (la V con los dos puntos, static/vimo.svg) en blanco; otra marca, la
	 * inicial, como en el diseño.
	 */
	let { marca, tam = 36 }: { marca: string; tam?: number } = $props();
	const vimo = $derived(/vimo/i.test(marca));
</script>

<span class="logo" style:width="{tam}px" style:height="{tam}px" style:font-size="{tam * 0.47}px">
	{#if vimo}
		<svg viewBox="0 0 140 137" aria-hidden="true" width={tam * 0.62} height={tam * 0.62}>
			<g fill="currentColor"
				><circle cx="22.5" cy="21" r="18" /><circle cx="118.5" cy="21" r="18" /></g
			>
			<path
				d="M40.5 62 70.5 114 100.5 62"
				fill="none"
				stroke="currentColor"
				stroke-width="33"
				stroke-linecap="round"
				stroke-linejoin="round"
			/>
		</svg>
	{:else}
		{inicial(marca)}
	{/if}
</span>

<style>
	.logo {
		border-radius: 8px;
		flex: none;
		background: var(--acc);
		color: #fff;
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: 700;
		letter-spacing: -0.02em;
	}
</style>
