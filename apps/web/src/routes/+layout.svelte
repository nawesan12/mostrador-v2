<script lang="ts">
	import '@fontsource-variable/instrument-sans';
	import '../app.css';
	import Toasts from '#lib/components/Toasts.svelte';
	import { setUi, UiState } from '#lib/ui.svelte.ts';

	let { children } = $props();
	setUi(new UiState());
</script>

{@render children()}
<Toasts />
