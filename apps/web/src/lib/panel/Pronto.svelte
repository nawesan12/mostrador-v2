<script lang="ts">
	import Vacio from '#lib/components/Vacio.svelte';

	/** Sección del panel que todavía no está construida (se reemplaza en el PR de cada una). */
	let { seccion }: { seccion: string } = $props();
</script>

<div class="pronto">
	<Vacio
		hornero="pensando"
		titulo={seccion}
		texto="Esta sección se está armando. Llega en las próximas versiones del panel."
	/>
</div>

<style>
	.pronto {
		min-height: 60vh;
		display: grid;
		place-content: center;
	}
	.pronto :global(.vacio) {
		max-width: 420px;
	}
</style>
