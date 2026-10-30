<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
	import { imagenTienda, textosCompartir } from '#lib/tienda/compartir.ts';
	import Plantilla from '#lib/tienda/Plantilla.svelte';

	let { data } = $props();
</script>

<Seo
	title="Nosotros · {data.tienda.nombre}"
	description={textosCompartir(data.tienda).descripcion}
	path="/t/{data.tienda.slug}/nosotros"
	noindex={false}
	imagen={imagenTienda(data.tienda)}
/>
<Plantilla tienda={data.tienda} pagina="nosotros" />
