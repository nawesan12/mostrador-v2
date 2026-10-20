<script lang="ts">
	import '@fontsource-variable/instrument-sans';
	import '../app.css';
	import Toasts from '#lib/components/Toasts.svelte';
	import { setUi, UiState } from '#lib/ui.svelte.ts';
	import { onMount } from 'svelte';

	let { children } = $props();
	setUi(new UiState());
	// Marca de hidratación: los e2e esperan esto antes de hacer clic.
	onMount(() => (document.documentElement.dataset.listo = '1'));
</script>

{@render children()}
<Toasts />
