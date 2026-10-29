<script lang="ts">
	import { enhance } from '$app/forms';
	import type { PageProps } from './$types';

	/** Baja de los avisos de /estado (sin diseño en el handoff: piezas del sitio). */
	let { data, form }: PageProps = $props();
	const ok = $derived(!!form && 'ok' in form);
</script>

<svelte:head>
	<title>Dejar de recibir avisos del estado · Tu Local Web</title>
	<meta name="robots" content="noindex" />
</svelte:head>

<div class="ancho">
	<div class="tarjeta">
		<img
			src={ok ? '/hornero/saludo.webp' : '/hornero/desenchufado.webp'}
			alt=""
			width="110"
			height="110"
		/>
		{#if ok}
			<h1>Listo, no te avisamos más.</h1>
			<p>El estado del sistema sigue a la vista cuando lo necesites.</p>
			<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- página pública -->
			<a class="btn" href="/estado">Ver el estado</a>
		{:else}
			<h1>¿Dejás de recibir los avisos?</h1>
			<p>No te vamos a escribir cuando algo falle ni cuando se resuelva.</p>
			<form method="POST" use:enhance>
				<input type="hidden" name="token" value={data.token} />
				<button class="btn" type="submit">Dejar de recibirlos</button>
			</form>
		{/if}
	</div>
</div>

<style>
	.ancho {
		max-width: 1240px;
		margin: 0 auto;
		padding: clamp(48px, 8vw, 104px) clamp(16px, 3vw, 28px);
		display: grid;
		place-items: center;
	}
	.tarjeta {
		width: min(100%, 460px);
		padding: clamp(24px, 4vw, 36px);
		border-radius: 28px;
		background: #fff;
		box-shadow:
			0 2px 4px rgba(58, 29, 18, 0.05),
			0 30px 60px -30px rgba(58, 29, 18, 0.45);
		text-align: center;
	}
	img {
		object-fit: contain;
	}
	h1 {
		margin: 10px 0 0;
		font-size: 26px;
		line-height: 1.1;
		letter-spacing: -0.03em;
		font-weight: 600;
	}
	p {
		margin: 8px 0 0;
		font-size: 15px;
		line-height: 1.55;
		color: #5c4536;
	}
	.btn {
		display: inline-block;
		margin-top: 18px;
		padding: 13px 20px;
		border: 0;
		border-radius: 999px;
		background: #2a140c;
		color: #fff;
		font-size: 14.5px;
		font-weight: 600;
	}
</style>
