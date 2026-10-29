<script lang="ts">
	/**
	 * Isotipo de Tu Local Web: la casita con la bolsa. A color por defecto; con `color`, de un solo tono
	 * (para fondos oscuros o de color, como en el manual de marca). Es el mismo dibujo que /mark.svg.
	 */
	let { color, alto = 20 }: { color?: string; alto?: number } = $props();
	const techo = $derived(color ?? '#D9652F');
	const manija = $derived(color ?? '#4B2B16');
</script>

<svg viewBox="0 0 100 100" style:width="{alto}px" style:height="{alto}px" aria-hidden="true">
	<path
		d="M11.5 48 41 22q9-7.5 18 0L88.5 48"
		fill="none"
		stroke={techo}
		stroke-width="19"
		stroke-linecap="round"
		stroke-linejoin="round"
	/>
	<rect
		x="22"
		y="58"
		width="56"
		height="30"
		rx="7.5"
		fill={color ? 'none' : '#F8E6D4'}
		stroke={techo}
		stroke-width="6.5"
	/>
	<path
		d="M38 57v-4.5a12 12 0 0 1 24 0V57"
		fill="none"
		stroke={manija}
		stroke-width="6"
		stroke-linecap="round"
	/>
</svg>

<style>
	svg {
		display: block;
		flex: none;
	}
</style>
