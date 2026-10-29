<script lang="ts">
	import { resolve } from '$app/paths';
	import Icon from '#lib/components/Icon.svelte';
	import { anim } from '#lib/anim.ts';
	import { phArrowRightBold } from '#lib/icons.ts';

	/**
	 * El llamado a crear la tienda al final de una sección: el botón inflado y la letra chica. Va al
	 * onboarding, que es el registro. `izquierda`: alineado con el texto en vez de centrado.
	 */
	let {
		texto = 'Crear mi tienda gratis',
		izquierda = false
	}: { texto?: string; izquierda?: boolean } = $props();
</script>

<div class="cta" class:izquierda use:anim={'reveal'}>
	<a class="btn-inflado btn-grande" href={resolve('/onboarding')}
		>{texto}<Icon icon={phArrowRightBold} size={16} /></a
	>
	<span class="nota">14 días gratis, sin tarjeta</span>
</div>

<style>
	.cta {
		position: relative;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 10px;
		padding: 8px var(--lado) 72px;
	}
	.cta.izquierda {
		align-items: flex-start;
		padding: 28px 0 0;
	}
	.nota {
		font-size: 13.5px;
		color: #7a6252;
	}
	@media (max-width: 639px) {
		.cta:not(.izquierda) a {
			align-self: stretch;
		}
	}
</style>
