<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
	import { imagenTienda, textosCompartir } from '#lib/tienda/compartir.ts';
	import { onMount } from 'svelte';
	import Plantilla from '#lib/tienda/Plantilla.svelte';
	import { compra, preparar } from '#lib/tienda/medicion.ts';

	let { data } = $props();
	onMount(() => {
		preparar(data.tienda);
		compra(data.pedido);
	});
</script>

<Seo
	title="Pedido #{data.pedido.numero} · {data.tienda.nombre}"
	description={textosCompartir(data.tienda).descripcion}
	path="/t/{data.tienda.slug}"
	noindex
	imagen={imagenTienda(data.tienda)}
/>
<Plantilla tienda={data.tienda} pagina="listo" pedido={data.pedido} />
