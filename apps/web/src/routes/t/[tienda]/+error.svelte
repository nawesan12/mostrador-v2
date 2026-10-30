<script lang="ts">
	import { page } from '$app/state';
	import Plantilla from '#lib/tienda/Plantilla.svelte';
	import type { TiendaPublica } from '#lib/tienda/tipos.ts';
	import ErrorGeneral from '../../+error.svelte';

	/**
	 * Un producto o una página de la tienda que no existe: se ve con la plantilla de la tienda (su
	 * cabecera, su pie y un aviso con su estilo), no con la página de error de Tu Local Web.
	 */
	const tienda = $derived(page.data.tienda as TiendaPublica | undefined);
</script>

<svelte:head>
	{#if tienda && page.status === 404}
		<title>No encontrado · {tienda.nombre}</title>
		<meta name="robots" content="noindex" />
	{/if}
</svelte:head>

{#if tienda && page.status === 404}
	<Plantilla {tienda} pagina="no-existe" />
{:else}
	<ErrorGeneral />
{/if}
