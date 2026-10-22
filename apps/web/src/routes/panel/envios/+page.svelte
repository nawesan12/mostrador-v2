<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
	import MotorLista from '#lib/panel/listas/MotorLista.svelte';
	import { configEnvios } from '#lib/panel/listas/envios.ts';

	/** Envíos (PanelListas): etiquetas, seguimiento y reclamos. */
	let { data } = $props();
	const config = configEnvios();
</script>

<Seo title="Envíos · Panel" description="Envíos de tu tienda" path="/panel/envios" noindex />

<MotorLista {config} lista={data.lista} />
