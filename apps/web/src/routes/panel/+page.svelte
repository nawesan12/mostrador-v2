<script lang="ts">
	import { resolve } from '$app/paths';
	import Seo from '#lib/components/Seo.svelte';
	import Vacio from '#lib/components/Vacio.svelte';

	let { data } = $props();
</script>

<Seo title="Panel · {data.tienda.nombre}" description="Panel de Mostrador" path="/panel" noindex />

<!-- Provisorio: el shell del panel (sidebar, topbar, dashboard) llega en el próximo PR del panel. -->
<main>
	<Vacio
		hornero="saludo"
		titulo="Hola, {data.tienda.nombre}"
		texto="Tu demo está lista. El panel completo se está armando y aparece acá muy pronto."
	>
		<a class="btn btn-soft btn-sm" href={resolve('/demo')}>Probar otro rubro</a>
	</Vacio>
</main>

<style>
	main {
		min-height: 100svh;
		display: grid;
		place-content: center;
		padding: 24px 16px;
		background: var(--bg);
	}
	main :global(.vacio) {
		max-width: 420px;
	}
</style>
