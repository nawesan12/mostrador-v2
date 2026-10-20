<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { rutaDe } from './navegacion.ts';
	import { getShell } from './shell.svelte.ts';

	/** Barra de abajo en el celular (handoff): Inicio, Pedidos, Hornero, Productos y Más. */
	let { seccion, pedidos }: { seccion: string; pedidos: number } = $props();
	const shell = getShell();
	const enInicio = $derived(seccion === 'Resumen');
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- rutas del panel armadas con rutaDe -->
<nav class="abajo" aria-label="Accesos">
	<a href="/panel" class:on={enInicio}><Icon icon={I.phHouseSimpleFill} size={23} />Inicio</a>
	<a href={rutaDe('Pedidos')} class:on={seccion === 'Pedidos'}
		><Icon icon={I.phPackageFill} size={23} />Pedidos{#if pedidos}<span class="badge"
				>{pedidos}</span
			>{/if}</a
	>
	<button
		type="button"
		class="hornero"
		aria-label="Hornero"
		title="Hornero llega con la próxima versión del panel"
		disabled><span><img src="/hornero/avatar.png" alt="" /></span></button
	>
	<a href={rutaDe('Productos')} class:on={seccion === 'Productos'}
		><Icon icon={I.phTagFill} size={23} />Productos</a
	>
	<button type="button" class:on={shell.barra} onclick={() => (shell.barra = true)}
		><Icon icon={I.phDotsThreeCircleFill} size={23} />Más</button
	>
</nav>

<!-- eslint-enable svelte/no-navigation-without-resolve -->

<style>
	.abajo {
		display: none;
	}
	@media (max-width: 759px) {
		.abajo {
			position: fixed;
			left: 0;
			right: 0;
			bottom: 0;
			z-index: 62;
			height: calc(76px + env(safe-area-inset-bottom));
			padding: 0 6px calc(10px + env(safe-area-inset-bottom));
			display: flex;
			align-items: flex-end;
			justify-content: space-around;
			background: var(--bga);
			backdrop-filter: blur(16px) saturate(1.2);
			-webkit-backdrop-filter: blur(16px) saturate(1.2);
			border-top: 1px solid var(--ln);
		}
	}
	a,
	button:not(.hornero) {
		position: relative;
		flex: 1;
		max-width: 84px;
		height: 58px;
		border: 0;
		background: transparent;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 3px;
		font-size: 11px;
		font-weight: 600;
		color: var(--mu);
	}
	.on {
		color: #d2582f !important;
	}
	.badge {
		position: absolute;
		top: 4px;
		left: 50%;
		margin-left: 6px;
		min-width: 18px;
		height: 18px;
		padding: 0 5px;
		border-radius: 999px;
		background: #d2582f;
		color: #fff;
		font-size: 10.5px;
		font-weight: 700;
		display: grid;
		place-items: center;
		box-shadow: 0 0 0 2px var(--bg);
	}
	.hornero {
		position: relative;
		width: 64px;
		height: 64px;
		margin-bottom: 6px;
		border-radius: 50%;
		border: 0;
		padding: 0;
		background: linear-gradient(140deg, #e8875f, #c24e25);
		box-shadow:
			0 12px 26px -10px rgba(210, 88, 47, 0.8),
			0 0 0 4px var(--bg);
		transform: translateY(-14px);
	}
	.hornero span {
		position: absolute;
		inset: 4px;
		border-radius: 50%;
		overflow: hidden;
		background: #fbf7f2;
	}
	.hornero img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
</style>
