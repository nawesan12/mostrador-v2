<script lang="ts">
	import TplMini from '#lib/components/TplMini.svelte';
	import { CON_PORTADA, type Plantilla } from '#lib/plantillas.ts';

	/** Tarjeta de plantilla: la portada real si tiene versión trabajada, si no la miniatura. */
	let { t }: { t: Plantilla } = $props();
</script>

{#if CON_PORTADA.has(t.id)}
	<img
		class="portada"
		src="/plantillas/{t.id}.jpg"
		alt=""
		loading="lazy"
		width="560"
		height="611"
	/>
{:else}
	<TplMini {t} />
{/if}

<style>
	.portada {
		display: block;
		width: 100%;
		height: 100%;
		object-fit: cover;
		object-position: top center;
	}
</style>
