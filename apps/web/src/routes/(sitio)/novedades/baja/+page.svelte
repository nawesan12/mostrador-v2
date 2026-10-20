<script lang="ts">
	import { enhance } from '$app/forms';
	import type { PageProps } from './$types';

	/** Baja de las novedades desde el link del mail (sin diseño en el handoff: piezas del sitio). */
	let { data, form }: PageProps = $props();
	const ok = $derived(!!form && 'ok' in form);
</script>

<svelte:head>
	<title>Darte de baja de las novedades · Mostrador</title>
	<meta name="robots" content="noindex" />
</svelte:head>

<div class="ancho">
	<div class="tarjeta">
		<img
			src={ok ? '/hornero/saludo.png' : '/hornero/pensando.png'}
			alt=""
			width="120"
			height="120"
		/>
		{#if ok}
			<h1>Listo, no te escribimos más.</h1>
			<p>Si te arrepentís, te volvés a sumar desde la página de novedades.</p>
			<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- página pública -->
			<a class="sec" href="/novedades">Ver las novedades</a>
		{:else}
			<h1>¿Te damos de baja?</h1>
			<p>Dejás de recibir el resumen mensual de novedades de Mostrador.</p>
			{#if form && 'error' in form}<div class="err" role="alert">{form.error}</div>{/if}
			<form method="POST" use:enhance>
				<input type="hidden" name="token" value={data.token} />
				<button type="submit">Darme de baja</button>
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
	form button,
	.sec {
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
	.err {
		margin-top: 14px;
		padding: 12px 14px;
		border-radius: 14px;
		background: rgba(184, 58, 46, 0.07);
		border: 1px solid rgba(184, 58, 46, 0.28);
		font-size: 14px;
	}
</style>
