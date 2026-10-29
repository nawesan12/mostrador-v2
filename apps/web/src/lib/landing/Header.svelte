<script lang="ts">
	import { resolve } from '$app/paths';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import Marca from './Marca.svelte';
	import { getLanding } from './estado.svelte.ts';

	/** Header fijo de la landing y su menú desplegable (tablet/mobile), del handoff. */
	const l = getLanding();
	let scrolled = $state(false);

	const NAV = [
		['#producto', 'Producto'],
		['#plantillas', 'Plantillas'],
		['#integraciones', 'Integraciones'],
		['#precios', 'Precios'],
		['#clientes', 'Clientes'],
		['#seguridad', 'Seguridad']
	] as const;
	const MENU = [
		[
			'#producto',
			I.phSquaresFourFill,
			'Producto',
			'Tienda, WhatsApp, local y facturación',
			'#FCEDE5',
			'#C24E25'
		],
		[
			'#plantillas',
			I.phPaintBrushBroadFill,
			'Plantillas',
			'Más de 15 diseños por rubro',
			'#EFE9F7',
			'#6A4FA0'
		],
		[
			'#hornero',
			I.phSparkleFill,
			'Tincho IA',
			'Tu asistente que vende por vos',
			'#FDF3DD',
			'#9A6A10'
		],
		[
			'#integraciones',
			I.phPlugsConnectedFill,
			'Integraciones',
			'Mercado Pago, ML, ARCA y más',
			'#E6F7FD',
			'#0B8DB8'
		],
		[
			'#precios',
			I.phTagFill,
			'Precios',
			'Calculá cuánto ahorrás',
			'rgba(47,143,99,.13)',
			'#2F8F63'
		],
		['#clientes', I.phHeartFill, 'Clientes', 'Historias reales', '#FCEEF3', '#C2416A'],
		['#seguridad', I.phShieldCheckFill, 'Seguridad', 'Copias, permisos y 2FA', '#F1E9DF', '#6B5244']
	] as const;

	// Los CTA llevan al onboarding, que es el registro (handoff: Onboarding Mostrador.dc.html).
	const registro = resolve('/onboarding');
	const sombra = $derived(scrolled || l.menu);

	function entrar(el: HTMLElement, i: number) {
		el.animate(
			[
				{ opacity: 0, transform: 'translateY(10px)' },
				{ opacity: 1, transform: 'none' }
			],
			{ duration: 420, delay: 40 + i * 35, fill: 'both', easing: 'cubic-bezier(.2,.7,.1,1)' }
		);
	}
</script>

<svelte:window
	onscroll={() => (scrolled = window.scrollY > 12)}
	onkeydown={(e) => e.key === 'Escape' && (l.menu = false)}
	onresize={() => window.innerWidth >= 1024 && (l.menu = false)}
/>

<header class:sombra class:arriba={l.menu}>
	<div class="fila">
		<a href={resolve('/')} class="logo" aria-label="Tu Local Web, inicio">
			<Marca alto={32} />
			<img src="/wordmark-brown.webp" alt="Tu Local Web" />
		</a>
		<nav class="nav" aria-label="Secciones">
			{#each NAV as [h, t] (h)}
				<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- ancla de esta misma página -->
				<a href={h}>{t}</a>
			{/each}
		</nav>
		<a class="ingresar" href={resolve('/ingresar')}>Ingresar</a>
		<span class="esp"></span>
		<a class="cta" href={registro}
			><span class="largo">Crear mi tienda</span><span class="corto">Empezar</span></a
		>
		<button
			type="button"
			class="hamb"
			class:abierto={l.menu}
			aria-label="Menú"
			aria-expanded={l.menu}
			onclick={() => (l.menu = !l.menu)}
		>
			{#key l.menu}
				<span class="hamb-ico"><Icon icon={l.menu ? I.phXBold : I.phListBold} size={19} /></span>
			{/key}
		</button>
	</div>
</header>

{#if l.menu}
	<button type="button" class="fondo" aria-label="Cerrar menú" onclick={() => (l.menu = false)}
	></button>
	<div class="menu" data-lenis-prevent>
		<div class="menu-in">
			<div class="grilla">
				{#each MENU as [h, ico, t, d, bg, fg], i (h)}
					<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- ancla de esta misma página -->
					<a href={h} class="tarjeta" onclick={() => (l.menu = false)} use:entrar={i}>
						<span class="ico" style:background={bg} style:color={fg}
							><Icon icon={ico} size={20} /></span
						>
						<span><span class="t">{t}</span><span class="d">{d}</span></span>
					</a>
				{/each}
			</div>
			<a href="#hornero" class="hornero" onclick={() => (l.menu = false)} use:entrar={MENU.length}>
				<img src="/hornero/saludo.webp" alt="" />
				<span class="h-txt"
					><span class="t">Conocé a Tincho</span><span class="d"
						>Atiende y vende por vos las 24 horas</span
					></span
				>
				<Icon icon={I.phArrowRightBold} size={16} style="position:relative" />
			</a>
			<div class="botones" use:entrar={MENU.length + 1}>
				<a class="b-ing" href={resolve('/ingresar')}>Ingresar</a>
				<a class="b-cta" href={registro}
					>Crear tienda <Icon icon={I.phArrowRightBold} size={14} /></a
				>
			</div>
		</div>
	</div>
{/if}

<style>
	header {
		position: sticky;
		top: 0;
		z-index: 50;
		transition:
			box-shadow 0.3s,
			background 0.3s;
		background: rgba(255, 255, 255, 0.86);
		backdrop-filter: blur(16px) saturate(1.3);
		border-bottom: 1px solid rgba(58, 29, 18, 0.08);
	}
	header.arriba {
		z-index: 51;
	}
	header.sombra {
		box-shadow: 0 10px 30px -18px rgba(42, 20, 12, 0.35);
		background: rgba(255, 255, 255, 0.94);
	}
	.fila {
		max-width: 1240px;
		margin: 0 auto;
		height: 72px;
		padding: 0 var(--lado);
		display: flex;
		align-items: center;
		gap: 32px;
	}
	.logo {
		display: flex;
		align-items: center;
		gap: 8px;
		flex: none;
	}
	.logo img {
		height: 20px;
		width: auto;
		display: block;
	}
	.nav {
		display: flex;
		gap: 4px;
		flex: 1;
		min-width: 0;
		overflow: hidden;
		height: 36px;
		flex-wrap: wrap;
	}
	.nav a {
		color: #5c4536;
		font-size: 14.5px;
		font-weight: 500;
		padding: 8px 12px;
		border-radius: 999px;
		white-space: nowrap;
	}
	.nav a:hover {
		color: #2a140c;
		background: rgba(58, 29, 18, 0.05);
	}
	.ingresar {
		color: #2a140c;
		font-size: 14.5px;
		font-weight: 600;
		flex: none;
		white-space: nowrap;
	}
	.esp {
		display: none;
		flex: 1;
	}
	.cta {
		display: flex;
		align-items: center;
		gap: 8px;
		color: #fff;
		background: #2a140c;
		font-size: 14.5px;
		font-weight: 600;
		padding: 10px 18px;
		border-radius: 999px;
		flex: none;
		white-space: nowrap;
		box-shadow:
			inset 0 1px 0 rgba(255, 255, 255, 0.15),
			0 8px 20px -10px rgba(42, 20, 12, 0.7);
	}
	.cta:hover {
		color: #fff;
		background: #3d2016;
	}
	.corto {
		display: none;
	}
	.hamb {
		display: none;
		width: 44px;
		height: 44px;
		flex: none;
		border-radius: 50%;
		border: 1px solid rgba(58, 29, 18, 0.14);
		background: #fff;
		color: #2a140c;
		transition:
			background 0.25s,
			color 0.25s;
		place-items: center;
		box-shadow: 0 4px 12px -6px rgba(58, 29, 18, 0.25);
	}
	.hamb.abierto {
		background: #2a140c;
		border-color: #2a140c;
		color: #fff;
	}
	.hamb-ico {
		display: grid;
		animation: gira 300ms cubic-bezier(0.2, 0.7, 0.1, 1);
	}
	@keyframes gira {
		from {
			transform: rotate(-90deg) scale(0.6);
			opacity: 0;
		}
	}
	@media (max-width: 1023px) {
		.fila {
			gap: 10px;
		}
		.nav,
		.ingresar {
			display: none;
		}
		.esp {
			display: block;
		}
		.hamb {
			display: grid;
		}
	}
	@media (max-width: 639px) {
		.fila {
			height: 64px;
			padding: 0 14px;
		}
		/* En el celular el isotipo alcanza: así entran Empezar y el menú. */
		.logo img {
			display: none;
		}
		/* Empezar, el botón inflado naranja: el llamado más notorio de la pantalla. */
		.cta {
			min-height: 44px;
			padding: 12px 20px;
			font-size: 15px;
			font-weight: 700;
			background: linear-gradient(180deg, #ee7a4a 0%, #d2582f 55%, #bc4920 100%);
			box-shadow:
				inset 0 2px 0 rgba(255, 255, 255, 0.38),
				inset 0 -3px 0 rgba(120, 38, 12, 0.35),
				0 1px 2px rgba(120, 38, 12, 0.3),
				0 14px 30px -12px rgba(210, 88, 47, 0.85);
		}
		.cta:hover {
			background: linear-gradient(180deg, #ee7a4a 0%, #d2582f 55%, #bc4920 100%);
		}
	}
	@media (max-width: 419px) {
		.largo {
			display: none;
		}
		.corto {
			display: inline;
		}
	}
	/* Menú desplegable */
	.fondo {
		position: fixed;
		inset: 0;
		z-index: 48;
		border: 0;
		background: rgba(42, 20, 12, 0.38);
		backdrop-filter: blur(3px);
		animation: fade 260ms both;
	}
	@keyframes fade {
		from {
			opacity: 0;
		}
	}
	.menu {
		position: fixed;
		left: 0;
		right: 0;
		top: 72px;
		z-index: 49;
		max-height: calc(100dvh - 72px);
		overflow: auto;
		background: #fff;
		border-radius: 0 0 28px 28px;
		box-shadow: 0 40px 80px -30px rgba(42, 20, 12, 0.55);
		transform-origin: top center;
		animation: baja 380ms cubic-bezier(0.16, 1, 0.3, 1) both;
	}
	@keyframes baja {
		from {
			opacity: 0;
			transform: translateY(-16px) scaleY(0.96);
		}
	}
	@media (max-width: 639px) {
		.menu {
			top: 64px;
			max-height: calc(100dvh - 64px);
		}
	}
	.menu-in {
		max-width: 720px;
		margin: 0 auto;
		padding: 14px 14px 18px;
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	.grilla {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
		gap: 8px;
	}
	.tarjeta {
		display: flex;
		flex-direction: column;
		gap: 10px;
		padding: 14px;
		border-radius: 20px;
		background: #fff;
		color: #2a140c;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.06);
	}
	.tarjeta:active {
		transform: scale(0.97);
	}
	.ico {
		width: 40px;
		height: 40px;
		border-radius: 12px;
		display: grid;
		place-items: center;
	}
	.t {
		display: block;
		font-size: 16px;
		font-weight: 600;
		letter-spacing: -0.01em;
	}
	.d {
		display: block;
		font-size: 12.5px;
		line-height: 1.35;
		color: #7a6252;
		margin-top: 2px;
	}
	.hornero {
		position: relative;
		overflow: hidden;
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 12px 14px 12px 8px;
		border-radius: 20px;
		background: linear-gradient(140deg, #d2582f, #b9481f);
		color: #fff;
	}
	.hornero:hover {
		color: #fff;
	}
	.hornero img {
		position: relative;
		width: 64px;
		height: 64px;
		object-fit: contain;
		flex: none;
	}
	.h-txt {
		position: relative;
		flex: 1;
		min-width: 0;
	}
	.h-txt .t {
		font-size: 15px;
	}
	.h-txt .d {
		color: #fce3d6;
	}
	.botones {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 8px;
	}
	.b-ing,
	.b-cta {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		font-size: 15px;
		font-weight: 600;
		padding: 14px;
		border-radius: 999px;
		white-space: nowrap;
	}
	.b-ing {
		color: #2a140c;
		background: #fff;
		border: 1px solid rgba(58, 29, 18, 0.14);
	}
	.b-cta {
		color: #fff;
		background: #2a140c;
	}
	.b-cta:hover {
		color: #fff;
	}
</style>
