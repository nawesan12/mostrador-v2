<script lang="ts">
	import { resolve } from '$app/paths';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { ARTICULOS, categoria, delaCategoria, minutos } from '#lib/ayuda/articulos.ts';
	import Lateral from '#lib/ayuda/Lateral.svelte';

	/** Un artículo de la ayuda. El handoff no lo diseña: se arma con la portada y las tarjetas. */
	let { data } = $props();
	const a = $derived(ARTICULOS.find((x) => x.slug === data.slug)!);
	const c = $derived(categoria(a.categoria)!);
	const otros = $derived(delaCategoria(a.categoria).filter((x) => x.slug !== a.slug));
	let voto = $state<'si' | 'no' | null>(null);
</script>

<Seo
	title="{a.titulo} · Ayuda · Mostrador"
	description={a.resumen}
	path="/ayuda/{a.categoria}/{a.slug}"
/>

<section class="hero">
	<div class="ancho">
		<nav class="migas" aria-label="Ruta">
			<a href={resolve('/(sitio)/ayuda')}>Ayuda</a><Icon icon={I.phCaretRightBold} size={11} /><a
				href={resolve('/(sitio)/ayuda/[categoria]', { categoria: c.slug })}>{c.nombre}</a
			>
		</nav>
		<h1>{a.titulo}</h1>
		<div class="sub">{a.resumen} · {minutos(a)} min de lectura</div>
	</div>
</section>

<div class="ancho">
	<div class="dos">
		<div>
			<article class="tarjeta cuerpo">
				{#each a.cuerpo as b, i (i)}
					{#if b.t === 'p'}
						<p>{b.x}</p>
					{:else if b.t === 'h'}
						<h2>{b.x}</h2>
					{:else if b.t === 'nota'}
						<div class="nota">
							<Icon icon={I.phSparkleFill} size={16} style="color:#b0441d;flex:none" /><span
								>{b.x}</span
							>
						</div>
					{:else if b.t === 'pasos'}
						<ol>
							{#each b.x as x (x)}<li>{x}</li>{/each}
						</ol>
					{:else}
						<ul>
							{#each b.x as x (x)}<li>{x}</li>{/each}
						</ul>
					{/if}
				{/each}
				<div class="voto">
					{#if voto}
						<span
							>{voto === 'si'
								? '¡Gracias! Nos alegra que te sirva.'
								: 'Gracias. Si te quedó una duda, preguntale a Hornero o escribinos.'}</span
						>
					{:else}
						<span>¿Te sirvió este artículo?</span>
						<span class="botones">
							<button type="button" onclick={() => (voto = 'si')}
								><Icon icon={I.phThumbsUpBold} size={15} />Sí</button
							>
							<button type="button" onclick={() => (voto = 'no')}
								><Icon icon={I.phThumbsDownBold} size={15} />No</button
							>
						</span>
					{/if}
				</div>
			</article>
			{#if otros.length}
				<h2 class="mas">Más de {c.nombre.toLowerCase()}</h2>
				<div class="tarjeta">
					{#each otros as o (o.slug)}
						<a
							href={resolve('/(sitio)/ayuda/[categoria]/[articulo]', {
								categoria: o.categoria,
								articulo: o.slug
							})}
							class="art"
						>
							<Icon icon={I.phArticle} size={20} style="color:#b0441d;flex:none" />
							<span class="a-txt"
								><span class="a-t">{o.titulo}</span><span class="a-s"
									>{minutos(o)} min de lectura</span
								></span
							>
							<Icon icon={I.phCaretRightBold} size={14} style="color:#9A8576" />
						</a>
					{/each}
				</div>
			{/if}
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
	.migas {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 14px;
		font-weight: 600;
		color: #9a8576;
	}
	.migas a {
		color: #5c4536;
		text-decoration: none;
	}
	h1 {
		margin: 18px 0 0;
		font-size: clamp(30px, 4.5vw, 52px);
		line-height: 1.05;
		letter-spacing: -0.04em;
		font-weight: 600;
		max-width: 860px;
		text-wrap: balance;
	}
	.sub {
		margin-top: 10px;
		font-size: 15.5px;
		color: #5c4536;
		text-wrap: pretty;
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
	.cuerpo {
		padding: clamp(20px, 3vw, 32px);
		font-size: 16px;
		line-height: 1.65;
		color: #3a2418;
	}
	.cuerpo p {
		margin: 0 0 16px;
		text-wrap: pretty;
	}
	.cuerpo h2 {
		font-size: 20px;
		margin: 24px 0 8px;
	}
	.cuerpo ol,
	.cuerpo ul {
		margin: 0 0 16px;
		padding-left: 22px;
	}
	.cuerpo li {
		margin: 0 0 8px;
		padding-left: 4px;
		text-wrap: pretty;
	}
	.cuerpo ol li::marker {
		color: #b0441d;
		font-weight: 700;
	}
	.nota {
		display: flex;
		gap: 10px;
		padding: 14px 16px;
		border-radius: 16px;
		background: #fcede5;
		font-size: 15px;
		margin: 0 0 16px;
	}
	.nota :global(svg) {
		margin-top: 4px;
	}
	.voto {
		margin-top: 22px;
		padding-top: 18px;
		border-top: 1px solid rgba(58, 29, 18, 0.08);
		display: flex;
		align-items: center;
		gap: 8px;
		flex-wrap: wrap;
		font-size: 14.5px;
		font-weight: 600;
	}
	.voto span {
		margin-right: 6px;
	}
	.botones {
		display: flex;
		gap: 8px;
	}
	.voto button {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 8px 14px;
		border-radius: 999px;
		border: 1.5px solid rgba(58, 29, 18, 0.16);
		background: #fff;
		font:
			600 13.5px 'Instrument Sans',
			sans-serif;
		color: #2a140c;
	}
	.mas {
		margin: 30px 0 14px;
		font-size: 22px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}
	.art {
		display: flex;
		align-items: center;
		gap: 14px;
		padding: 16px 18px;
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
		font-size: 15px;
		font-weight: 600;
	}
	.a-s {
		display: block;
		font-size: 12.5px;
		color: #7a6252;
		margin-top: 2px;
	}
	@media (max-width: 860px) {
		.dos {
			grid-template-columns: minmax(0, 1fr);
		}
	}
</style>
