<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import { LOGOS } from './datos.ts';

	/** La marca de un servicio en su cuadrito (ícono o sigla), como en el handoff. */
	let {
		k,
		tam = 44,
		icono = 24,
		radio = 13
	}: { k: string; tam?: number; icono?: number; radio?: number } = $props();
	const l = $derived(LOGOS[k]);
</script>

<span
	class="logo"
	style:width="{tam}px"
	style:height="{tam}px"
	style:border-radius="{radio}px"
	style:background={l.bg}
>
	{#if l.icon}
		<Icon icon={l.icon} size={icono} style="color:{l.col}" />
	{:else}
		<span class="sigla" style:color={l.col} style:font-size="{Math.round(tam / 5)}px">{l.word}</span
		>
	{/if}
</span>

<style>
	.logo {
		flex: none;
		display: grid;
		place-items: center;
	}
	.sigla {
		font-weight: 800;
	}
</style>
