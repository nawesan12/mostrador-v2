<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
	import Plantilla from '#lib/tienda/Plantilla.svelte';

	let { data } = $props();
</script>

<Seo
	title="Nosotros · {data.tienda.nombre}"
	description={data.tienda.nombre}
	path="/t/{data.tienda.slug}/nosotros"
	noindex={false}
/>
<Plantilla tienda={data.tienda} pagina="nosotros" />
