<script lang="ts">
	import { SITE } from '#lib/site.ts';
	/**
	 * Título, descripción, canonical, Open Graph y Twitter de una página pública, más JSON-LD opcional
	 * (mismo componente que Caja). Las imágenes viven en static/og/ en JPG: WhatsApp no muestra la
	 * vista previa de imágenes pesadas.
	 */
	let {
		title,
		description,
		path,
		image = 'general',
		imageAlt = 'Tu Local Web: tu tienda online, WhatsApp, Mercado Libre y el local en un solo panel.',
		noindex = false,
		jsonLd
	}: {
		title: string;
		description: string;
		/** Ruta canónica ("/", "/legales/terminos"). */
		path: string;
		/** Nombre de la imagen en static/og/ (sin extensión). */
		image?:
			| 'home'
			| 'general'
			| 'caja'
			| 'hornero'
			| 'migrar'
			| 'partners'
			| 'plantillas'
			| 'precios'
			| 'seguridad';
		imageAlt?: string;
		noindex?: boolean;
		jsonLd?: Record<string, unknown>;
	} = $props();

	const url = $derived(SITE + path);
	const img = $derived(`${SITE}/og/${image}.jpg`);
	// Un cierre de etiqueta dentro del JSON terminaría el bloque antes de tiempo: se escapa `<`.
	const ld = $derived(
		jsonLd
			? `<script type="application/ld+json">${JSON.stringify(jsonLd).replaceAll('<', '\\u003c')}</` +
					'script>'
			: null
	);
</script>

<svelte:head>
	<title>{title}</title>
	<meta name="description" content={description} />
	{#if noindex}<meta name="robots" content="noindex, nofollow" />{:else}<link
			rel="canonical"
			href={url}
		/>{/if}
	<meta property="og:type" content="website" />
	<meta property="og:site_name" content="Tu Local Web" />
	<meta property="og:locale" content="es_AR" />
	<meta property="og:url" content={url} />
	<meta property="og:title" content={title} />
	<meta property="og:description" content={description} />
	<meta property="og:image" content={img} />
	<meta property="og:image:type" content="image/jpeg" />
	<meta property="og:image:width" content="1200" />
	<meta property="og:image:height" content="630" />
	<meta property="og:image:alt" content={imageAlt} />
	<meta name="twitter:card" content="summary_large_image" />
	<meta name="twitter:title" content={title} />
	<meta name="twitter:description" content={description} />
	<meta name="twitter:image" content={img} />
	<meta name="twitter:image:alt" content={imageAlt} />
	<!-- eslint-disable-next-line svelte/no-at-html-tags -- JSON propio, con `<` escapado -->
	{#if ld}{@html ld}{/if}
</svelte:head>
