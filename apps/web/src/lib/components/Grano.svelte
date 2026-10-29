<!--
	Textura de grano sobre toda la página (handoff: assets/grain.png, opacidad .55, multiply).
	No intercepta clics y no se imprime.
-->
<script lang="ts">
	/** `opacidad`: sobre fondo blanco (la landing) va más suave. */
	let { opacidad = 0.55 }: { opacidad?: number } = $props();
</script>

<div class="grano" aria-hidden="true" style:opacity={opacidad}></div>

<style>
	.grano {
		position: fixed;
		inset: 0;
		z-index: 90;
		pointer-events: none;
		background-image: url('/textura/grain.webp');
		mix-blend-mode: multiply;
	}
	@media print {
		.grano {
			display: none;
		}
	}
</style>
