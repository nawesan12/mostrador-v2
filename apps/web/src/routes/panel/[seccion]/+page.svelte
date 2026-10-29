<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
	import Pronto from '#lib/panel/Pronto.svelte';

	let { data } = $props();
</script>

<Seo title="{data.seccion} · Panel" description="Panel de Tu Local Web" path="/panel" noindex />
<Pronto seccion={data.seccion} />
