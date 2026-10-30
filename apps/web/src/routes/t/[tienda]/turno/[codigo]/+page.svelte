<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
	import { imagenTienda, textosCompartir } from '#lib/tienda/compartir.ts';
	import Plantilla from '#lib/tienda/Plantilla.svelte';

	let { data, form } = $props();
</script>

<Seo
	title="Tu turno · {data.tienda.nombre}"
	description={textosCompartir(data.tienda).descripcion}
	path="/t/{data.tienda.slug}"
	noindex
	imagen={imagenTienda(data.tienda)}
/>
<Plantilla
	tienda={data.tienda}
	pagina="turno"
	turno={data.turno}
	aviso={data.pago}
	error={form?.error ?? null}
/>
