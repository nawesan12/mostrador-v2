<script lang="ts">
	import { ESTADOS, type Estado } from '#lib/marcas.ts';

	/** `texto`: cómo se nombra el estado si no es el de siempre (un retiro: "Listo para retirar"). */
	let { estado, texto }: { estado: Estado; texto?: string } = $props();
	const [bg, fg] = $derived(ESTADOS[estado]);
</script>

<span class="chip-estado" style:--chip-bg={bg} style:--chip-fg={fg}>{texto ?? estado}</span>
