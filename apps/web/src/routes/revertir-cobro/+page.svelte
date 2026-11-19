<script lang="ts">
	import { enhance } from '$app/forms';
	import { resolve } from '$app/paths';
	import Seo from '#lib/components/Seo.svelte';
	import type { PageProps } from './$types';

	/** «No fui yo»: vuelve atrás el cambio de adónde llega la plata y cierra todas las sesiones. */
	let { data, form }: PageProps = $props();
	let enviando = $state(false);
</script>

<Seo
	title="Volver atrás el cambio · Tu Local Web"
	description="Volvé atrás un cambio de tus datos de cobro."
	path="/revertir-cobro"
	noindex
/>

<div class="pag">
	<div class="arriba">
		<a href={resolve('/')}><img src="/wordmark-brown.webp" alt="Tu Local Web" class="marca" /></a>
	</div>
	<div class="centro">
		<div class="caja col">
			{#if form && 'hecho' in form && form.hecho}
				<h1>Listo, lo volvimos atrás</h1>
				<p class="bajada">
					{form.hecho === 'mercadopago'
						? 'Desconectamos la cuenta de Mercado Pago nueva y cerramos todas las sesiones de tu equipo. Entrá y conectá la tuya.'
						: 'Volvieron tus datos para transferencias de antes y cerramos todas las sesiones de tu equipo.'}
				</p>
				<p class="bajada">Entrá y cambiá tu contraseña: alguien la puede tener.</p>
				<a class="cta" href={resolve('/ingresar')}>Ingresar</a>
			{:else if !data.token}
				<h1>Este link ya no sirve</h1>
				<p class="bajada">
					Abrilo desde el último mail que te mandamos, o entrá a tu panel y revisá Configuración ›
					Negocio.
				</p>
				<a class="cta" href={resolve('/ingresar')}>Ir a ingresar</a>
			{:else}
				<h1>¿No fuiste vos?</h1>
				<p class="bajada">
					Volvemos atrás el cambio de adónde te llega la plata y cerramos todas las sesiones de tu
					equipo, también la tuya.
				</p>
				{#if form && 'mensaje' in form}<div class="err-form" role="alert">{form.mensaje}</div>{/if}
				<form
					method="POST"
					use:enhance={() => {
						enviando = true;
						return async ({ update }) => {
							await update({ navigate: false });
							enviando = false;
						};
					}}
				>
					<input type="hidden" name="token" value={data.token} />
					<button class="cta" disabled={enviando}
						>{enviando ? 'Volviendo atrás…' : 'Volver atrás y cerrar sesiones'}</button
					>
				</form>
				<p class="chico">Si fuiste vos, no hace falta que hagas nada.</p>
			{/if}
		</div>
	</div>
</div>

<style>
	.pag {
		min-height: 100dvh;
		display: flex;
		flex-direction: column;
		padding: clamp(20px, 3vw, 36px) clamp(16px, 5vw, 72px);
		color: #2a140c;
		background: #fbf7f2;
	}
	.marca {
		height: 21px;
		width: auto;
		display: block;
	}
	.centro {
		flex: 1;
		display: flex;
		align-items: center;
		padding: clamp(28px, 5vw, 56px) 0;
	}
	.caja {
		width: 100%;
		max-width: 440px;
		margin: 0 auto;
	}
	.col {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
	h1 {
		margin: 0;
		font-size: clamp(30px, 3.6vw, 40px);
		line-height: 1.04;
		letter-spacing: -0.04em;
		font-weight: 600;
		text-wrap: balance;
	}
	.bajada {
		margin: 0;
		font-size: 15.5px;
		line-height: 1.5;
		color: #5c4536;
		text-wrap: pretty;
	}
	.chico {
		margin: 0;
		font-size: 13.5px;
		color: #6b5244;
	}
	.err-form {
		padding: 12px 14px;
		border-radius: 14px;
		background: rgba(184, 58, 46, 0.08);
		color: #8f2a20;
		font-size: 14px;
		line-height: 1.45;
	}
	.cta {
		display: inline-flex;
		justify-content: center;
		align-items: center;
		width: 100%;
		min-height: 50px;
		padding: 14px 20px;
		border: 0;
		border-radius: 999px;
		background: #b0441d;
		color: #fff;
		font: inherit;
		font-size: 15.5px;
		font-weight: 600;
		text-decoration: none;
		cursor: pointer;
	}
	.cta:disabled {
		opacity: 0.6;
	}
</style>
