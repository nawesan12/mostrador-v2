<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
	import { onMount } from 'svelte';
	import Plantilla from '#lib/tienda/Plantilla.svelte';
	import { compra, preparar } from '#lib/tienda/medicion.ts';

	let { data } = $props();
	onMount(() => {
		preparar(data.tienda);
		compra(data.pedido);
	});
</script>

<Seo
	title="Pedido #{data.pedido.numero} · {data.tienda.nombre}"
	description={data.tienda.nombre}
	path="/t/{data.tienda.slug}"
	noindex
/>
<Plantilla tienda={data.tienda} pagina="listo" pedido={data.pedido} />
