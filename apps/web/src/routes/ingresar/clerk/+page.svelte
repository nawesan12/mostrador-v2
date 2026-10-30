<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import Seo from '#lib/components/Seo.svelte';
	import { completar, ErrorClerk } from '#lib/clerk.ts';

	/*
	 * Vuelta de Google o Apple (por Clerk). Termina el ingreso en el navegador y postea el token de
	 * sesión de Clerk al servidor, que lo valida con Go y guarda la sesión de Tu Local Web.
	 */
	let token = $state('');
	let error = $state('');
	let form: HTMLFormElement;

	onMount(async () => {
		try {
			token = await completar(page.url.searchParams.get('rotating_token_nonce') ?? '');
		} catch (e) {
			error = e instanceof ErrorClerk ? e.codigo.replace(/[^a-z_]/g, '').slice(0, 40) : '';
			error ||= 'clerk_no_responde';
		}
		await tick();
		form.requestSubmit();
	});
</script>

<Seo
	title="Ingresando · Tu Local Web"
	description="Entrá a tu panel de Tu Local Web."
	path="/ingresar/clerk"
	noindex
/>

<div class="pag">
	<a href={resolve('/')}><img src="/wordmark-brown.webp" alt="Tu Local Web" class="marca" /></a>
	<div class="centro" role="status" aria-live="polite">
		<img src="/hornero/pensando.webp" alt="" class="ave" />
		<h1>Entrando a tu cuenta</h1>
		<p>Un segundo, estamos terminando de validar tu ingreso.</p>
	</div>
	<form method="POST" bind:this={form} hidden>
		<input type="hidden" name="token" value={token} />
		<input type="hidden" name="error" value={error} />
	</form>
</div>

<style>
	.pag {
		min-height: 100vh;
		display: flex;
		flex-direction: column;
		padding: clamp(20px, 3vw, 36px) clamp(20px, 5vw, 72px);
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
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 10px;
		text-align: center;
		padding: 40px 0;
	}
	.ave {
		width: 120px;
		height: auto;
	}
	h1 {
		margin: 8px 0 0;
		font-size: clamp(28px, 3.2vw, 36px);
		line-height: 1.05;
		letter-spacing: -0.04em;
		font-weight: 600;
		text-wrap: balance;
	}
	p {
		margin: 0;
		max-width: 340px;
		font-size: 15.5px;
		line-height: 1.5;
		color: #5c4536;
		text-wrap: pretty;
	}
</style>
