<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
	import Plantilla from '#lib/tienda/Plantilla.svelte';

	let { data, form } = $props();
</script>

<Seo
	title="Tu turno · {data.tienda.nombre}"
	description={data.tienda.nombre}
	path="/t/{data.tienda.slug}"
	noindex
/>
<Plantilla
	tienda={data.tienda}
	pagina="turno"
	turno={data.turno}
	aviso={data.pago}
	error={form?.error ?? null}
/>
