<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
	import Plantilla from '#lib/tienda/Plantilla.svelte';
	import { imagenTienda, textosCompartir } from '#lib/tienda/compartir.ts';

	let { data } = $props();
	const textos = $derived(textosCompartir(data.tienda));
</script>

<Seo
	title={textos.titulo}
	description={textos.descripcion}
	path="/t/{data.tienda.slug}"
	imagen={imagenTienda(data.tienda)}
	imageAlt={data.tienda.nombre}
/>
<Plantilla tienda={data.tienda} pagina="inicio" proximos={data.proximos} />
