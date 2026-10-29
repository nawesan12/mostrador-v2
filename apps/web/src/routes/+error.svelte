<script lang="ts">
	import { page } from '$app/state';
	import CookieBanner from '#lib/components/CookieBanner.svelte';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import SitioFooter from '#lib/sitio/SitioFooter.svelte';
	import SitioHeader from '#lib/sitio/SitioHeader.svelte';

	/**
	 * Página de error de toda la app (design/Mostrador 404.dc.html). El 404 es el del handoff; para
	 * otros errores se usa la misma pantalla con otro texto.
	 */
	const es404 = $derived(page.status === 404);
	const ACCESOS = [
		['/', I.phHouseFill, 'Inicio'],
		['/ingresar', I.phSignInBold, 'Ingresar a mi panel'],
		['/ayuda', I.phLifebuoyFill, 'Centro de ayuda'],
		['/estado', I.phPulseFill, 'Estado del sistema']
	] as const;
</script>

<svelte:head>
	<title>{es404 ? 'Página no encontrada' : 'Algo salió mal'} · Tu Local Web</title>
	<meta name="robots" content="noindex" />
</svelte:head>

<!-- eslint-disable svelte/no-navigation-without-resolve -- rutas públicas fijas -->
<div class="sitio">
	<SitioHeader />
	<main>
		<div class="ancho">
			<div class="in">
				<div class="txt">
					<div class="codigo">{page.status}</div>
					<h1>{es404 ? 'El horno está vacío.' : 'Algo se nos quemó.'}</h1>
					<p>
						{#if es404}
							Buscamos por todos lados y esta página no existe, o la movimos. Probá con alguna de
							estas:
						{:else}
							No pudimos mostrar esta página. Ya nos avisaron; probá de nuevo en un rato o andá a
							alguna de estas:
						{/if}
					</p>
					<form class="buscar" action="/ayuda" method="GET" role="search">
						<Icon icon={I.phMagnifyingGlassBold} size={18} style="color:#7A6252;flex:none" />
						<input
							name="q"
							placeholder="¿Qué estabas buscando?"
							aria-label="¿Qué estabas buscando?"
						/>
						<button type="submit">Buscar</button>
					</form>
					<div class="accesos">
						{#each ACCESOS as [href, ico, t] (href)}
							<a {href}><Icon icon={ico} size={19} style="color:#b0441d" />{t}</a>
						{/each}
					</div>
				</div>
				<img src="/hornero/horno-vacio.webp" alt="" width="400" height="400" />
			</div>
		</div>
	</main>
	<SitioFooter />
</div>
<CookieBanner />

<style>
	.sitio {
		min-height: 100vh;
		display: flex;
		flex-direction: column;
		background: #fbf7f2;
		color: #2a140c;
	}
	main {
		flex: 1;
	}
	.ancho {
		max-width: 1240px;
		margin: 0 auto;
		padding: 0 clamp(16px, 3vw, 28px);
	}
	.in {
		padding: clamp(56px, 9vw, 120px) 0;
		display: flex;
		align-items: center;
		gap: clamp(24px, 5vw, 72px);
		flex-wrap: wrap-reverse;
		justify-content: center;
	}
	.txt {
		flex: 1 1 440px;
		min-width: 0;
		max-width: 560px;
	}
	.codigo {
		font-size: clamp(90px, 14vw, 190px);
		line-height: 0.8;
		letter-spacing: -0.07em;
		font-weight: 600;
		color: #b0441d;
	}
	h1 {
		margin: 18px 0 0;
		font-size: clamp(32px, 4vw, 50px);
		line-height: 1.02;
		letter-spacing: -0.04em;
		font-weight: 600;
	}
	p {
		margin: 12px 0 0;
		font-size: 17px;
		line-height: 1.55;
		color: #5c4536;
	}
	.buscar {
		margin-top: 20px;
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 7px 7px 7px 16px;
		border-radius: 999px;
		background: #fff;
		box-shadow: 0 16px 40px -24px rgba(58, 29, 18, 0.5);
	}
	.buscar input {
		flex: 1;
		min-width: 0;
		border: 0;
		outline: 0;
		background: transparent;
		font-size: 15.5px;
		font-weight: 500;
		color: #2a140c;
		padding: 8px 0;
	}
	.buscar button {
		padding: 11px 18px;
		border: 0;
		border-radius: 999px;
		background: #2a140c;
		color: #fff;
		font-size: 14px;
		font-weight: 600;
	}
	.accesos {
		margin-top: 16px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(170px, 1fr));
		gap: 8px;
	}
	.accesos a {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 14px;
		border-radius: 18px;
		background: #fff;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.06);
		color: #2a140c;
		font-size: 14.5px;
		font-weight: 600;
	}
	.in > img {
		width: clamp(220px, 30vw, 400px);
		height: auto;
		flex: none;
	}
	@media (max-width: 640px) {
		.in > img {
			width: 180px;
		}
		.accesos {
			grid-template-columns: 1fr 1fr;
		}
	}
</style>
