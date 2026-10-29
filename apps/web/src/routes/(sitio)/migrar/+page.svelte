<script lang="ts">
	import { page } from '$app/state';
	import { replaceState } from '$app/navigation';
	import { resolve } from '$app/paths';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import Etiqueta from '#lib/sitio/Etiqueta.svelte';
	import Faq from '#lib/sitio/Faq.svelte';
	import Titulo from '#lib/sitio/Titulo.svelte';

	/**
	 * Migrar (design/Mostrador Migrar.dc.html), con la pestaña de cada plataforma (?desde=). El campo y
	 * "Empezar" llevan al onboarding con el origen (?desde=, ?origen=); la importación la hace ese flujo.
	 */
	// Lo que trae cada plataforma, según lo que hoy hace el importador (sin prometer de más).
	const ORIGENES = [
		{
			id: 'tiendanube',
			tab: 'Desde Tiendanube',
			ico: I.phCloudArrowDownFill,
			t: 'Conectás tu tienda y la traemos',
			d: 'Autorizás el acceso desde Tiendanube y traemos todo por su API oficial, sin exportar nada.',
			campo: null,
			rotulo: '',
			trae: [
				'Productos con variantes y todas sus fotos',
				'Categorías',
				'Clientes',
				'Pedidos anteriores, en un historial aparte',
				'Tus links viejos, redirigidos a los nuevos'
			]
		},
		{
			id: 'shopify',
			tab: 'Desde Shopify',
			ico: I.phLinkBold,
			t: 'Pegás el link de tu tienda',
			d: 'Leemos tu catálogo público, sin claves ni exportar nada.',
			campo: 'mitienda.myshopify.com',
			rotulo: 'Link de tu tienda en Shopify',
			trae: [
				'Productos con variantes, fotos y precios',
				'Categorías',
				'Tus links viejos, redirigidos a los nuevos',
				'El stock no lo publica Shopify: lo revisás después'
			]
		},
		{
			id: 'woocommerce',
			tab: 'Desde WooCommerce',
			ico: I.phLinkBold,
			t: 'Pegás el link de tu tienda',
			d: 'Leemos tu catálogo público, sin claves ni plugins.',
			campo: 'mitienda.com.ar',
			rotulo: 'Link de tu tienda en WooCommerce',
			trae: [
				'Productos con variantes, fotos y precios',
				'Categorías',
				'Tus links viejos, redirigidos a los nuevos'
			]
		},
		{
			id: 'mercadolibre',
			tab: 'Desde Mercado Libre',
			ico: I.phStorefrontFill,
			t: 'Conectás tu cuenta de Mercado Libre',
			d: 'Publicás tus productos en Mercado Libre desde el panel. Traer tus publicaciones de hoy llega pronto.',
			campo: null,
			rotulo: '',
			trae: [
				'Tus productos publicados en Mercado Libre desde el panel',
				'Precio y stock desde un solo lugar'
			]
		},
		{
			id: 'empretienda',
			tab: 'Desde Empretienda',
			ico: I.phHeadsetFill,
			t: 'Te la mudamos nosotros',
			d: 'Empretienda no deja leer las tiendas desde nuestros servidores. Escribinos y la mudamos con vos.',
			campo: null,
			rotulo: '',
			trae: ['Productos con fotos y precios', 'Categorías']
		}
	] as const;
	/** Las que se importan solas desde el onboarding (con ?desde= y ?origen=). */
	const SOLAS = ['tiendanube', 'shopify', 'woocommerce'];
	const PASOS = [
		['Conectás o pegás el link', 'En dos minutos'],
		['Revisás lo importado', 'Precios, stock y fotos, antes de vender'],
		['Elegís plantilla y cobros', 'Mercado Pago, transferencia y envíos'],
		['Vendés desde Tu Local Web', 'Y apuntás tu dominio cuando quieras']
	] as const;
	const VENTAJAS = [
		[I.phGiftFill, '14 días gratis', 'Probás todo sin tarjeta antes de elegir un plan.'],
		[
			I.phCloudArrowDownFill,
			'Importar no se cobra',
			'Traer tu tienda es gratis, en cualquier plan.'
		],
		[
			I.phChartLineUpFill,
			'Tus links viejos siguen andando',
			'Cuando apuntás tu dominio a Tu Local Web, cada link viejo de producto lleva al nuevo.'
		]
	] as const;
	const FAQ = [
		[
			'¿Tengo que cerrar mi tienda vieja?',
			'No. Las dos conviven mientras revisás todo; cuando estés conforme, apuntás tu dominio a Tu Local Web.'
		],
		[
			'¿Pierdo mi posicionamiento en Google?',
			'Cada producto importado guarda su dirección vieja: cuando tu dominio apunta a Tu Local Web, esa dirección lleva al producto nuevo con una redirección 301.'
		],
		[
			'¿Cuánto cuesta migrar?',
			'Nada: importar tu tienda es gratis y tenés 14 días para probar sin tarjeta.'
		],
		[
			'¿Y si tengo muchos productos?',
			'Se importan en tandas, sin que tengas que dejar la pantalla abierta. Si son más de 10.000, escribinos.'
		]
	] as const;

	let elegido = $state(
		Math.max(
			0,
			ORIGENES.findIndex((o) => o.id === page.url.searchParams.get('desde'))
		)
	);
	const o = $derived(ORIGENES[elegido]);
	// En el celular la fila de pestañas se desliza: la elegida siempre queda a la vista.
	let tabs = $state<HTMLElement>();
	$effect(() => {
		const b = tabs?.querySelectorAll('button')[elegido];
		if (tabs && b) tabs.scrollTo({ left: b.offsetLeft - 16, behavior: 'smooth' });
	});
	// La pestaña queda en la URL (?desde=), para compartir el link de cada plataforma.
	function elegir(i: number) {
		elegido = i;
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- resolve adentro; sólo suma la query
		replaceState(`${resolve('/(sitio)/migrar')}?desde=${ORIGENES[i].id}`, {});
	}
</script>

<Seo
	title="Migrar a Tu Local Web · Tu Local Web"
	description="Traé tu tienda gratis desde Tiendanube, Shopify o WooCommerce: productos con fotos y variantes, y desde Tiendanube también clientes y pedidos."
	path="/migrar"
	image="migrar"
/>

<section class="hero">
	<div class="ancho">
		<div class="hero-in">
			<div class="cab">
				<div class="cab-txt">
					<Etiqueta texto="Migración gratis" />
					<h1>Mudate a Tu Local Web.</h1>
					<p>
						Traemos tus productos con fotos y variantes, y desde Tiendanube también tus clientes y
						pedidos. Tus links viejos siguen llevando a tus productos.
					</p>
				</div>
				<img src="/hornero/caja.webp" alt="" width="260" height="260" />
			</div>
			<div class="tabs" role="group" aria-label="Plataforma de origen" bind:this={tabs}>
				{#each ORIGENES as og, i (og.id)}
					<button type="button" aria-pressed={elegido === i} onclick={() => elegir(i)}
						>{og.tab}</button
					>
				{/each}
			</div>
			<div class="detalle">
				<div class="conectar">
					<span class="c-ico"><Icon icon={o.ico} size={26} style="color:#F0A58A" /></span>
					<div class="c-t">{o.t}</div>
					<p>{o.d}</p>
					{#if SOLAS.includes(o.id)}
						<form action="/onboarding" method="GET">
							<input type="hidden" name="desde" value={o.id} />
							{#if o.campo}
								<input
									name="origen"
									placeholder={o.campo}
									aria-label={o.rotulo}
									autocomplete="off"
								/>
							{/if}
							<button type="submit">Empezar</button>
						</form>
					{:else}
						<form action={o.id === 'empretienda' ? '/contacto' : '/onboarding'} method="GET">
							<button type="submit">{o.id === 'empretienda' ? 'Escribinos' : 'Empezar'}</button>
						</form>
					{/if}
				</div>
				<div class="trae">
					<div class="tr-t">Lo que traemos</div>
					<div class="tr-g">
						{#each o.trae as t (t)}
							<div>
								<Icon
									icon={I.phCheckCircleFill}
									size={18}
									style="color:#237550;flex:none;margin-top:1px"
								/>{t}
							</div>
						{/each}
					</div>
				</div>
			</div>
		</div>
	</div>
</section>

<section class="bloque">
	<div class="ancho">
		<Titulo etiqueta="Cómo funciona" titulo="Cuatro pasos, cero cortes" />
		<div class="pasos">
			{#each PASOS as [t, d], i (t)}
				<div class="paso">
					<div class="p-n" class:ultimo={i === PASOS.length - 1} aria-hidden="true">0{i + 1}</div>
					<div class="p-t">{t}</div>
					<div class="p-d">{d}</div>
				</div>
			{/each}
		</div>
		<div class="ventajas">
			{#each VENTAJAS as [ico, t, d] (t)}
				<div class="ventaja">
					<Icon icon={ico} size={26} style="color:#b0441d;flex:none" />
					<div>
						<b>{t}</b>
						<p>{d}</p>
					</div>
				</div>
			{/each}
		</div>
	</div>
</section>

<section class="bloque beige">
	<div class="ancho">
		<div class="dos">
			<div>
				<Etiqueta texto="Preguntas" />
				<div class="faq"><Faq items={FAQ} /></div>
			</div>
		</div>
	</div>
</section>

<style>
	.ancho {
		max-width: 1240px;
		margin: 0 auto;
		padding: 0 clamp(16px, 3vw, 28px);
	}
	.hero {
		background: #f4ece2;
	}
	.hero-in {
		padding: clamp(48px, 7vw, 96px) 0 clamp(36px, 5vw, 56px);
	}
	.cab {
		display: flex;
		align-items: center;
		gap: 28px;
		flex-wrap: wrap;
	}
	.cab-txt {
		flex: 1 1 520px;
		min-width: 0;
	}
	h1 {
		margin: 14px 0 0;
		font-size: clamp(44px, 6.6vw, 92px);
		line-height: 0.94;
		letter-spacing: -0.05em;
		font-weight: 600;
	}
	.cab-txt > p {
		margin: 16px 0 0;
		font-size: 18px;
		line-height: 1.55;
		color: #5c4536;
		max-width: 520px;
	}
	.cab > img {
		width: clamp(160px, 22vw, 260px);
		height: auto;
		flex: none;
		margin: 0 auto;
	}
	.tabs {
		position: relative;
		margin-top: 30px;
		display: flex;
		gap: 6px;
		padding: 6px;
		border-radius: 999px;
		background: #fff;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.08);
		overflow-x: auto;
		scrollbar-width: none;
		width: fit-content;
		max-width: 100%;
	}
	.tabs button {
		flex: none;
		padding: 11px 18px;
		border: 0;
		border-radius: 999px;
		background: #fff;
		color: #2a140c;
		font-size: 14.5px;
		font-weight: 600;
		white-space: nowrap;
	}
	.tabs button[aria-pressed='true'] {
		background: #2a140c;
		color: #fff;
	}
	.detalle {
		margin-top: 22px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 360px), 1fr));
		gap: 12px;
	}
	.conectar {
		padding: 26px;
		border-radius: 28px;
		background: #2a140c;
		color: #fbf7f2;
	}
	.c-ico {
		width: 50px;
		height: 50px;
		border-radius: 16px;
		background: rgba(251, 247, 242, 0.08);
		display: grid;
		place-items: center;
	}
	.c-t {
		margin-top: 16px;
		font-size: 24px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}
	.conectar p {
		margin: 8px 0 0;
		font-size: 15.5px;
		line-height: 1.6;
		color: #d9c8ba;
	}
	form {
		margin-top: 18px;
		display: flex;
		padding: 5px;
		border-radius: 999px;
		background: #fff;
	}
	form input {
		flex: 1;
		min-width: 0;
		padding: 10px 14px;
		border: 0;
		background: transparent;
		font-size: 14.5px;
		font-weight: 500;
		color: #2a140c;
		outline: none;
	}
	form button {
		padding: 11px 18px;
		border: 0;
		border-radius: 999px;
		background: #b0441d;
		color: #fff;
		font-size: 14px;
		font-weight: 600;
		white-space: nowrap;
	}
	.trae {
		padding: 26px;
		border-radius: 24px;
		background: #fff;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.06);
	}
	.tr-t {
		font-size: 13px;
		font-weight: 700;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: #7a6252;
	}
	.tr-g {
		margin-top: 14px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 10px;
	}
	.tr-g div {
		display: flex;
		gap: 10px;
		align-items: flex-start;
		font-size: 14.5px;
		line-height: 1.45;
	}
	.bloque {
		padding: clamp(56px, 8vw, 104px) 0;
	}
	.beige {
		background: #f4ece2;
	}
	.pasos {
		margin-top: 28px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 230px), 1fr));
		gap: 12px;
	}
	.paso {
		padding: 22px;
		border-radius: 24px;
		background: #fff;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.06);
	}
	.p-n {
		font-size: 42px;
		font-weight: 600;
		letter-spacing: -0.04em;
		color: #e7d6c6;
	}
	.p-n.ultimo {
		color: #b0441d;
	}
	.p-t {
		margin-top: 10px;
		font-size: 17px;
		font-weight: 600;
	}
	.p-d {
		margin-top: 4px;
		font-size: 14px;
		color: #7a6252;
	}
	.ventajas {
		margin-top: 14px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 300px), 1fr));
		gap: 12px;
	}
	.ventaja {
		display: flex;
		gap: 14px;
		padding: 20px;
		border-radius: 24px;
		background: #fcede5;
	}
	.ventaja b {
		font-size: 16px;
	}
	.ventaja p {
		margin: 4px 0 0;
		font-size: 14px;
		line-height: 1.5;
		color: #5c4536;
	}
	.dos {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 380px), 1fr));
		gap: clamp(20px, 4vw, 48px);
		align-items: start;
	}
	.faq {
		margin-top: 14px;
	}
	@media (max-width: 640px) {
		.cab > img {
			width: 140px;
		}
		.tabs {
			width: auto;
			margin-left: calc(-1 * clamp(16px, 3vw, 28px));
			margin-right: calc(-1 * clamp(16px, 3vw, 28px));
			border-radius: 0;
			padding: 6px clamp(16px, 3vw, 28px);
			background: transparent;
			box-shadow: none;
		}
		.tabs button {
			box-shadow: 0 1px 2px rgba(58, 29, 18, 0.08);
		}
	}
</style>
