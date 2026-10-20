<script lang="ts">
	import { resolve } from '$app/paths';
	import Icon from '#lib/components/Icon.svelte';
	import { anim } from '#lib/anim.ts';
	import { phMinusBold } from '#lib/icons.ts';
	import { getLanding } from './estado.svelte.ts';

	/**
	 * Widget flotante de Hornero (handoff): aparece a los 2,6 s; al minimizarlo no desaparece, queda el
	 * avatar con "¿Te ayudo?" y un aviso.
	 */
	const l = getLanding();
	const registro = `${resolve('/ingresar')}?m=registro`;

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

{#if l.widget}
	<div class="widget" use:anim={'rise'}>
		{#if l.widgetOpen}
			<div class="globo" use:entra role="dialog" aria-label="Hornero">
				<div class="g-cab">
					<span>¡Hola! Soy Hornero</span>
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
					<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- registro es /ingresar con query (resolve adentro) -->
					<a href={registro}>Crear cuenta</a>
				</div>
			</div>
		{:else}
			<div class="ayudo" use:entra>¿Te ayudo?</div>
		{/if}
		<button
			type="button"
			class="avatar"
			title="Hornero"
			aria-label="Hornero"
			aria-expanded={l.widgetOpen}
			onclick={() => (l.widgetOpen = !l.widgetOpen)}
		>
			<span class="pulso" use:anim={{ t: 'pulse', dur: 2600 }}></span>
			<span class="foto"><img src="/hornero/avatar.png" alt="" /></span>
			{#if !l.widgetOpen}<span class="aviso">1</span>{/if}
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
		background: #f4ece2;
		padding: 7px 12px;
		border-radius: 999px;
	}
	.g-acc a.calc {
		color: #fff;
		background: #d2582f;
	}
	.ayudo {
		padding: 10px 14px;
		border-radius: 999px;
		background: #fff;
		box-shadow:
			0 18px 40px -16px rgba(58, 29, 18, 0.45),
			0 0 0 1px rgba(58, 29, 18, 0.06);
		font-size: 13px;
		font-weight: 600;
		white-space: nowrap;
		transform-origin: right center;
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
	.pulso {
		position: absolute;
		inset: 0;
		border-radius: 50%;
		border: 2px solid #d2582f;
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
	.aviso {
		position: absolute;
		top: 0;
		right: 0;
		min-width: 20px;
		height: 20px;
		padding: 0 5px;
		border-radius: 999px;
		background: #2f8f63;
		color: #fff;
		font-size: 11px;
		font-weight: 700;
		display: grid;
		place-items: center;
		box-shadow: 0 0 0 2px #fbf7f2;
	}
</style>
