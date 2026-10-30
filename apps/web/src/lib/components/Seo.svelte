<script lang="ts">
	import { unsplash } from '#lib/landing/formato.ts';
	import { page } from '$app/state';
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
		foto = null,
		imagen = null,
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
			| 'hornero'
			| 'migrar'
			| 'partners'
			| 'plantillas'
			| 'precios'
			| 'seguridad';
		imageAlt?: string;
		/** Foto propia (id de Unsplash o https), p. ej. la portada de una tienda: reemplaza a `image`. */
		foto?: string | null;
		/**
		 * La imagen al compartir ya armada, 1200 × 630 en JPEG (URL absoluta, p. ej. la de una
		 * tienda: lib/tienda/compartir.ts). Pisa a `foto` y a `image`.
		 */
		imagen?: string | null;
		noindex?: boolean;
		jsonLd?: Record<string, unknown>;
	} = $props();

	// Una tienda con dominio propio: el canonical es el de su dominio, sin /t/<slug>.
	const tienda = $derived(
		page.data.tienda as { slug: string; dominio?: string | null } | undefined
	);
	const url = $derived.by(() => {
		const base = tienda?.dominio ? `/t/${tienda.slug}` : null;
		if (base && (path === base || path.startsWith(`${base}/`)))
			return `https://${tienda!.dominio}${path.slice(base.length) || '/'}`;
		return SITE + path;
	});
	// Con foto propia, recortada a 1200 × 630 como pide la vista previa al compartir.
	const img = $derived(
		imagen ??
			(foto
				? foto.startsWith('/')
					? SITE + foto
					: unsplash(foto, 1200, '&h=630')
				: `${SITE}/og/${image}.jpg`)
	);
	// Las medidas van cuando se saben: la imagen armada o una de static/og.
	const medidas = $derived(!!imagen || !foto);
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
	{#if imagen}<meta property="og:image:secure_url" content={imagen} />{/if}
	{#if medidas}
		<meta property="og:image:type" content="image/jpeg" />
		<meta property="og:image:width" content="1200" />
		<meta property="og:image:height" content="630" />
	{/if}
	<meta property="og:image:alt" content={imageAlt} />
	<meta name="twitter:card" content="summary_large_image" />
	<meta name="twitter:title" content={title} />
	<meta name="twitter:description" content={description} />
	<meta name="twitter:image" content={img} />
	<meta name="twitter:image:alt" content={imageAlt} />
	<!-- eslint-disable-next-line svelte/no-at-html-tags -- JSON propio, con `<` escapado -->
	{#if ld}{@html ld}{/if}
</svelte:head>
