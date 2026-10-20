<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import { phArrowCounterClockwiseBold, phQrCodeBold } from '#lib/icons.ts';

	/** Footer de las páginas públicas (handoff). "Configurar cookies" reabre el banner. */
	const COLUMNAS = [
		[
			'Producto',
			[
				['/#producto', 'Tienda online'],
				['/#plantillas', 'Plantillas'],
				['/#precios', 'Precios'],
				['/novedades', 'Novedades'],
				['/estado', 'Estado del sistema']
			]
		],
		[
			'Empresa',
			[
				['/nosotros', 'Sobre nosotros'],
				['/partners', 'Partners y agencias'],
				['/seguridad', 'Seguridad'],
				['/contacto', 'Contacto y ventas']
			]
		],
		[
			'Ayuda',
			[
				['/ayuda', 'Centro de ayuda'],
				['/migrar?desde=tiendanube', 'Migrar desde Tiendanube'],
				['/migrar?desde=empretienda', 'Migrar desde Empretienda'],
				['/migrar?desde=mercadoshops', 'Migrar desde Mercado Shops']
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

	const configurar = () => window.dispatchEvent(new CustomEvent('mostrador:configurar-cookies'));
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- páginas públicas y anclas de la landing -->
<footer>
	<div class="in">
		<div class="cols">
			<div class="marca">
				<img src="/wordmark-brown.png" alt="mostrador" width="104" height="22" />
				<p>Todo lo que tienen las otras, por menos plata y más seguro. Hecho en Argentina.</p>
			</div>
			{#each COLUMNAS as [titulo, links] (titulo)}
				<div class="col">
					<span class="col-t">{titulo}</span>
					{#each links as [href, t] (t)}<a {href}>{t}</a>{/each}
				</div>
			{/each}
		</div>
		<div class="barra">
			<span
				>© 2026 Mostrador Tecnología S.A.S. · <span class="nw">CUIT 30-71000000-0</span> · Av.
				Corrientes 1234, piso 8, C1043 CABA <b>(datos de ejemplo)</b></span
			>
			<span class="legal">
				<a class="arrep" href="/legales/arrepentimiento"
					><Icon icon={phArrowCounterClockwiseBold} size={13} />Botón de arrepentimiento</a
				>
				<a href="/legales/defensa-del-consumidor">Defensa del consumidor</a>
				<button type="button" onclick={configurar}>Configurar cookies</button>
				<span class="fiscal"><Icon icon={phQrCodeBold} size={15} />Data fiscal</span>
			</span>
		</div>
	</div>
</footer>

<!-- eslint-enable svelte/no-navigation-without-resolve -->

<style>
	footer {
		background: #f4ece2;
		border-top: 1px solid rgba(58, 29, 18, 0.1);
	}
	.in {
		max-width: 1240px;
		margin: 0 auto;
		padding: 56px clamp(16px, 3vw, 28px) 28px;
	}
	.cols {
		display: flex;
		gap: 36px;
		flex-wrap: wrap;
	}
	.marca {
		flex: 1 1 240px;
		max-width: 320px;
	}
	.marca img {
		height: 22px;
		width: auto;
		display: block;
	}
	.marca p {
		margin: 14px 0 0;
		font-size: 14px;
		line-height: 1.55;
		color: #5c4536;
	}
	.col {
		flex: 1 1 150px;
		display: flex;
		flex-direction: column;
		gap: 9px;
		font-size: 14px;
	}
	.col-t {
		font-size: 13px;
		font-weight: 700;
		color: #2a140c;
		margin-bottom: 2px;
	}
	.col a {
		color: #5c4536;
	}
	.col a:hover {
		color: #2a140c;
	}
	.barra {
		margin-top: 40px;
		padding-top: 18px;
		border-top: 1px solid rgba(58, 29, 18, 0.1);
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 12px 24px;
		flex-wrap: wrap;
		font-size: 12.5px;
		color: #7a6252;
	}
	.barra b {
		color: #9a6a10;
	}
	.legal {
		display: flex;
		gap: 8px 16px;
		flex-wrap: wrap;
		align-items: center;
	}
	.legal a {
		color: #5c4536;
		text-decoration: underline;
	}
	.legal .arrep {
		color: #2a140c;
		font-weight: 700;
		display: inline-flex;
		align-items: center;
		gap: 5px;
	}
	.legal button {
		border: 0;
		background: transparent;
		padding: 0;
		font: inherit;
		font-weight: 500;
		color: #5c4536;
		text-decoration: underline;
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
	@media (max-width: 640px) {
		.cols {
			gap: 28px 20px;
		}
		.col {
			flex: 1 1 140px;
		}
	}
	.nw {
		white-space: nowrap;
	}
</style>
