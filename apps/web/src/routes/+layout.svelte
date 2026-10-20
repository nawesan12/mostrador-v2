<script lang="ts">
	import '@fontsource-variable/instrument-sans';

	let { children } = $props();
</script>

{@render children()}

<style>
	:global(html) {
		background: #fbf7f2;
		color: #2a140c;
		font-family: 'Instrument Sans Variable', system-ui, sans-serif;
		-webkit-font-smoothing: antialiased;
	}
	:global(body) {
		margin: 0;
	}
</style>
