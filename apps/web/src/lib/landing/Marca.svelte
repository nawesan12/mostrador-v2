<script lang="ts">
	/** Isotipo (arco) + wordmark, como el header del handoff. */
	let { color = '#D2582F', alto = 20 }: { color?: string; alto?: number } = $props();
</script>

<svg viewBox="0 0 100 76" style:width="{alto * 1.3}px" style:height="{alto}px" aria-hidden="true">
	<path
		fill={color}
		d="M3 76a3 3 0 0 1-3-3V5a5 5 0 0 1 5-5h90a5 5 0 0 1 5 5v68a3 3 0 0 1-3 3H82a3 3 0 0 1-3-3V21H21v52a3 3 0 0 1-3 3Z"
	></path>
</svg>

<style>
	svg {
		display: block;
	}
</style>
