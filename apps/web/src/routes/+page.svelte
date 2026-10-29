<script lang="ts">
	import { onMount } from 'svelte';
	import CookieBanner from '#lib/components/CookieBanner.svelte';
	import Grano from '#lib/components/Grano.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import Header from '#lib/landing/Header.svelte';
	import Etapas from '#lib/landing/Etapas.svelte';
	import Caja from '#lib/landing/Caja.svelte';
	import ConectorIA from '#lib/landing/ConectorIA.svelte';
	import CtaFinal from '#lib/landing/CtaFinal.svelte';
	import Footer from '#lib/landing/Footer.svelte';
	import Historias from '#lib/landing/Historias.svelte';
	import Integraciones from '#lib/landing/Integraciones.svelte';
	import Migrar from '#lib/landing/Migrar.svelte';
	import PlantillaModal from '#lib/landing/PlantillaModal.svelte';
	import Plantillas from '#lib/landing/Plantillas.svelte';
	import Precios from '#lib/landing/Precios.svelte';
	import Producto from '#lib/landing/Producto.svelte';
	import Seguridad from '#lib/landing/Seguridad.svelte';
	import WidgetHornero from '#lib/landing/WidgetHornero.svelte';
	import Hero from '#lib/landing/Hero.svelte';
	import HorneroSeccion from '#lib/landing/HorneroSeccion.svelte';
	import Lenis from '#lib/landing/Lenis.svelte';
	import Marquesina from '#lib/landing/Marquesina.svelte';
	import { CHAT } from '#lib/landing/datos.ts';
	import { Landing, setLanding } from '#lib/landing/estado.svelte.ts';

	/* Landing de Tu Local Web: recreación de la vista "Landing" de design/Mostrador v4.dc.html. */
	const l = setLanding(new Landing());
	onMount(() => {
		l.iniciar(CHAT.length);
		return () => l.detener();
	});
</script>

<Seo
	title="Tu Local Web · Vendé en todos lados, pagá menos"
	description="Tienda online, WhatsApp, Mercado Libre, local y factura de ARCA en un solo panel. Más barato que Tiendanube en cada plan."
	path="/"
	image="home"
	jsonLd={{
		'@context': 'https://schema.org',
		'@type': 'Organization',
		name: 'Tu Local Web',
		url: 'https://tulocalweb.com.ar',
		logo: 'https://tulocalweb.com.ar/mark.svg'
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
		<Plantillas />
		<Producto />
		<ConectorIA />
		<Caja />
		<Integraciones />
		<Precios />
		<Historias />
		<Seguridad />
		<Migrar />
		<CtaFinal />
	</main>
	<Footer />
</div>

<PlantillaModal />
<WidgetHornero />

<CookieBanner />

<style>
	.landing {
		position: relative;
		overflow-x: clip;
		background: #fbf7f2;
		color: #2a140c;
	}
</style>
