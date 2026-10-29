<script lang="ts">
	import { resolve } from '$app/paths';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { ARTICULOS, categoria, minutos } from '#lib/ayuda/articulos.ts';
	import Lateral from '#lib/ayuda/Lateral.svelte';

	/** Una sección de la ayuda: la portada del handoff en chico y sus artículos. */
	let { data } = $props();
	const c = $derived(categoria(data.slug)!);
	const articulos = $derived(ARTICULOS.filter((a) => data.articulos.includes(a.slug)));
</script>

<Seo
	title="{c.nombre} · Centro de ayuda · Tu Local Web"
	description="Guías de {c.nombre.toLowerCase()} en Tu Local Web."
	path="/ayuda/{c.slug}"
/>

<section class="hero">
	<div class="ancho">
		<a class="volver" href={resolve('/(sitio)/ayuda')}
			><Icon icon={I.phCaretLeftBold} size={13} />Centro de ayuda</a
		>
		<div class="h-in">
			<span class="c-ico" style:background={c.bg}
				><Icon icon={c.icono} size={26} style="color:{c.fg}" /></span
			>
			<h1>{c.nombre}</h1>
		</div>
	</div>
</section>

<div class="ancho">
	<div class="dos">
		<div class="tarjeta">
			{#each articulos as a (a.slug)}
				<a
					href={resolve('/(sitio)/ayuda/[categoria]/[articulo]', {
						categoria: c.slug,
						articulo: a.slug
					})}
					class="art"
				>
					<Icon icon={I.phArticle} size={20} style="color:#b0441d;flex:none" />
					<span class="a-txt"
						><span class="a-t">{a.titulo}</span><span class="a-s"
							>{a.resumen} · {minutos(a)} min de lectura</span
						></span
					>
					<Icon icon={I.phCaretRightBold} size={14} style="color:#9A8576" />
				</a>
			{/each}
		</div>
		<Lateral />
	</div>
</div>

<style>
	.ancho {
		max-width: 1240px;
		margin: 0 auto;
		padding: 0 clamp(16px, 3vw, 28px);
	}
	.hero {
		background: #f4ece2 url('/textura/arch-pattern-light.svg');
		padding: 24px 0 clamp(32px, 5vw, 56px);
	}
	.volver {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		font-size: 14px;
		font-weight: 600;
		color: #5c4536;
		text-decoration: none;
	}
	.h-in {
		margin-top: 22px;
		display: flex;
		align-items: center;
		gap: 16px;
		flex-wrap: wrap;
	}
	.c-ico {
		width: 58px;
		height: 58px;
		border-radius: 18px;
		display: grid;
		place-items: center;
	}
	h1 {
		margin: 0;
		font-size: clamp(32px, 5vw, 56px);
		line-height: 1.05;
		letter-spacing: -0.045em;
		font-weight: 600;
	}
	.dos {
		padding: clamp(32px, 5vw, 56px) 0 clamp(48px, 7vw, 80px);
		display: grid;
		grid-template-columns: minmax(0, 1fr) minmax(0, 380px);
		gap: clamp(18px, 3vw, 36px);
		align-items: start;
	}
	.tarjeta {
		border-radius: 24px;
		background: #fff;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.06);
		overflow: hidden;
	}
	.art {
		display: flex;
		align-items: center;
		gap: 14px;
		padding: 18px;
		text-decoration: none;
		color: #2a140c;
		border-top: 1px solid rgba(58, 29, 18, 0.07);
	}
	.art:first-child {
		border-top: 0;
	}
	.a-txt {
		flex: 1;
		min-width: 0;
	}
	.a-t {
		display: block;
		font-size: 16px;
		font-weight: 600;
	}
	.a-s {
		display: block;
		font-size: 13px;
		color: #7a6252;
		margin-top: 3px;
		line-height: 1.4;
	}
	@media (max-width: 860px) {
		.dos {
			grid-template-columns: minmax(0, 1fr);
		}
	}
</style>
