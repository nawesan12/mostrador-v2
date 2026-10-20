<script lang="ts">
	import Etiqueta from './Etiqueta.svelte';

	/** Rótulo + título de sección de las páginas del sitio (handoff). */
	let { etiqueta, titulo, color }: { etiqueta: string; titulo: string; color?: string } = $props();
</script>

<Etiqueta texto={etiqueta} {color} />
<h2>{titulo}</h2>

<style>
	h2 {
		margin: 12px 0 0;
		font-size: clamp(32px, 4.2vw, 52px);
		line-height: 1;
		letter-spacing: -0.04em;
		font-weight: 600;
	}
</style>
