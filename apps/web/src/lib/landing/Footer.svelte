<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import Marca from './Marca.svelte';

	/** Footer del sitio (handoff): columnas, barra legal y el botón de arrepentimiento (obligatorio). */
	const COLUMNAS = [
		[
			'Producto',
			[
				['#producto', 'Tienda online'],
				['#producto', 'WhatsApp'],
				['#producto', 'Punto de venta'],
				['#hornero', 'Asistente IA']
			]
		],
		[
			'Empresa',
			[
				['#precios', 'Precios'],
				['/seguridad', 'Seguridad'],
				['/estado', 'Estado del sistema'],
				['/contacto', 'Contacto'],
				['/nosotros', 'Sobre nosotros'],
				['/partners', 'Partners'],
				['/novedades', 'Novedades']
			]
		],
		[
			'Migrar desde',
			[
				['/migrar?desde=tiendanube', 'Tiendanube'],
				['/migrar?desde=empretienda', 'Empretienda'],
				['/migrar?desde=mercadolibre', 'Mercado Libre'],
				['/ayuda', 'Centro de ayuda']
			]
		],
		[
			'Legales',
			[
				['/legales/terminos', 'Términos y condiciones'],
				['/legales/privacidad', 'Política de privacidad'],
				['/legales/cookies', 'Política de cookies'],
				['/legales/arrepentimiento', 'Botón de arrepentimiento'],
				['/legales/defensa-del-consumidor', 'Defensa del consumidor'],
				['/legales/tratamiento-de-datos', 'Tratamiento de datos (para tiendas)'],
				['/legales/uso-aceptable', 'Uso aceptable'],
				['/legales/reembolsos', 'Reembolsos de planes']
			]
		]
	] as const;
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- anclas de la landing y páginas públicas que llegan en el próximo PR -->
<footer>
	<div class="cols">
		<div class="marca">
			<div class="logo"><Marca alto={18} /><img src="/wordmark-brown.png" alt="mostrador" /></div>
			<div class="lema">Hecho en Argentina para negocios argentinos.</div>
		</div>
		{#each COLUMNAS as [titulo, links] (titulo)}
			<div class="col">
				<span class="col-t">{titulo}</span>
				{#each links as [href, t] (t)}<a {href}>{t}</a>{/each}
			</div>
		{/each}
	</div>
	<div class="barra">
		<div class="barra-in">
			<span>© 2026 Mostrador · Nahuel Santillán · <span class="nw">CUIT 20-44859234-1</span></span>
			<span class="legal">
				<a class="arrep" href="/legales/arrepentimiento"
					><Icon icon={I.phArrowCounterClockwiseBold} size={13} />Botón de arrepentimiento</a
				>
				<a href="/legales/defensa-del-consumidor">Defensa del consumidor</a>
				<a href="/legales/cookies">Cookies</a>
				<span class="fiscal"><Icon icon={I.phQrCodeBold} size={15} />Data fiscal</span>
			</span>
		</div>
	</div>
</footer>

<!-- eslint-enable svelte/no-navigation-without-resolve -->

<style>
	footer {
		border-top: 1px solid rgba(58, 29, 18, 0.1);
	}
	.cols {
		max-width: 1240px;
		margin: 0 auto;
		padding: 56px var(--lado) 110px;
		display: flex;
		gap: 48px;
		flex-wrap: wrap;
	}
	.marca {
		flex: 1 1 260px;
	}
	.logo {
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.logo img {
		height: 18px;
		width: auto;
	}
	.lema {
		margin-top: 16px;
		font-size: 14px;
		color: #7a6252;
		max-width: 260px;
		line-height: 1.5;
	}
	.col {
		flex: 1 1 140px;
		display: flex;
		flex-direction: column;
		gap: 10px;
		font-size: 14.5px;
	}
	.col-t {
		font-size: 13px;
		font-weight: 600;
		color: #2a140c;
	}
	.col a {
		color: #5c4536;
	}
	.barra {
		max-width: 1240px;
		margin: 0 auto;
		padding: 0 var(--lado) 26px;
	}
	.barra-in {
		padding-top: 18px;
		border-top: 1px solid rgba(58, 29, 18, 0.1);
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 10px 22px;
		flex-wrap: wrap;
		font-size: 12.5px;
		color: #7a6252;
	}
	.legal {
		display: flex;
		gap: 8px 16px;
		flex-wrap: wrap;
		align-items: center;
	}
	.legal a {
		color: #5c4536;
	}
	.legal .arrep {
		color: #2a140c;
		font-weight: 700;
		display: inline-flex;
		align-items: center;
		gap: 5px;
	}
	.fiscal {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		padding: 5px 9px;
		border-radius: 6px;
		background: #fff;
		color: #1a1a1a;
		font-weight: 700;
		box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.06);
	}
	.nw {
		white-space: nowrap;
	}
</style>
