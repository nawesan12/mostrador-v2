<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
	import Plantilla from '#lib/tienda/Plantilla.svelte';

	let { data, form } = $props();
</script>

<Seo
	title="Finalizar compra · {data.tienda.nombre}"
	description={data.tienda.nombre}
	path="/t/{data.tienda.slug}/checkout"
	noindex
/>
<Plantilla tienda={data.tienda} pagina="checkout" error={form?.error ?? null} />
