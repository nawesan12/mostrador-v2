<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
	import { afterNavigate } from '$app/navigation';
	import Plantilla from '#lib/tienda/Plantilla.svelte';
	import { imagenProducto, textosProducto } from '#lib/tienda/compartir.ts';
	import { preparar, verProducto } from '#lib/tienda/medicion.ts';

	let { data } = $props();
	const textos = $derived(textosProducto(data.tienda, data.producto));
	afterNavigate(() => {
		preparar(data.tienda);
		verProducto(data.producto);
	});
</script>

<Seo
	title={textos.titulo}
	description={textos.descripcion}
	imagen={imagenProducto(data.tienda, data.producto)}
	imageAlt={data.producto.nombre}
	path="/t/{data.tienda.slug}/producto/{data.producto.slug ?? data.producto.id}"
/>
<Plantilla tienda={data.tienda} pagina="producto" producto={data.producto} />
