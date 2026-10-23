<script lang="ts">
	import { afterNavigate } from '$app/navigation';
	import AvisoCookies from '#lib/tienda/AvisoCookies.svelte';
	import { pagina, preparar } from '#lib/tienda/medicion.ts';

	let { data, children } = $props();

	// La medición de la tienda (GA4 y píxel de Meta, sólo con consentimiento): una página vista por
	// navegación. Las plantillas no saben nada de esto.
	afterNavigate(() => {
		preparar(data.tienda);
		pagina();
	});
</script>

<svelte:head>
	<link rel="preconnect" href="https://fonts.googleapis.com" />
	<link
		href="https://fonts.googleapis.com/css2?family=Instrument+Sans:wght@400;500;600;700&family=Instrument+Serif:ital@0;1&display=swap"
		rel="stylesheet"
	/>
</svelte:head>

{@render children()}
{#key data.tienda.slug}
	<AvisoCookies tienda={data.tienda} />
{/key}
