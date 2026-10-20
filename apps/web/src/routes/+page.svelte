<script lang="ts">
	import { onMount } from 'svelte';
	import CookieBanner from '#lib/components/CookieBanner.svelte';
	import Grano from '#lib/components/Grano.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import Header from '#lib/landing/Header.svelte';
	import Etapas from '#lib/landing/Etapas.svelte';
	import Hero from '#lib/landing/Hero.svelte';
	import HorneroSeccion from '#lib/landing/HorneroSeccion.svelte';
	import Lenis from '#lib/landing/Lenis.svelte';
	import Marquesina from '#lib/landing/Marquesina.svelte';
	import { CHAT } from '#lib/landing/datos.ts';
	import { Landing, setLanding } from '#lib/landing/estado.svelte.ts';

	/* Landing de Mostrador: recreación de la vista "Landing" de design/Mostrador v4.dc.html. */
	const l = setLanding(new Landing());
	onMount(() => {
		l.iniciar(CHAT.length);
		return () => l.detener();
	});
</script>

<Seo
	title="Mostrador · Vendé en todos lados, pagá menos"
	description="Tienda online, WhatsApp, Mercado Libre, local y factura de ARCA en un solo panel. Más barato que Tiendanube en cada plan."
	path="/"
	image="home"
	jsonLd={{
		'@context': 'https://schema.org',
		'@type': 'Organization',
		name: 'Mostrador',
		url: 'https://mostrador.iamnaw.dev',
		logo: 'https://mostrador.iamnaw.dev/mark.svg'
	}}
/>

<Grano />
<Lenis />

<div class="landing">
	<Header />
	<main>
		<Hero />
		<Marquesina />
		<HorneroSeccion />
		<Etapas />
	</main>
</div>

<CookieBanner />

<style>
	.landing {
		position: relative;
		overflow-x: clip;
		background: #fbf7f2;
		color: #2a140c;
	}
</style>
