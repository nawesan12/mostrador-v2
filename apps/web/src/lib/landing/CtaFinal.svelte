<script lang="ts">
	import { resolve } from '$app/paths';
	import Icon from '#lib/components/Icon.svelte';
	import { anim } from '#lib/anim.ts';
	import { phArrowRightBold } from '#lib/icons.ts';
	import { unsplash } from './formato.ts';

	/** "Abrí Tu Local Web hoy." — CTA final de la landing. */
	// Los CTA llevan al onboarding, que es el registro (handoff: Onboarding Mostrador.dc.html).
	const registro = resolve('/onboarding');
</script>

<section>
	<div class="caja">
		<img
			src={unsplash('1556740758-90de374c12ad', 1800)}
			data-pxi
			alt="Dueñas de un local"
			loading="lazy"
		/>
		<div class="sombra"></div>
		<div class="txt">
			<h2 use:anim={'reveal'}>Abrí Tu&nbsp;Local&nbsp;Web&nbsp;hoy.</h2>
			<p use:anim={'reveal'}>
				14 días gratis, sin tarjeta. Si venís de otra plataforma, la migración corre por nuestra
				cuenta.
			</p>
			<div class="ctas" use:anim={'reveal'}>
				<a class="pri" href={registro}
					>Crear mi tienda gratis <Icon icon={phArrowRightBold} size={16} /></a
				>
				<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- /contacto llega con las páginas públicas -->
				<a class="sec" href="/contacto">Hablar con ventas</a>
			</div>
		</div>
	</div>
</section>

<style>
	section {
		padding: 0 var(--lado) 100px;
	}
	.caja {
		position: relative;
		max-width: 1240px;
		margin: 0 auto;
		border-radius: 32px;
		overflow: hidden;
		min-height: 480px;
		display: flex;
		align-items: center;
		box-shadow: 0 40px 80px -40px rgba(42, 20, 12, 0.6);
	}
	img {
		position: absolute;
		inset: 0;
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.sombra {
		position: absolute;
		inset: 0;
		background: linear-gradient(
			90deg,
			rgba(42, 20, 12, 0.92) 0%,
			rgba(42, 20, 12, 0.7) 45%,
			rgba(42, 20, 12, 0.1)
		);
	}
	.txt {
		position: relative;
		padding: 64px;
		color: #fff;
		max-width: 640px;
	}
	h2 {
		margin: 0;
		font-size: clamp(40px, 5.4vw, 72px);
		line-height: 0.98;
		letter-spacing: -0.045em;
		font-weight: 600;
		text-wrap: balance;
	}
	p {
		margin: 20px 0 0;
		font-size: 18px;
		line-height: 1.55;
		color: #eadbcf;
	}
	.ctas {
		margin-top: 32px;
		display: flex;
		gap: 12px;
		flex-wrap: wrap;
	}
	.ctas a {
		display: flex;
		align-items: center;
		gap: 10px;
		color: #fff;
		font-size: 16px;
		font-weight: 600;
		border-radius: 999px;
	}
	.pri {
		padding: 16px 26px;
		border: 0;
		background: linear-gradient(180deg, #ee7a4a 0%, #d2582f 55%, #bc4920 100%);
		box-shadow: var(--inflado-naranja);
	}
	.sec {
		padding: 16px 24px;
		border: 0;
		background: linear-gradient(180deg, rgba(255, 255, 255, 0.18), rgba(255, 255, 255, 0.06));
		box-shadow: var(--inflado-vidrio);
	}
	.ctas a:hover {
		color: #fff;
	}
	@media (max-width: 640px) {
		section {
			padding-bottom: 80px;
		}
		.caja {
			min-height: 440px;
			border-radius: 26px;
		}
		.sombra {
			background: linear-gradient(
				180deg,
				rgba(42, 20, 12, 0.88) 0%,
				rgba(42, 20, 12, 0.72) 60%,
				rgba(42, 20, 12, 0.5)
			);
		}
		.txt {
			padding: 36px 24px;
		}
		p {
			font-size: 16.5px;
		}
		.ctas a {
			width: 100%;
			justify-content: center;
		}
	}
</style>
