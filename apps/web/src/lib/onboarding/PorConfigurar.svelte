<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import { phArrowUpRightBold } from '#lib/icons.ts';
	import Logo from './Logo.svelte';

	/**
	 * Con una cuenta real, los pasos de Canales y de Cobros no simulan conexiones: muestran lo que
	 * conviene dejar listo y llevan a la pantalla del panel donde se hace de verdad. Se abre en otra
	 * pestaña para no perder el onboarding (igual queda guardado).
	 */
	type Item = { k: string; t: string; d: string; href: string };
	let { items, nota }: { items: Item[]; nota: string } = $props();
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- secciones del panel armadas en datos.ts -->
<div class="lista">
	{#each items as it (it.k)}
		<a class="item" href={it.href} target="_blank" rel="noopener">
			<Logo k={it.k} tam={36} icono={18} radio={11} />
			<span class="i-txt"><span class="i-t">{it.t}</span><span class="i-d">{it.d}</span></span>
			<span class="ir">Configurar<Icon icon={phArrowUpRightBold} size={13} /></span>
		</a>
	{/each}
</div>
<!-- eslint-enable svelte/no-navigation-without-resolve -->
<p class="nota">{nota}</p>

<style>
	.lista {
		margin-top: 22px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 340px), 1fr));
		gap: 10px;
	}
	.item {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 14px 16px;
		border-radius: 18px;
		background: #fff;
		border: 1px solid rgba(58, 29, 18, 0.1);
		color: #2a140c;
		transition: border-color 0.15s;
	}
	.item:hover {
		color: #2a140c;
		border-color: rgba(210, 88, 47, 0.45);
	}
	.i-txt {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 2px;
	}
	.i-t {
		font-size: 15px;
		font-weight: 600;
	}
	.i-d {
		font-size: 13px;
		color: #7a6252;
		line-height: 1.4;
	}
	.ir {
		flex: none;
		display: inline-flex;
		align-items: center;
		gap: 5px;
		font-size: 13.5px;
		font-weight: 600;
		color: #b0441d;
	}
	.nota {
		margin: 16px 0 0;
		font-size: 13.5px;
		color: #7a6252;
	}
</style>
