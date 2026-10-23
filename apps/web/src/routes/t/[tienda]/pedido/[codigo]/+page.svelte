<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
	import Plantilla from '#lib/tienda/Plantilla.svelte';

	let { data } = $props();
</script>

<Seo
	title="Pedido #{data.pedido.numero} · {data.tienda.nombre}"
	description={data.tienda.nombre}
	path="/t/{data.tienda.slug}"
	noindex
/>
<Plantilla tienda={data.tienda} pagina="listo" pedido={data.pedido} />
