<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
	import Plantilla from '#lib/tienda/Plantilla.svelte';

	let { data } = $props();
</script>

<Seo
	title={data.tienda.nombre}
	description={data.tienda.portada.titulo}
	path="/t/{data.tienda.slug}"
/>
<Plantilla tienda={data.tienda} pagina="inicio" />
