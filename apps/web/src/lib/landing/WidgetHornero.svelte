<script lang="ts">
	import { resolve } from '$app/paths';
	import Icon from '#lib/components/Icon.svelte';
	import { anim } from '#lib/anim.ts';
	import { phMinusBold } from '#lib/icons.ts';
	import { banner } from '#lib/consentimiento.svelte.ts';
	import { getLanding } from './estado.svelte.ts';

	/**
	 * Widget flotante de Tincho (handoff): a los 2,6 s aparece el avatar, cerrado; el globo se abre
	 * al tocarlo y al minimizarlo vuelve a quedar el avatar. No se muestra con el menú mobile abierto (taparía los links)
	 * ni mientras el banner de cookies está a la vista.
	 */
	const l = getLanding();
	// Los CTA llevan al onboarding, que es el registro (handoff: Onboarding Mostrador.dc.html).
	const registro = resolve('/onboarding');

	function entra(el: HTMLElement) {
		el.animate(
			[
				{ opacity: 0, transform: 'scale(.9) translateY(6px)' },
				{ opacity: 1, transform: 'none' }
			],
			{ duration: 300, easing: 'cubic-bezier(.2,.7,.1,1)' }
		);
	}
</script>

{#if l.widget && !l.menu && !banner.abierto}
	<div class="widget" use:anim={'rise'}>
		{#if l.widgetOpen}
			<div class="globo" use:entra role="dialog" aria-label="Tincho">
				<div class="g-cab">
					<span>¡Hola! Soy Tincho</span>
					<button
						type="button"
						title="Minimizar"
						aria-label="Minimizar"
						onclick={() => (l.widgetOpen = false)}><Icon icon={phMinusBold} size={14} /></button
					>
				</div>
				<div class="g-txt">¿Te calculo cuánto ahorrás si te pasás desde tu plataforma actual?</div>
				<div class="g-acc">
					<a class="calc" href="#precios">Calcular</a>
					<a href="#hornero">¿Qué hacés?</a>
					<a href={registro}>Crear cuenta</a>
				</div>
			</div>
		{/if}
		<button
			type="button"
			class="avatar"
			title="Tincho"
			aria-label="Tincho"
			aria-expanded={l.widgetOpen}
			onclick={() => (l.widgetOpen = !l.widgetOpen)}
		>
			<span class="foto"><img src="/hornero/avatar.webp" alt="" /></span>
		</button>
	</div>
{/if}

<style>
	.widget {
		position: fixed;
		right: 24px;
		bottom: 24px;
		z-index: 95;
		display: flex;
		align-items: flex-end;
		gap: 12px;
	}
	.globo {
		max-width: 290px;
		padding: 14px 16px;
		border-radius: 20px 20px 4px 20px;
		background: #fff;
		box-shadow:
			0 24px 50px -18px rgba(58, 29, 18, 0.5),
			0 0 0 1px rgba(58, 29, 18, 0.06);
		transform-origin: bottom right;
	}
	.g-cab {
		display: flex;
		justify-content: space-between;
		gap: 10px;
		font-size: 14px;
		font-weight: 600;
	}
	.g-cab button {
		border: 0;
		background: transparent;
		color: #7a6252;
		padding: 0;
		display: grid;
		place-items: center;
	}
	.g-txt {
		margin-top: 4px;
		font-size: 13.5px;
		line-height: 1.45;
		color: #5c4536;
	}
	.g-acc {
		margin-top: 10px;
		display: flex;
		gap: 6px;
		flex-wrap: wrap;
	}
	.g-acc a {
		font-size: 13px;
		font-weight: 600;
		color: #2a140c;
		padding: 7px 12px;
		border-radius: 999px;
		border: 0;
		background: linear-gradient(180deg, #ffffff 0%, #f5efe9 100%);
		box-shadow: var(--inflado-claro);
	}
	.g-acc a.calc {
		color: #fff;
		border: 0;
		background: linear-gradient(180deg, #ee7a4a 0%, #d2582f 55%, #bc4920 100%);
		box-shadow: var(--inflado-naranja);
	}
	.avatar {
		position: relative;
		width: 64px;
		height: 64px;
		flex: none;
		border: 0;
		padding: 0;
		background: transparent;
	}
	.foto {
		position: absolute;
		inset: 0;
		border-radius: 50%;
		overflow: hidden;
		background: #fcede5;
		box-shadow:
			0 0 0 3px #d2582f,
			0 14px 30px -10px rgba(210, 88, 47, 0.7);
	}
	.foto img {
		display: block;
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
</style>
