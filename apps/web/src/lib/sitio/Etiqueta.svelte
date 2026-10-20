<script lang="ts">
	/** El rótulo con punto que encabeza las secciones del sitio (handoff). */
	let { texto, color = '#D2582F' }: { texto: string; color?: string } = $props();
</script>

<div class="etiq" style:color><span></span>{texto}</div>

<style>
	.etiq {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 14px;
		font-weight: 600;
	}
	span {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		background: currentColor;
	}
</style>
