<!--
	Textura de grano sobre toda la página (handoff: assets/grain.png, opacidad .55, multiply).
	No intercepta clics y no se imprime.
-->
<div class="grano" aria-hidden="true"></div>

<style>
	.grano {
		position: fixed;
		inset: 0;
		z-index: 90;
		pointer-events: none;
		background-image: url('/textura/grain.webp');
		opacity: 0.55;
		mix-blend-mode: multiply;
	}
	@media print {
		.grano {
			display: none;
		}
	}
</style>
