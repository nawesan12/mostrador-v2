<script lang="ts">
	import { resolve } from '$app/paths';
	import Icon from '#lib/components/Icon.svelte';
	import { anim } from '#lib/anim.ts';
	import * as I from '#lib/icons.ts';
	import { getLanding } from './estado.svelte.ts';
	import { unsplash } from './formato.ts';

	/*
	 * Hero del manual de marca (Panel de identidad visual, 29/10): el texto a la izquierda y la foto de
	 * un local a la derecha, hasta el borde, con el lado izquierdo curvo e inclinado. En mobile se apila.
	 * Los textos y los botones son los del handoff.
	 */
	// Los CTA llevan al onboarding, que es el registro (handoff: Onboarding Mostrador.dc.html).
	const registro = resolve('/onboarding');
	const FOTO = '1556741533-6e6a62bd8b49';
	const l = getLanding();
</script>

<section class="hero">
	<div class="luz"></div>
	<!-- La forma de la foto, en unidades de su caja: escala con cualquier tamaño. -->
	<svg class="defs" aria-hidden="true" width="0" height="0">
		<clipPath id="hero-forma" clipPathUnits="objectBoundingBox">
			<path d="M.1 0H1v1H.33C.24 1 .19.96 .17.89L.01.13C0 .06.04 0 .1 0Z" />
		</clipPath>
	</svg>
	<div class="in">
		<div class="texto">
			<a class="pildora" href="#hornero" use:anim={'rise'}>
				<span class="avatar"><img src="/hornero/avatar.webp" alt="" /></span>Conocé a Hornero, tu
				asistente con IA<Icon icon={I.phArrowRight} size={14} />
			</a>
			<h1 use:anim={{ t: 'rise', delay: 80 }}>
				Todo tu negocio. <span>Sin complicaciones.</span>
			</h1>
			<p class="bajada" use:anim={{ t: 'rise', delay: 160 }}>
				Tienda online, WhatsApp, Mercado Libre, local y facturación. Para el que recién arranca y
				para el que ya factura millones — sin comisión por venta.
			</p>
			<div class="ctas" use:anim={{ t: 'rise', delay: 240 }}>
				<a class="cta-pri" href={registro}
					>Crear mi tienda gratis <Icon icon={I.phArrowRightBold} size={16} /></a
				>
				<button type="button" class="cta-sec" onclick={() => (l.tplI = 0)}
					><Icon icon={I.phSquaresFour} size={17} />Ver plantillas</button
				>
			</div>
			<div class="checks" use:anim={{ t: 'fade', delay: 380 }}>
				<span
					><Icon icon={I.phCheckCircleFill} size={16} style="color:#237550" />14 días gratis, sin
					tarjeta</span
				>
				<span
					><Icon icon={I.phCheckCircleFill} size={16} style="color:#237550" />0% de comisión por
					venta</span
				>
				<span
					><Icon icon={I.phCheckCircleFill} size={16} style="color:#237550" />Te migramos gratis</span
				>
			</div>
		</div>
	</div>
	<div class="foto" use:anim={{ t: 'fade', delay: 120 }}>
		<img
			src={unsplash(FOTO, 1600)}
			srcset="{unsplash(FOTO, 800)} 800w, {unsplash(FOTO, 1200)} 1200w, {unsplash(
				FOTO,
				1600
			)} 1600w"
			sizes="(max-width: 899px) 100vw, 50vw"
			alt="Una comerciante atiende a una clienta en el mostrador de su local"
			fetchpriority="high"
		/>
	</div>
</section>

<style>
	.hero {
		position: relative;
		overflow: hidden;
	}
	.luz {
		position: absolute;
		inset: 0;
		background: radial-gradient(900px 600px at 15% 10%, rgba(232, 135, 95, 0.16), transparent 65%);
	}
	.defs {
		position: absolute;
	}
	.in {
		position: relative;
		max-width: 1320px;
		margin: 0 auto;
		padding: 0 40px;
		min-height: clamp(580px, 80vh, 760px);
		display: flex;
		align-items: center;
	}
	.texto {
		width: 48%;
		padding: 56px 0 72px;
	}
	.foto {
		position: absolute;
		top: 0;
		bottom: 0;
		left: 51%;
		right: 0;
		clip-path: url(#hero-forma);
	}
	.foto img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		/* La comerciante, a la derecha de la foto, queda siempre a la vista. */
		object-position: 62% 50%;
		display: block;
		/* Más cálida, como la luz de las fotos del manual. */
		filter: sepia(0.16) saturate(1.08);
	}
	.pildora {
		display: inline-flex;
		align-items: center;
		gap: 10px;
		padding: 6px 14px 6px 6px;
		border-radius: 999px;
		background: #fff;
		border: 1px solid rgba(58, 29, 18, 0.1);
		color: #2a140c;
		font-size: 13.5px;
		font-weight: 500;
		white-space: nowrap;
		max-width: 100%;
		box-shadow: 0 4px 14px -6px rgba(58, 29, 18, 0.2);
	}
	.avatar {
		width: 26px;
		height: 26px;
		border-radius: 50%;
		overflow: hidden;
		background: #fcede5;
	}
	.avatar img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	h1 {
		text-wrap: balance;
		margin: 26px 0 0;
		font-size: clamp(46px, 5.1vw, 76px);
		line-height: 0.98;
		letter-spacing: -0.045em;
		font-weight: 600;
	}
	h1 span {
		/* Texto grande: el terracota de la marca ya pasa AA (3:1). */
		color: #d2582f;
	}
	.bajada {
		text-wrap: pretty;
		margin: 24px 0 0;
		max-width: 540px;
		font-size: 19px;
		line-height: 1.55;
		color: #5c4536;
	}
	.ctas {
		margin-top: 32px;
		display: flex;
		gap: 12px;
		flex-wrap: wrap;
	}
	.cta-pri {
		display: flex;
		align-items: center;
		gap: 10px;
		color: #fff;
		background: linear-gradient(180deg, #e06a3e, #d2582f 60%, #bf4b22);
		font-size: 16px;
		font-weight: 600;
		padding: 17px 28px;
		border-radius: 999px;
		white-space: nowrap;
		box-shadow:
			inset 0 1px 0 rgba(255, 255, 255, 0.3),
			inset 0 -2px 0 rgba(0, 0, 0, 0.12),
			0 16px 34px -14px rgba(210, 88, 47, 0.75);
	}
	.cta-pri:hover {
		color: #fff;
		transform: translateY(-1px);
	}
	.cta-sec {
		font: inherit;
		cursor: pointer;
		display: flex;
		align-items: center;
		gap: 10px;
		color: #2a140c;
		background: #fff;
		border: 1px solid rgba(58, 29, 18, 0.14);
		font-size: 16px;
		font-weight: 600;
		padding: 17px 24px;
		border-radius: 999px;
		white-space: nowrap;
		box-shadow: 0 6px 16px -10px rgba(58, 29, 18, 0.3);
	}
	.cta-sec:hover {
		color: #2a140c;
		border-color: rgba(58, 29, 18, 0.3);
	}
	.checks {
		margin-top: 22px;
		display: flex;
		gap: 10px 22px;
		flex-wrap: wrap;
		font-size: 14px;
		color: #7a6252;
	}
	.checks span {
		display: flex;
		align-items: center;
		gap: 6px;
	}

	/* Tablet y mobile: el texto arriba y la foto abajo, hasta el borde derecho. */
	@media (max-width: 899px) {
		.in {
			min-height: 0;
			padding: 0 16px;
		}
		.texto {
			width: 100%;
			padding: 36px 0 32px;
		}
		.foto {
			position: relative;
			left: auto;
			margin-left: 16px;
			height: clamp(300px, 62vw, 460px);
		}
		h1 {
			font-size: clamp(38px, 10.5vw, 56px);
		}
		.bajada {
			font-size: 17px;
		}
		.ctas a {
			flex: 1 1 auto;
			justify-content: center;
		}
	}
</style>
