<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import type { IconData } from '#lib/icons.ts';

	/** Logo de una integración del handoff: imagen, ícono (Iconify) o sigla en recuadro de color. */
	type Marca = {
		icon?: IconData;
		img?: string;
		col?: string;
		w?: string;
		abbr?: string;
		bg?: string;
		fg?: string;
	};
	let { m, size = 22 }: { m: Marca; size?: number } = $props();
</script>

{#if m.img}
	<img src={m.img} alt="" width={size} height={size} class="img" />
{:else if m.icon}
	<Icon icon={m.icon} {size} style={m.col ? `color:${m.col}` : ''} />
{:else if m.w}
	<span class="sigla" style:background={m.bg ?? '#fff'} style:color={m.fg}>{m.abbr}</span>
{/if}

<style>
	.img {
		object-fit: contain;
		flex: none;
	}
	.sigla {
		height: 22px;
		padding: 0 6px;
		border-radius: 5px;
		display: grid;
		place-items: center;
		font-size: 10px;
		font-weight: 800;
		flex: none;
	}
</style>
