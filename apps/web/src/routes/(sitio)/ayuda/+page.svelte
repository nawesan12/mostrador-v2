<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import {
		ARTICULOS,
		CATEGORIAS,
		FAQ,
		buscar,
		delaCategoria,
		destacados,
		minutos,
		categoria
	} from '#lib/ayuda/articulos.ts';
	import Lateral from '#lib/ayuda/Lateral.svelte';
	import Faq from '#lib/sitio/Faq.svelte';

	/** Centro de ayuda (design/Mostrador Ayuda.dc.html). El buscador va en la URL (?q=). */
	const q = $derived(page.url.searchParams.get('q') ?? '');
	let texto = $derived(q);
	const resultados = $derived(q.trim() ? buscar(q, 12) : []);
	const MAS_BUSCADO = [
		'Cuotas sin interés',
		'Factura A',
		'Andreani',
		'Dominio propio',
		'Migrar de Tiendanube'
	];
	function buscarEn(t: string) {
		const v = t.trim();
		const destino = v
			? `${resolve('/(sitio)/ayuda')}?q=${encodeURIComponent(v)}`
			: resolve('/(sitio)/ayuda');
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- la ayuda con su búsqueda
		void goto(destino);
	}
	const enlace = (a: (typeof ARTICULOS)[number]) =>
		resolve('/(sitio)/ayuda/[categoria]/[articulo]', { categoria: a.categoria, articulo: a.slug });
	const plural = (n: number) => `${n} ${n === 1 ? 'artículo' : 'artículos'}`;
</script>

<Seo
	title="Centro de ayuda · Mostrador"
	description="Guías, preguntas frecuentes y Hornero para resolver todo al instante."
	path="/ayuda"
/>

<section class="hero">
	<div class="ancho">
		<div class="h-in">
			<img src="/hornero/lupa.png" alt="" width="110" height="110" />
			<h1>¿En qué te ayudamos?</h1>
			<form
				class="buscador"
				role="search"
				onsubmit={(e) => {
					e.preventDefault();
					buscarEn(texto);
				}}
			>
				<Icon icon={I.phMagnifyingGlassBold} size={19} style="color:#7A6252" />
				<input
					bind:value={texto}
					placeholder="Buscá: cuotas, envíos, factura A, dominio…"
					aria-label="Buscar en la ayuda"
				/>
				<button>Buscar</button>
			</form>
			<div class="populares">
				<span>Lo más buscado:</span>
				{#each MAS_BUSCADO as t (t)}
					<button type="button" onclick={() => buscarEn(t)}>{t}</button>
				{/each}
			</div>
		</div>
	</div>
</section>

<div class="ancho">
	{#if q.trim()}
		<div class="resultados">
			<div class="r-t">
				{resultados.length ? `${plural(resultados.length)} para «${q}»` : `Nada para «${q}»`}
			</div>
			<div class="tarjeta lista">
				{#each resultados as a (a.slug)}
					<a href={enlace(a)} class="art">
						<Icon icon={I.phArticle} size={20} style="color:#C24E25;flex:none" />
						<span class="a-txt"
							><span class="a-t">{a.titulo}</span><span class="a-s"
								>{categoria(a.categoria)?.nombre} · {minutos(a)} min de lectura</span
							></span
						>
						<Icon icon={I.phCaretRightBold} size={14} style="color:#9A8576" />
					</a>
				{:else}
					<div class="nada">Probá con otras palabras o preguntale a Hornero acá abajo.</div>
				{/each}
			</div>
		</div>
	{/if}

	<div class="categorias">
		{#each CATEGORIAS as c (c.slug)}
			<a href={resolve('/(sitio)/ayuda/[categoria]', { categoria: c.slug })} class="cat">
				<span class="c-ico" style:background={c.bg}
					><Icon icon={c.icono} size={23} style="color:{c.fg}" /></span
				>
				<span class="c-n">{c.nombre}</span>
				<span class="c-c">{plural(delaCategoria(c.slug).length)}</span>
			</a>
		{/each}
	</div>

	<div class="dos">
		<div>
			<h2>Artículos más leídos</h2>
			<div class="tarjeta lista">
				{#each destacados() as a (a.slug)}
					<a href={enlace(a)} class="art">
						<Icon icon={I.phArticle} size={20} style="color:#C24E25;flex:none" />
						<span class="a-txt"
							><span class="a-t">{a.titulo}</span><span class="a-s"
								>{categoria(a.categoria)?.nombre.split(' ')[0]} · {minutos(a)} min de lectura</span
							></span
						>
						<Icon icon={I.phCaretRightBold} size={14} style="color:#9A8576" />
					</a>
				{/each}
			</div>
			<h2 class="faq-t">Preguntas frecuentes</h2>
			<div class="faq"><Faq items={FAQ} /></div>
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
	}
	.h-in {
		padding: clamp(48px, 7vw, 96px) 0;
		text-align: center;
	}
	.h-in img {
		object-fit: contain;
	}
	h1 {
		margin: 10px 0 0;
		font-size: clamp(40px, 6vw, 76px);
		line-height: 1;
		letter-spacing: -0.05em;
		font-weight: 600;
		text-wrap: balance;
	}
	.buscador {
		margin: 24px auto 0;
		max-width: 640px;
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 8px 8px 8px 18px;
		border-radius: 999px;
		background: #fff;
		box-shadow: 0 20px 50px -26px rgba(58, 29, 18, 0.5);
	}
	.buscador input {
		flex: 1;
		min-width: 0;
		border: 0;
		outline: 0;
		background: transparent;
		font:
			500 16px 'Instrument Sans',
			sans-serif;
		color: #2a140c;
		padding: 8px 0;
	}
	.buscador button {
		padding: 12px 20px;
		border: 0;
		border-radius: 999px;
		background: #2a140c;
		color: #fff;
		font:
			600 14.5px 'Instrument Sans',
			sans-serif;
	}
	.populares {
		margin-top: 14px;
		display: flex;
		justify-content: center;
		gap: 6px;
		flex-wrap: wrap;
		font-size: 13.5px;
	}
	.populares span {
		color: #7a6252;
		padding: 6px 0;
	}
	.populares button {
		padding: 6px 12px;
		border-radius: 999px;
		border: 0;
		background: #fff;
		font:
			600 13.5px 'Instrument Sans',
			sans-serif;
		color: #2a140c;
		white-space: nowrap;
	}
	.resultados {
		padding-top: clamp(32px, 5vw, 56px);
	}
	.r-t {
		font-size: 20px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}
	.resultados .lista {
		margin-top: 14px;
	}
	.nada {
		padding: 18px;
		font-size: 14.5px;
		color: #5c4536;
	}
	.categorias {
		padding: clamp(48px, 7vw, 80px) 0 0;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 250px), 1fr));
		gap: 12px;
	}
	.cat {
		display: flex;
		flex-direction: column;
		gap: 12px;
		padding: 20px;
		border-radius: 24px;
		background: #fff;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.06);
		text-decoration: none;
		color: #2a140c;
		transition: transform 0.2s;
	}
	.cat:hover {
		transform: translateY(-2px);
	}
	.c-ico {
		width: 46px;
		height: 46px;
		border-radius: 14px;
		display: grid;
		place-items: center;
	}
	.c-n {
		font-size: 17px;
		font-weight: 600;
	}
	.c-c {
		font-size: 13.5px;
		color: #7a6252;
	}
	.dos {
		padding: clamp(48px, 7vw, 80px) 0;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 380px), 1fr));
		gap: clamp(18px, 3vw, 36px);
		align-items: start;
	}
	h2 {
		margin: 0;
		font-size: 24px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}
	.faq-t {
		margin-top: 30px;
	}
	.tarjeta {
		border-radius: 24px;
		background: #fff;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.06);
		overflow: hidden;
	}
	.dos .lista,
	.faq {
		margin-top: 14px;
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
</style>
