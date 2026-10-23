<script lang="ts">
	import type { ComponentProps } from 'svelte';
	import Ceibo from './plantillas/Ceibo.svelte';
	import Motor from './plantillas/Motor.svelte';

	/**
	 * Dibuja la tienda con su plantilla: Ceibo tiene su propio diseño; las demás salen del motor de
	 * plantillas del handoff (Plantillas Tienda.dc.html), un marcado con temas por plantilla.
	 */
	let props: ComponentProps<typeof Ceibo> = $props();
</script>

{#if props.tienda.plantilla === 'ceibo'}
	<Ceibo {...props} />
{:else}
	<Motor {...props} />
{/if}
