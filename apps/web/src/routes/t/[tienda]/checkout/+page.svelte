<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
	import { imagenTienda, textosCompartir } from '#lib/tienda/compartir.ts';
	import { onMount } from 'svelte';
	import Plantilla from '#lib/tienda/Plantilla.svelte';
	import { carritoDe } from '#lib/tienda/carrito.svelte.ts';
	import { iniciarCompra, preparar } from '#lib/tienda/medicion.ts';

	let { data, form } = $props();
	onMount(() => {
		preparar(data.tienda);
		iniciarCompra(carritoDe(data.tienda.slug).lineas);
	});
</script>

<Seo
	title="Finalizar compra · {data.tienda.nombre}"
	description={textosCompartir(data.tienda).descripcion}
	path="/t/{data.tienda.slug}/checkout"
	noindex
	imagen={imagenTienda(data.tienda)}
/>
<Plantilla tienda={data.tienda} pagina="checkout" error={form?.error ?? null} />
