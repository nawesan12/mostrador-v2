<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
	import FranjaPlan from '#lib/panel/FranjaPlan.svelte';
	import { anim } from '#lib/panel/anim.ts';
	import Informes from './Informes.svelte';

	let { data } = $props();
</script>

{#if data.informes}
	<Informes data={data.informes} />
{:else}
	<Seo
		title="Informes · Panel"
		description="Informes de tu tienda"
		path="/panel/informes"
		noindex
	/>
	<div class="pagina" use:anim={{ t: 'rise' }}>
		<h1>Informes</h1>
		<p class="c-s">Ventas, productos, clientes y margen de los últimos 12&nbsp;meses.</p>
		<FranjaPlan
			funcion="newsletter"
			texto="Con Crece ves cómo crecen tus ventas mes a mes y lo que notó Tincho."
		/>
	</div>
{/if}

<style>
	.pagina {
		max-width: 1100px;
		margin: 0 auto;
	}
	h1 {
		margin: 0;
		font-size: 30px;
		letter-spacing: -0.03em;
		font-weight: 600;
	}
	.c-s {
		margin: 6px 0 18px;
		color: var(--mu);
		font-size: 14px;
	}
</style>
