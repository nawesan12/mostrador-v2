<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
	import Plantilla from '#lib/tienda/Plantilla.svelte';

	let { data } = $props();
</script>

<Seo
	title="{data.producto.nombre} · {data.tienda.nombre}"
	description={data.producto.descripcion || data.producto.nombre}
	path="/t/{data.tienda.slug}/producto/{data.producto.id}"
/>
<Plantilla tienda={data.tienda} pagina="producto" producto={data.producto} />
