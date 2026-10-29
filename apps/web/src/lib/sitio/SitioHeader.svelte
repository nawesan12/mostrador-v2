<script lang="ts">
	import { page } from '$app/state';
	import Marca from '#lib/landing/Marca.svelte';

	/**
	 * Header de las páginas públicas (legales, seguridad, contacto, ayuda…), igual en todas las del
	 * handoff. En el celular la nav no entra al lado de los botones: baja a una segunda fila que se
	 * desliza de costado (las mismas piezas, en otro orden).
	 */
	const NAV = [
		['/#producto', 'Producto'],
		['/#plantillas', 'Plantillas'],
		['/#precios', 'Precios'],
		['/migrar', 'Migrar'],
		['/ayuda', 'Ayuda'],
		['/novedades', 'Novedades']
	] as const;
	const activo = (href: string) => !href.startsWith('/#') && page.url.pathname.startsWith(href);
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- anclas de la landing y páginas públicas -->
<header>
	<div class="in">
		<a class="marca" href="/" aria-label="Tu Local Web, inicio"
			><Marca alto={28} /><img
				src="/wordmark-brown.webp"
				alt="Tu Local Web"
				width="175"
				height="20"
			/></a
		>
		<nav aria-label="Secciones">
			{#each NAV as [href, t] (href)}
				<a {href} class:on={activo(href)} aria-current={activo(href) ? 'page' : undefined}>{t}</a>
			{/each}
		</nav>
		<a class="ingresar" href="/ingresar">Ingresar</a>
		<a class="crear" href="/onboarding">Crear mi tienda</a>
	</div>
</header>

<!-- eslint-enable svelte/no-navigation-without-resolve -->

<style>
	header {
		position: sticky;
		top: 0;
		z-index: 40;
		background: rgba(251, 247, 242, 0.88);
		backdrop-filter: blur(14px);
		-webkit-backdrop-filter: blur(14px);
		border-bottom: 1px solid rgba(58, 29, 18, 0.1);
	}
	.in {
		max-width: 1240px;
		margin: 0 auto;
		padding: 0 clamp(16px, 3vw, 28px);
		height: 64px;
		display: flex;
		align-items: center;
		gap: 18px;
	}
	.marca {
		display: flex;
		align-items: center;
		gap: 8px;
		flex: none;
	}
	.marca img {
		height: 20px;
		width: auto;
		display: block;
	}
	nav {
		flex: 1;
		min-width: 0;
		display: flex;
		gap: 2px;
		overflow-x: auto;
		scrollbar-width: none;
		margin-left: 10px;
	}
	nav::-webkit-scrollbar {
		display: none;
	}
	nav a {
		flex: none;
		padding: 8px 12px;
		border-radius: 999px;
		font-size: 14.5px;
		font-weight: 500;
		white-space: nowrap;
		color: #5c4536;
		transition:
			color 0.15s,
			background 0.15s;
	}
	nav a:hover,
	nav a.on {
		color: #2a140c;
		background: rgba(58, 29, 18, 0.05);
	}
	.ingresar {
		flex: none;
		font-size: 14.5px;
		font-weight: 600;
		color: #2a140c;
		white-space: nowrap;
	}
	.crear {
		flex: none;
		display: flex;
		align-items: center;
		gap: 7px;
		padding: 10px 16px;
		border-radius: 999px;
		color: #fff;
		font-size: 14px;
		font-weight: 600;
		white-space: nowrap;
		border: 0;
		background: linear-gradient(180deg, #ee7a4a 0%, #d2582f 55%, #bc4920 100%);
		box-shadow: var(--inflado-naranja);
	}
	.crear:hover {
		color: #fff;
	}
	@media (max-width: 760px) {
		header {
			background: rgba(251, 247, 242, 0.97);
		}
		.in {
			height: auto;
			flex-wrap: wrap;
			gap: 0 14px;
			padding-top: 10px;
		}
		.marca {
			flex: 1;
			height: 44px;
		}
		/* En el celular el isotipo alcanza: así entran Ingresar y Crear mi tienda. */
		.marca img {
			display: none;
		}
		nav {
			order: 3;
			flex: 1 0 100%;
			margin: 4px calc(-1 * clamp(16px, 3vw, 28px)) 0;
			padding: 0 clamp(10px, 3vw, 22px) 8px;
		}
		nav a {
			font-size: 14px;
			padding: 7px 11px;
		}
		.crear {
			padding: 9px 14px;
			font-size: 13.5px;
		}
	}
</style>
