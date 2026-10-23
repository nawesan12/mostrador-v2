<script lang="ts">
	import { page } from '$app/state';
	import Seo from '#lib/components/Seo.svelte';
	import Plantilla from '#lib/tienda/Plantilla.svelte';

	let { data } = $props();
	const cat = $derived(page.url.searchParams.get('cat'));
	const categoria = $derived(data.tienda.categorias.some((c) => c.nombre === cat) ? cat : null);
</script>

<Seo
	title="{categoria ?? 'Productos'} · {data.tienda.nombre}"
	description="Todos los productos de {data.tienda.nombre}"
	path="/t/{data.tienda.slug}/catalogo"
/>
<Plantilla
	tienda={data.tienda}
	pagina="catalogo"
	{categoria}
	orden={page.url.searchParams.get('orden') ?? 'vendidos'}
/>
