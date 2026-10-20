<script lang="ts">
	import Icon from './Icon.svelte';
	import { phHandbag } from '#lib/icons.ts';
	import { sobreAcento, type Plantilla } from '#lib/plantillas.ts';

	/**
	 * Vista en miniatura de una tienda con su plantilla, idéntica a la del handoff (galería del
	 * onboarding, sección Plantillas de la landing y selector del panel). Tres layouts:
	 * editorial (ed), grilla (grid) y catálogo (cat).
	 */
	let {
		t: base,
		paleta = null
	}: {
		t: Plantilla;
		/** Colores que pisan los de la plantilla (handoff: "Con tus colores", "Probá otros colores"). */
		paleta?: { bg: string; tx: string; ac: string } | null;
	} = $props();
	const t = $derived(paleta ? { ...base, ...paleta } : base);
	const onAc = $derived(sobreAcento(t.ac));
</script>

<div class="mini" style:background={t.bg} style:color={t.tx} style:font-family={t.ff}>
	<div class="nav">
		<span class="logo" style:background={t.ac}></span>
		<b>{t.tienda}</b><span class="esp"></span>
		<span class="tenue">Tienda</span><span class="tenue">Nosotros</span><Icon
			icon={phHandbag}
			size={11}
		/>
	</div>

	{#if t.layout === 'ed'}
		<div class="ed-hero" style:border-radius={t.radio}>
			<img src={t.img} alt="" loading="lazy" />
			<div class="ed-sombra"></div>
			<div class="ed-texto">
				<div class="ed-lema">{t.lema}</div>
				<div class="ed-cta">Ver colección</div>
			</div>
		</div>
		<div class="ed-prods">
			{#each t.productos as p (p.pos)}
				<div>
					<div class="foto" style:border-radius={t.radio}>
						<img
							src={t.img}
							alt=""
							loading="lazy"
							style:object-position={p.pos}
							style:transform="scale(1.6)"
						/>
					</div>
					<div class="raya" style:width="70%"></div>
					<div class="precio">{p.precio}</div>
				</div>
			{/each}
		</div>
	{:else if t.layout === 'grid'}
		<div class="gr-banda" style:border-radius={t.radio} style:background={t.ac} style:color={onAc}>
			<span class="gr-lema">{t.lema}</span>
			<span class="gr-cta" style:background={t.bg} style:color={t.tx}>Comprar</span>
		</div>
		<div class="gr-prods">
			{#each t.productos as p (p.pos)}
				<div>
					<div class="foto" style:border-radius={t.radio}>
						<img
							src={t.img}
							alt=""
							loading="lazy"
							style:object-position={p.pos}
							style:transform="scale(1.5)"
						/>
					</div>
					<div class="raya" style:width="75%"></div>
					<div class="precio" style:color={t.ac}>{p.precio}</div>
				</div>
			{/each}
		</div>
	{:else}
		<div class="ca">
			<div class="ca-menu">
				<b>Categorías</b><span>Destacados</span><span>Novedades</span><span>Combos</span><span
					>Regalos</span
				>
				<span class="ca-ofertas" style:background={t.ac} style:color={onAc}>Ofertas</span>
			</div>
			<div class="ca-cuerpo">
				<div class="ca-hero" style:border-radius={t.radio}>
					<img src={t.img} alt="" loading="lazy" />
					<div class="ca-lema">{t.lema}</div>
				</div>
				<div class="ca-prods">
					{#each t.productos as p (p.pos)}
						<div class="ca-prod">
							<span class="ca-mini"
								><img
									src={t.img}
									alt=""
									loading="lazy"
									style:object-position={p.pos}
									style:transform="scale(1.6)"
								/></span
							>
							<span class="ca-precio">{p.precio}</span>
						</div>
					{/each}
				</div>
			</div>
		</div>
	{/if}
</div>

<style>
	.mini {
		height: 100%;
		display: flex;
		flex-direction: column;
		overflow: hidden;
	}
	img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		display: block;
	}
	.nav {
		display: flex;
		align-items: center;
		gap: 7px;
		padding: 9px 11px;
		font-size: 8.5px;
		flex: none;
	}
	.logo {
		width: 14px;
		height: 14px;
		border-radius: 50%;
		flex: none;
	}
	.nav b {
		font-size: 10.5px;
		letter-spacing: -0.01em;
		white-space: nowrap;
	}
	.esp {
		flex: 1;
	}
	.tenue {
		opacity: 0.6;
	}
	/* Editorial */
	.ed-hero {
		position: relative;
		margin: 0 9px;
		flex: 1 1 56%;
		min-height: 0;
		overflow: hidden;
	}
	.ed-sombra {
		position: absolute;
		inset: 0;
		background: linear-gradient(0deg, rgba(0, 0, 0, 0.55), transparent 60%);
	}
	.ed-texto {
		position: absolute;
		left: 10px;
		right: 10px;
		bottom: 10px;
		color: #fff;
	}
	.ed-lema {
		font-size: 15px;
		font-weight: 700;
		line-height: 1.05;
		letter-spacing: -0.02em;
	}
	.ed-cta {
		margin-top: 6px;
		display: inline-block;
		font-size: 7.5px;
		font-weight: 700;
		padding: 4px 8px;
		border-radius: 999px;
		background: #fff;
		color: #111;
	}
	.ed-prods {
		padding: 9px;
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(60px, 1fr));
		gap: 7px;
		flex: none;
		max-height: 38%;
		overflow: hidden;
	}
	.foto {
		aspect-ratio: 1;
		overflow: hidden;
		background: rgba(0, 0, 0, 0.06);
	}
	.raya {
		margin-top: 4px;
		height: 4px;
		border-radius: 2px;
		background: currentColor;
		opacity: 0.25;
	}
	.precio {
		margin-top: 3px;
		font-size: 8px;
		font-weight: 700;
	}
	/* Grilla */
	.gr-banda {
		margin: 0 9px;
		padding: 9px 11px;
		display: flex;
		align-items: center;
		gap: 8px;
		flex: none;
	}
	.gr-lema {
		flex: 1;
		font-size: 11px;
		font-weight: 700;
		letter-spacing: -0.01em;
		line-height: 1.1;
	}
	.gr-cta {
		font-size: 7.5px;
		font-weight: 700;
		padding: 4px 7px;
		border-radius: 999px;
	}
	.gr-prods {
		padding: 9px;
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(62px, 1fr));
		gap: 8px;
		flex: 1;
		min-height: 0;
		overflow: hidden;
		align-content: start;
	}
	/* Catálogo */
	.ca {
		flex: 1;
		min-height: 0;
		display: flex;
		gap: 8px;
		padding: 0 9px 9px;
	}
	.ca-menu {
		width: 30%;
		flex: none;
		display: flex;
		flex-direction: column;
		gap: 6px;
		padding-top: 4px;
		font-size: 8.5px;
	}
	.ca-menu b {
		font-size: 9px;
	}
	.ca-menu span {
		opacity: 0.65;
	}
	.ca-menu .ca-ofertas {
		opacity: 1;
		margin-top: 4px;
		padding: 5px 6px;
		border-radius: 6px;
		font-weight: 700;
		font-size: 7.5px;
		text-align: center;
	}
	.ca-cuerpo {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 7px;
	}
	.ca-hero {
		position: relative;
		flex: 1;
		min-height: 0;
		overflow: hidden;
	}
	.ca-lema {
		position: absolute;
		left: 8px;
		bottom: 8px;
		right: 8px;
		font-size: 12px;
		font-weight: 700;
		line-height: 1.05;
		color: #fff;
		text-shadow: 0 1px 8px rgba(0, 0, 0, 0.4);
	}
	.ca-prods {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(54px, 1fr));
		gap: 6px;
		flex: none;
		max-height: 40%;
		overflow: hidden;
	}
	.ca-prod {
		display: flex;
		gap: 5px;
		align-items: center;
		padding: 4px;
		border-radius: 8px;
		background: rgba(0, 0, 0, 0.05);
	}
	.ca-mini {
		width: 22px;
		height: 22px;
		flex: none;
		border-radius: 5px;
		overflow: hidden;
	}
	.ca-precio {
		font-size: 7.5px;
		font-weight: 700;
	}
</style>
