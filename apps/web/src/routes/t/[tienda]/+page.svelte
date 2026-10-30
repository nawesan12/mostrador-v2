<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
	import Plantilla from '#lib/tienda/Plantilla.svelte';

	let { data } = $props();
</script>

<Seo
	title={data.tienda.nombre}
	description={data.tienda.portada.lema ||
		data.tienda.portada.titulo ||
		`Comprá en ${data.tienda.nombre}`}
	path="/t/{data.tienda.slug}"
	foto={data.tienda.portada.foto ?? data.tienda.productos.find((p) => p.foto)?.foto ?? null}
	imageAlt={data.tienda.nombre}
/>
<Plantilla tienda={data.tienda} pagina="inicio" proximos={data.proximos} />
