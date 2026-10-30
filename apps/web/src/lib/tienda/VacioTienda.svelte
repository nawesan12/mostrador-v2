<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { varsTema, type TemaOpiniones } from './opiniones.ts';
	import type { TiendaPublica } from './tipos.ts';

	/**
	 * Lo que ve quien compra cuando no hay nada que mostrar: el catálogo de una tienda que todavía no
	 * cargó productos, o un producto que no existe. Vestido con el tema de la plantilla (el mismo de
	 * reseñas y preguntas). Sin fechas ni promesas: lo que hay y cómo seguir (WhatsApp, el local o
	 * conocer la tienda).
	 */
	let {
		tienda,
		tema,
		motivo
	}: {
		tienda: TiendaPublica;
		tema: TemaOpiniones;
		motivo: 'sin-productos' | 'no-existe';
	} = $props();

	const base = $derived(tienda.base ?? `/t/${tienda.slug}`);
	const whatsapp = $derived(tienda.whatsapp.replace(/\D/g, ''));
	const linkWa = $derived(
		whatsapp
			? `https://wa.me/${whatsapp.startsWith('54') ? whatsapp : '54' + whatsapp}?text=${encodeURIComponent('¡Hola! Vi la tienda online y quería consultarles.')}`
			: null
	);
	const hay = $derived(tienda.productos.length > 0);
	const texto = $derived(
		motivo === 'no-existe'
			? hay
				? 'Puede que lo hayan sacado o que el link esté mal escrito. Fijate en el resto de los productos.'
				: 'Puede que lo hayan sacado o que el link esté mal escrito.'
			: linkWa
				? 'Muy pronto vas a poder comprar desde acá. Si buscás algo, escribinos.'
				: tienda.envios.retiro && tienda.direccion
					? `Muy pronto vas a poder comprar desde acá. Mientras tanto, te esperamos en ${tienda.direccion}.`
					: 'Muy pronto vas a poder comprar desde acá.'
	);
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
<div class="vacio" style={varsTema(tema)}>
	<span class="icono"
		><Icon
			icon={motivo === 'no-existe' ? I.phMagnifyingGlassBold : I.phStorefrontFill}
			size={26}
		/></span
	>
	<h2>
		{motivo === 'no-existe' ? 'No encontramos este producto' : 'Estamos cargando los productos'}
	</h2>
	<p>{texto}</p>
	<div class="acciones">
		{#if motivo === 'no-existe' && hay}
			<a class="principal" href="{base}/catalogo">Ver los productos</a>
		{:else if linkWa}
			<a class="principal" href={linkWa} target="_blank" rel="noopener"
				><Icon icon={I.phWhatsappLogoFill} size={18} />Escribinos por WhatsApp</a
			>
		{/if}
		{#if motivo === 'no-existe'}
			<a class={hay || linkWa ? 'segundo' : 'principal'} href={base || '/'}>Volver al inicio</a>
		{:else}
			<a class={linkWa ? 'segundo' : 'principal'} href="{base}/nosotros">Conocé la tienda</a>
		{/if}
	</div>
</div>

<style>
	.vacio {
		box-sizing: border-box;
		max-width: 560px;
		margin: 0 auto;
		padding: clamp(32px, 6vw, 56px) clamp(20px, 5vw, 48px);
		border: 1px solid var(--op-ln);
		border-radius: var(--op-r);
		background: var(--op-sf);
		color: var(--op-tx);
		font-family: var(--op-f);
		text-align: center;
	}
	.icono {
		display: inline-grid;
		place-items: center;
		width: 60px;
		height: 60px;
		border-radius: 50%;
		background: var(--op-ac);
		color: var(--op-on);
	}
	h2 {
		margin: 18px 0 0;
		font-family: var(--op-tf);
		font-weight: var(--op-tw);
		text-transform: var(--op-tt);
		letter-spacing: var(--op-ls);
		font-size: clamp(24px, 3.4vw, 34px);
		line-height: 1.1;
		text-wrap: balance;
	}
	p {
		margin: 12px auto 0;
		max-width: 400px;
		font-size: 15.5px;
		line-height: 1.55;
		color: var(--op-mu);
		text-wrap: pretty;
	}
	.acciones {
		margin-top: 24px;
		display: flex;
		justify-content: center;
		flex-wrap: wrap;
		gap: 10px;
	}
	a {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		min-height: 46px;
		padding: 0 22px;
		border-radius: var(--op-rb);
		font: 600 15px var(--op-f);
		text-decoration: none;
	}
	.principal {
		background: var(--op-ac);
		color: var(--op-on);
	}
	.segundo {
		border: 1.5px solid var(--op-ln);
		color: var(--op-tx);
	}
</style>
