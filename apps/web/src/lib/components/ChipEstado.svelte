<script lang="ts">
	import { ESTADOS, type Estado } from '#lib/marcas.ts';

	let { estado }: { estado: Estado } = $props();
	const [bg, fg] = $derived(ESTADOS[estado]);
</script>

<span class="chip-estado" style:--chip-bg={bg} style:--chip-fg={fg}>{estado}</span>
