<script lang="ts">
	import CookieBanner from '#lib/components/CookieBanner.svelte';
	import SitioFooter from '#lib/sitio/SitioFooter.svelte';
	import SitioHeader from '#lib/sitio/SitioHeader.svelte';

	/** Páginas públicas del sitio (legales, seguridad, contacto…): mismo header y footer. */
	let { children } = $props();
</script>

<div class="sitio">
	<SitioHeader />
	<main>{@render children()}</main>
	<SitioFooter />
</div>
<CookieBanner />

<style>
	.sitio {
		min-height: 100vh;
		display: flex;
		flex-direction: column;
		background: #fbf7f2;
		color: #2a140c;
	}
	main {
		flex: 1;
	}
</style>
