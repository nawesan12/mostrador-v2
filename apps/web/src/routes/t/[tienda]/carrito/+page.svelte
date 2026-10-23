<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
	import Plantilla from '#lib/tienda/Plantilla.svelte';

	let { data } = $props();
</script>

<Seo
	title="Carrito · {data.tienda.nombre}"
	description={data.tienda.nombre}
	path="/t/{data.tienda.slug}/carrito"
	noindex={true}
/>
<Plantilla tienda={data.tienda} pagina="carrito" />
