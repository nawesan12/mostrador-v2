<script lang="ts">
	import { afterNavigate } from '$app/navigation';
	import AvisoCookies from '#lib/tienda/AvisoCookies.svelte';
	import { pagina, preparar } from '#lib/tienda/medicion.ts';

	let { data, children } = $props();

	// La medición de la tienda (GA4 y píxel de Meta, sólo con consentimiento): una página vista por
	// navegación. Las plantillas no saben nada de esto.
	afterNavigate(({ type, from, to }) => {
		preparar(data.tienda);
		pagina();
		// Visitas propias, sin cookies (decisiones §36): la carga completa ya la contó el servidor;
		// cada página nueva dentro de la tienda se avisa sin esperar respuesta.
		if (type === 'enter' || data.tienda.vitrina || !to?.route.id) return;
		if (from?.url.pathname === to.url.pathname) return;
		navigator.sendBeacon?.(
			`/t/${data.tienda.slug}/vista`,
			JSON.stringify({ r: to.route.id, p: to.params?.id ?? '' })
		);
	});

	import { page } from '$app/state';
	import { carritoDe } from '#lib/tienda/carrito.svelte.ts';

	// Carrito abandonado: en el checkout, cuando quien compra deja su email (y al cambiar el nombre),
	// se guarda el carrito para poder recordárselo. Escucha los campos `email` y `nombre` del form,
	// así ninguna plantilla tiene que hacer nada. Se manda sólo si cambió algo.
	let ultimo = '';
	function alSalir(e: FocusEvent) {
		const el = e.target;
		if (!(el instanceof HTMLInputElement) || !el.form || !['email', 'nombre'].includes(el.name))
			return;
		if (data.tienda.vitrina || !page.url.pathname.endsWith('/checkout')) return;
		const campo = (n: string) =>
			(el.form!.elements.namedItem(n) as HTMLInputElement | null)?.value.trim() ?? '';
		const email = campo('email');
		const carrito = carritoDe(data.tienda.slug).lineas;
		if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email) || email.length > 120 || !carrito.length) return;
		const cuerpo = JSON.stringify({ email, nombre: campo('nombre').slice(0, 80), carrito });
		if (cuerpo === ultimo) return;
		ultimo = cuerpo;
		fetch(`/t/${data.tienda.slug}/carrito/guardar`, {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: cuerpo,
			keepalive: true
		}).catch(() => {});
	}
</script>

<svelte:head>
	<link rel="preconnect" href="https://fonts.googleapis.com" />
	<link
		href="https://fonts.googleapis.com/css2?family=Instrument+Sans:wght@400;500;600;700&family=Instrument+Serif:ital@0;1&display=swap"
		rel="stylesheet"
	/>
</svelte:head>

<svelte:document onfocusout={alSalir} />

{@render children()}
{#key data.tienda.slug}
	<AvisoCookies tienda={data.tienda} />
{/key}
