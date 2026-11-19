<script lang="ts">
	import { page } from '$app/state';
	import ErrorMinimo from '#lib/components/ErrorMinimo.svelte';
	import Plantilla from '#lib/tienda/Plantilla.svelte';
	import type { TiendaPublica } from '#lib/tienda/tipos.ts';

	/**
	 * Un error dentro de una tienda se ve con la plantilla de la tienda (su cabecera, su pie y un
	 * aviso con su estilo), nunca con la página de Tu Local Web: el 404 con «No encontramos esto» y
	 * cualquier otro con «No pudimos cargar esto» y Reintentar (el carrito sigue guardado). Sirve para
	 * los 17 presets. Si falló la tienda misma (el layout), una versión neutra sin la marca del sitio.
	 */
	const tienda = $derived(page.data.tienda as TiendaPublica | undefined);
</script>

<svelte:head>
	{#if tienda}
		<title
			>{page.status === 404 ? 'No encontrado' : 'No pudimos cargar esto'} · {tienda.nombre}</title
		>
		<meta name="robots" content="noindex" />
	{/if}
</svelte:head>

{#if tienda}
	<Plantilla {tienda} pagina={page.status === 404 ? 'no-existe' : 'error'} />
{:else}
	<ErrorMinimo volver={null} />
{/if}
