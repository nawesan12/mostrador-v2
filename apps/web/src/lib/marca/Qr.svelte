<script lang="ts">
	import { renderSVG } from 'uqr';

	/** QR de las piezas de marca, con los colores del handoff (tostado sobre crema) y la zona clara de 4 módulos que pide el estándar (varios van sobre fondo oscuro). */
	let { texto, alt, style = '' }: { texto: string; alt: string; style?: string } = $props();

	const src = $derived(
		'data:image/svg+xml;charset=utf-8,' +
			encodeURIComponent(
				renderSVG(texto, { border: 4, ecc: 'M', whiteColor: '#FBF7F2', blackColor: '#2A140C' })
			)
	);
</script>

<img {src} {alt} {style} />
