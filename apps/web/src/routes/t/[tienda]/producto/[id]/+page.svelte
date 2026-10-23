<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
	import { afterNavigate } from '$app/navigation';
	import Plantilla from '#lib/tienda/Plantilla.svelte';
	import { preparar, verProducto } from '#lib/tienda/medicion.ts';

	let { data } = $props();
	afterNavigate(() => {
		preparar(data.tienda);
		verProducto(data.producto);
	});
</script>

<Seo
	title="{data.producto.nombre} · {data.tienda.nombre}"
	description={data.producto.descripcion || data.producto.nombre}
	path="/t/{data.tienda.slug}/producto/{data.producto.id}"
/>
<Plantilla tienda={data.tienda} pagina="producto" producto={data.producto} />
