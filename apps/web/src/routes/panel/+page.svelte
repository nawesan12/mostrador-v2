<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
	import Pronto from '#lib/panel/Pronto.svelte';

	let { data } = $props();
</script>

<Seo title="Panel · {data.tienda.nombre}" description="Panel de Mostrador" path="/panel" noindex />
<!-- El Resumen (dashboard) llega en el próximo PR. -->
<Pronto seccion="Resumen" />
