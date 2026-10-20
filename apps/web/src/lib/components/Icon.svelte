<script lang="ts">
	import type { IconData } from '#lib/icons.ts';

	/**
	 * Ícono de Phosphor o logo de marca. `icon` es un export de `#lib/icons.ts`
	 * (por ejemplo `phCheckCircleFill`), así el bundle sólo lleva los que se usan.
	 * Sin `label` es decorativo (aria-hidden); con `label`, se anuncia.
	 */
	let {
		icon,
		size = 20,
		label,
		class: cls = '',
		style = ''
	}: { icon: IconData; size?: number; label?: string; class?: string; style?: string } = $props();

	const [w, h, cuerpo] = $derived(icon);
	// Los logos con degradés traen ids fijos: si la primera copia está oculta (p. ej. el hero de
	// escritorio en el celular), url(#id) apunta a ella y el logo no se pinta. Cada instancia los
	// sufija con su propio id.
	const uid = $props.id();
	const body = $derived(
		cuerpo.includes('id="')
			? cuerpo.replace(/(id="|url\(#|href="#)([\w-]+)/g, `$1$2-${uid}`)
			: cuerpo
	);
</script>

<svg
	xmlns="http://www.w3.org/2000/svg"
	viewBox="0 0 {w} {h}"
	width={size}
	height={(size * h) / w}
	class={cls}
	{style}
	fill="currentColor"
	role={label ? 'img' : undefined}
	aria-label={label}
	aria-hidden={label ? undefined : 'true'}
	focusable="false"
>
	<!-- eslint-disable-next-line svelte/no-at-html-tags -- SVG de los sets de Iconify, generado en build -->
	{@html body}
</svg>

<style>
	svg {
		display: block;
		flex: none;
	}
</style>
