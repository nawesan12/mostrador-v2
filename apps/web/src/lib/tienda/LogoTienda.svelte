<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { TiendaPublica } from './tipos.ts';

	/**
	 * La marca de la cabecera: el logo subido en Mi tienda o, sin logo, el nombre como lo dibuja cada
	 * plantilla (children) o en letras.
	 */
	let {
		tienda,
		alto = 36,
		children
	}: {
		tienda: Pick<TiendaPublica, 'logo' | 'nombre'>;
		alto?: number;
		children?: Snippet;
	} = $props();
</script>

{#if tienda.logo}<img
		class="logo-tienda"
		src={tienda.logo}
		alt={tienda.nombre}
		style:height="{alto}px"
	/>{:else if children}{@render children()}{:else}{tienda.nombre}{/if}

<style>
	.logo-tienda {
		display: block;
		width: auto;
		max-width: min(180px, 42vw);
		object-fit: contain;
	}
</style>
