<script lang="ts">
	import Icon from './Icon.svelte';
	import { marca } from '#lib/marcas.ts';

	/** Cuadradito con el ícono o la sigla de un canal, medio de pago o tipo de aviso. */
	let {
		k,
		size = 38,
		radius = 12,
		fondo
	}: {
		k: string;
		size?: number;
		radius?: number;
		/** Pisa el fondo de la marca. */ fondo?: string;
	} = $props();

	const m = $derived(marca(k));
</script>

<span
	class="marca"
	style:width="{size}px"
	style:height="{size}px"
	style:border-radius="{radius}px"
	style:background={fondo ?? m.bg}
	style:color={m.col ?? 'inherit'}
>
	{#if m.icon}
		<Icon icon={m.icon} size={Math.round(size * 0.53)} />
	{:else}
		<span class="sigla" style:font-size="{Math.max(7, Math.round(size * 0.22))}px">{m.sigla}</span>
	{/if}
</span>

<style>
	.marca {
		flex: none;
		display: grid;
		place-items: center;
	}
	.sigla {
		font-weight: 800;
		line-height: 1;
		text-align: center;
	}
</style>
