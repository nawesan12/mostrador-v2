<script lang="ts">
	import type { ComponentProps } from 'svelte';
	import Ceibo from './plantillas/Ceibo.svelte';

	/**
	 * Dibuja la tienda con su plantilla. Por ahora todas usan Ceibo; las demás se suman de a una con
	 * scripts/plantilla.py (mismo motor, otro marcado).
	 */
	let props: ComponentProps<typeof Ceibo> = $props();
</script>

<Ceibo {...props} />
