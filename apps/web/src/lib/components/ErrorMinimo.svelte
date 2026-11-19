<script lang="ts">
	import { page } from '$app/state';

	/**
	 * Un error sin la cabecera ni las llamadas comerciales del sitio: para quien está en su panel o
	 * en el alta (la persona ya es cliente) y para una tienda que no se pudo cargar (quien compra no
	 * tiene por qué ver Tu Local Web). Qué pasó, reintentar y volver.
	 */
	let {
		volver = '/',
		textoVolver = 'Volver al inicio'
	}: { volver?: string | null; textoVolver?: string } = $props();
	const es404 = $derived(page.status === 404);
</script>

<svelte:head>
	<title>{es404 ? 'No encontrado' : 'No pudimos cargar esto'}</title>
	<meta name="robots" content="noindex" />
</svelte:head>

<main class="em">
	<div class="em-caja">
		<h1>{es404 ? 'Esta página no existe' : 'No pudimos cargar esto'}</h1>
		<p>
			{es404
				? 'Puede que el link esté mal escrito o que la hayan movido.'
				: 'Fue algo de nuestro lado. Probá de nuevo en un momento.'}
		</p>
		<div class="em-acc">
			{#if !es404}
				<button type="button" class="em-p" onclick={() => location.reload()}>Reintentar</button>
			{/if}
			{#if volver}
				<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- la vuelta la arma quien llama -->
				<a class={es404 ? 'em-p' : 'em-s'} href={volver}>{textoVolver}</a>
			{/if}
		</div>
	</div>
</main>

<style>
	.em {
		min-height: 100dvh;
		display: grid;
		place-items: center;
		padding: 24px 16px;
		background: #fbf7f2;
		color: #2a140c;
	}
	.em-caja {
		max-width: 420px;
		text-align: center;
	}
	h1 {
		margin: 0;
		font-size: clamp(26px, 4vw, 34px);
		line-height: 1.1;
		letter-spacing: -0.03em;
		font-weight: 600;
		text-wrap: balance;
	}
	p {
		margin: 12px 0 0;
		font-size: 15.5px;
		line-height: 1.5;
		color: #5c4536;
		text-wrap: pretty;
	}
	.em-acc {
		margin-top: 22px;
		display: flex;
		justify-content: center;
		flex-wrap: wrap;
		gap: 10px;
	}
	.em-p,
	.em-s {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		min-height: 46px;
		padding: 0 22px;
		border-radius: 999px;
		font: inherit;
		font-size: 15px;
		font-weight: 600;
		text-decoration: none;
		cursor: pointer;
	}
	.em-p {
		border: 0;
		background: #b0441d;
		color: #fff;
	}
	.em-s {
		border: 1.5px solid rgba(58, 29, 18, 0.16);
		background: none;
		color: #2a140c;
	}
</style>
