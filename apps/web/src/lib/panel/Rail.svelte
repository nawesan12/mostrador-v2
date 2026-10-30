<script lang="ts">
	import Marca from '#lib/landing/Marca.svelte';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { AREAS, rutaDe, type Area } from './navegacion.ts';
	import { getShell } from './shell.svelte.ts';

	/** Rail de áreas (handoff): marca, áreas, Tincho y la persona. No se muestra en el celular. */
	let { area, iniciales }: { area: Area; iniciales: string } = $props();
	let menu = $state(false);
	const shell = getShell();
	const areas = $derived(shell.modo === 'simple' ? AREAS.slice(0, 1) : AREAS);
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- rutas del panel armadas con rutaDe -->
<aside class="rail" aria-label="Áreas">
	<a class="marca" href="/panel" aria-label="Inicio del panel">
		<Marca color="#FBF3EA" alto={24} />
	</a>
	{#each areas as a (a.id)}
		{@const on = a.id === area.id}
		<a
			class="area"
			class:on
			href={rutaDe(a.items[0].label)}
			aria-current={on ? 'page' : undefined}
			onclick={() => (shell.barra = matchMedia('(max-width: 1099px)').matches)}
		>
			<Icon icon={on ? a.iconFill : a.icon} size={21} />
			{#if a.punto && !on}<span class="punto"></span>{/if}
			<span class="l">{a.label}</span>
		</a>
	{/each}
	<span class="esp"></span>
	<button
		type="button"
		class="hornero"
		title="Preguntale a Tincho"
		aria-label="Preguntale a Tincho"
		onclick={() => shell.hornero()}
	>
		<span class="pulso"></span>
		<span class="av"><img src="/hornero/avatar.webp" alt="" /></span>
	</button>
	<button
		type="button"
		class="cuenta"
		aria-label="Tu cuenta"
		aria-expanded={menu}
		onclick={() => (menu = !menu)}
	>
		<span class="yo ini">{iniciales}</span>
	</button>
	{#if menu}
		<div class="menu" role="menu">
			<a role="menuitem" href="/panel/configuracion" onclick={() => (menu = false)}
				><Icon icon={I.phGearSix} size={17} />Configuración</a
			>
			<form method="POST" action="/salir" data-sveltekit-reload>
				<button role="menuitem"><Icon icon={I.phSignOutBold} size={17} />Cerrar sesión</button>
			</form>
		</div>
	{/if}
</aside>

<!-- eslint-enable svelte/no-navigation-without-resolve -->

<style>
	.rail {
		width: 76px;
		flex: none;
		position: sticky;
		top: 0;
		height: 100vh;
		height: 100dvh;
		z-index: 76;
		background: #2a140c;
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: 16px 0;
		gap: 4px;
	}
	.marca {
		width: 44px;
		height: 44px;
		border-radius: 13px;
		background: linear-gradient(160deg, #e8875f, #d2582f 55%, #b9481f);
		display: grid;
		place-items: center;
		box-shadow:
			inset 0 1px 0 rgba(255, 255, 255, 0.3),
			0 8px 18px -8px rgba(0, 0, 0, 0.5);
		margin-bottom: 16px;
	}
	.area {
		position: relative;
		width: 62px;
		padding: 9px 0 7px;
		border-radius: 14px;
		color: #c9b4a4;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 4px;
		transition: background 0.2s;
	}
	.area:hover {
		background: rgba(251, 247, 242, 0.06);
		color: #fbf7f2;
	}
	.area.on {
		background: rgba(232, 135, 95, 0.2);
		color: #fbf7f2;
		box-shadow: inset 0 0 0 1px rgba(232, 135, 95, 0.35);
	}
	.area.on :global(svg) {
		color: #f5b796;
	}
	.l {
		font-size: 10.5px;
		font-weight: 500;
	}
	.on .l {
		font-weight: 600;
	}
	.punto {
		position: absolute;
		top: 7px;
		right: 17px;
		width: 7px;
		height: 7px;
		border-radius: 50%;
		background: #e8875f;
		box-shadow: 0 0 0 2px #2a140c;
	}
	.esp {
		flex: 1;
	}
	.hornero {
		position: relative;
		width: 48px;
		height: 48px;
		border-radius: 50%;
		border: 2px solid #e8875f;
		background: #fbf7f2;
		padding: 0;
		overflow: visible;
	}
	.pulso {
		position: absolute;
		inset: -2px;
		border-radius: 50%;
		border: 1px solid #e8875f;
		animation: pulso 2.8s ease-out infinite;
	}
	@keyframes pulso {
		from {
			transform: scale(1);
			opacity: 0.8;
		}
		to {
			transform: scale(1.35);
			opacity: 0;
		}
	}
	@media (prefers-reduced-motion: reduce) {
		.pulso {
			animation: none;
		}
	}
	.av {
		position: absolute;
		inset: 0;
		border-radius: 50%;
		overflow: hidden;
		display: block;
	}
	.av img {
		display: block;
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.yo {
		width: 38px;
		height: 38px;
		border-radius: 50%;
		object-fit: cover;
		margin-top: 12px;
		border: 2px solid rgba(251, 247, 242, 0.25);
	}
	.cuenta {
		margin-top: 12px;
		border-radius: 50%;
	}
	.cuenta .yo {
		margin-top: 0;
	}
	.menu {
		position: absolute;
		left: 70px;
		bottom: 14px;
		z-index: 30;
		min-width: 200px;
		padding: 6px;
		border-radius: 14px;
		background: var(--bg, #fff);
		color: var(--tx, #2a140c);
		box-shadow:
			0 14px 40px -12px rgba(0, 0, 0, 0.35),
			0 0 0 1px rgba(0, 0, 0, 0.06);
	}
	.menu a,
	.menu button {
		width: 100%;
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 12px;
		border-radius: 10px;
		font-size: 14px;
		font-weight: 500;
		color: inherit;
		text-decoration: none;
	}
	.menu a:hover,
	.menu button:hover {
		background: var(--sf, rgba(0, 0, 0, 0.05));
	}
	.ini {
		display: grid;
		place-items: center;
		background: #e8875f;
		color: #2a140c;
		font-size: 13px;
		font-weight: 700;
	}
	@media (max-width: 759px) {
		.rail {
			display: none;
		}
	}
</style>
