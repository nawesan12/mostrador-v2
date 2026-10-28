<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import Icon from '#lib/components/Icon.svelte';
	import Vacio from '#lib/components/Vacio.svelte';
	import * as I from '#lib/icons.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { deslizable } from '#lib/panel/deslizable.ts';
	import { nivel } from '#lib/partners/programa.ts';

	/**
	 * Panel de partners (/partners/panel). Sin diseño propio en el handoff: armado con las piezas del
	 * panel de la tienda (barra de arriba del armazón, pestañas y tarjetas de Configuración, KPIs de
	 * Cupones, filas de MotorLista). Los estilos compartidos de las pantallas viven acá.
	 */
	let { data, children } = $props();

	const SECCIONES = [
		['/partners/panel', I.phHouseFill, 'Inicio'],
		['/partners/panel/tiendas', I.phStorefrontFill, 'Mis tiendas'],
		['/partners/panel/link', I.phQrCodeFill, 'Link y cupón'],
		['/partners/panel/comisiones', I.phHandCoinsFill, 'Comisiones y pagos'],
		['/partners/panel/perfil', I.phIdentificationCardFill, 'Perfil público'],
		['/partners/panel/materiales', I.phImagesFill, 'Materiales']
	] as const;
	const actual = $derived(
		SECCIONES.find(([r]) => r === page.url.pathname.replace(/\/$/, ''))?.[2] ?? 'Inicio'
	);
	const p = $derived(data.panel);
	const primerNombre = $derived(p?.partner.nombre.split(' ')[0] ?? '');
</script>

<svelte:head>
	<title>{actual} · Partners · Mostrador</title>
	<meta name="robots" content="noindex" />
	<meta name="theme-color" content="#f7f1ea" />
</svelte:head>

<div class="socio">
	<div class="luz" aria-hidden="true"></div>
	<header class="top">
		<a class="marca" href={resolve('/partners/panel')} aria-label="Inicio del panel de partners">
			<svg viewBox="0 0 100 76" aria-hidden="true"
				><path
					fill="#FBF3EA"
					d="M3 76a3 3 0 0 1-3-3V5a5 5 0 0 1 5-5h90a5 5 0 0 1 5 5v68a3 3 0 0 1-3 3H82a3 3 0 0 1-3-3V21H21v52a3 3 0 0 1-3 3Z"
				></path></svg
			>
		</a>
		<nav class="miga" aria-label="Ubicación">
			<span class="area">Partners</span><Icon icon={I.phCaretRight} size={12} /><span class="sec"
				>{actual}</span
			>
		</nav>
		<span class="esp"></span>
		{#if data.demo}
			<span class="en-linea"><span class="punto"></span>Ejemplo</span>
		{/if}
		<a class="ico" href={resolve('/(sitio)/partners')} aria-label="Volver al programa de partners"
			><Icon icon={I.phArrowLeftBold} size={16} /></a
		>
	</header>

	<div class="contenido">
		{#if !p}
			<div class="no-socio">
				<Vacio
					hornero="pensando"
					titulo="Tu cuenta todavía no es partner"
					texto="Postulate al programa y, cuando te aprobemos, acá vas a ver tus tiendas, tu link y tus comisiones."
				>
					<a class="btn btn-primary" href="{resolve('/(sitio)/partners')}#sumate"
						>Quiero ser partner</a
					>
				</Vacio>
			</div>
		{:else}
			<div class="pagina">
				<div class="saludo" use:anim={{ t: 'rise' }}>
					<h1>Hola, {primerNombre}</h1>
					<div class="c-s">
						{p.partner.perfil.nombre} · {nivel(p.partner.nivelEfectivo).nombre} · {p.partner
							.porcentaje}&nbsp;% de&nbsp;comisión
					</div>
				</div>
				<div class="grilla" use:anim={{ t: 'rise', delay: 60 }}>
					<nav class="tabs" aria-label="Secciones" use:deslizable>
						{#each SECCIONES as [ruta, icono, t] (ruta)}
							<a
								href={resolve(ruta)}
								aria-current={actual === t ? 'page' : undefined}
								data-sveltekit-noscroll><Icon icon={icono} size={17} />{t}</a
							>
						{/each}
					</nav>
					<div class="cuerpo">
						{@render children()}
					</div>
				</div>
			</div>
		{/if}
	</div>
</div>

<style>
	/* Armazón: fondo, luz y barra de arriba del panel (Topbar), marca del Rail. */
	.socio {
		position: relative;
		min-height: 100vh;
		min-height: 100dvh;
		background: var(--bg);
		color: var(--tx);
		overflow-x: clip;
	}
	.luz {
		position: absolute;
		right: 0;
		top: 0;
		width: 700px;
		max-width: 100%;
		height: 420px;
		background: radial-gradient(closest-side, rgba(232, 135, 95, 0.16), transparent);
		pointer-events: none;
	}
	.top {
		position: sticky;
		top: 0;
		z-index: 10;
		height: 64px;
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 0 36px;
		background: var(--bga);
		backdrop-filter: blur(14px);
		-webkit-backdrop-filter: blur(14px);
	}
	.marca {
		width: 40px;
		height: 40px;
		flex: none;
		border-radius: 12px;
		background: linear-gradient(160deg, #e8875f, #d2582f 55%, #b9481f);
		display: grid;
		place-items: center;
		box-shadow:
			inset 0 1px 0 rgba(255, 255, 255, 0.3),
			0 8px 18px -8px rgba(0, 0, 0, 0.5);
	}
	.marca svg {
		width: 20px;
		height: 15px;
	}
	.miga {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 14px;
		color: var(--mu);
		white-space: nowrap;
		min-width: 0;
	}
	.sec {
		color: var(--tx);
		font-weight: 600;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.esp {
		flex: 1;
	}
	.en-linea {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 13px;
		font-weight: 500;
		color: var(--mu);
		padding: 7px 12px;
		border-radius: 999px;
		background: var(--sf);
		box-shadow: var(--sh);
		white-space: nowrap;
	}
	.punto {
		width: 7px;
		height: 7px;
		border-radius: 50%;
		background: #e3a93b;
	}
	.ico {
		width: 38px;
		height: 38px;
		flex: none;
		border-radius: 999px;
		background: var(--sf);
		color: var(--tx);
		display: grid;
		place-items: center;
		box-shadow: var(--sh);
	}
	.contenido {
		position: relative;
		padding: 20px 36px 120px;
		max-width: 1440px;
	}
	.no-socio {
		max-width: 460px;
		margin: 40px auto;
	}

	/* Página con pestañas (Configuración, sCfg). */
	.pagina {
		display: flex;
		flex-direction: column;
		gap: 18px;
	}
	h1 {
		margin: 0;
		font-size: 30px;
		letter-spacing: -0.035em;
		font-weight: 600;
	}
	.c-s {
		margin-top: 2px;
		font-size: 15px;
		color: var(--mu);
	}
	.grilla {
		display: grid;
		grid-template-columns: 240px minmax(0, 1fr);
		gap: 16px;
		align-items: start;
	}
	.tabs {
		padding: 8px;
		border-radius: 22px;
		background: var(--sf);
		box-shadow: var(--shd);
		display: flex;
		flex-direction: column;
		gap: 2px;
		overflow-x: auto;
		scrollbar-width: none;
	}
	.tabs a {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 12px;
		border-radius: 14px;
		color: var(--tx);
		font:
			500 14px 'Instrument Sans',
			sans-serif;
		text-decoration: none;
		white-space: nowrap;
	}
	.tabs a :global(svg) {
		color: var(--mu);
	}
	.tabs a[aria-current='page'] {
		background: var(--chip);
		font-weight: 600;
	}
	.tabs a[aria-current='page'] :global(svg) {
		color: #b0441d;
	}
	.cuerpo {
		display: flex;
		flex-direction: column;
		gap: 12px;
		min-width: 0;
	}

	/* ---- Piezas compartidas por las pantallas ---- */
	/* Tarjeta y encabezado (Configuración). */
	.socio :global(.tarjeta) {
		padding: 18px;
		border-radius: 22px;
		background: var(--sf);
		box-shadow: var(--shd);
		min-width: 0;
	}
	.socio :global(.t-cab) {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 10px;
		flex-wrap: wrap;
		font-size: 16px;
		font-weight: 600;
	}
	.socio :global(.t-d) {
		margin-top: 2px;
		font-size: 13px;
		font-weight: 500;
		color: var(--mu);
	}
	.socio :global(.oscuro) {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		padding: 9px 14px;
		border-radius: 999px;
		border: 0;
		background: var(--tx);
		color: var(--sf);
		font:
			600 13px 'Instrument Sans',
			sans-serif;
		text-decoration: none;
		white-space: nowrap;
	}
	.socio :global(.mas) {
		margin-top: 12px;
		display: inline-flex;
		align-items: center;
		gap: 6px;
		font-size: 13.5px;
		font-weight: 600;
		color: #b0441d;
		text-decoration: none;
		border: 0;
		background: transparent;
		padding: 0;
		font-family: inherit;
	}
	.socio :global(.t-cab .mas) {
		margin-top: 0;
	}
	/* KPIs (Cupones). */
	.socio :global(.kpis) {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(190px, 1fr));
		gap: 12px;
	}
	.socio :global(.kpi) {
		padding: 16px 18px;
		border-radius: 20px;
		background: var(--sf);
		box-shadow: var(--shd);
		min-width: 0;
	}
	.socio :global(.k-t) {
		display: flex;
		align-items: center;
		gap: 7px;
		font-size: 13px;
		color: var(--mu);
	}
	.socio :global(.k-v) {
		margin-top: 4px;
		font-size: 26px;
		font-weight: 600;
		letter-spacing: -0.03em;
		white-space: nowrap;
	}
	.socio :global(.k-d) {
		font-size: 12.5px;
		font-weight: 500;
		color: var(--mu);
	}
	/* Filas de lista (MotorLista). */
	.socio :global(.filas) {
		margin-top: 8px;
		display: flex;
		flex-direction: column;
	}
	.socio :global(.f-in) {
		min-height: 66px;
		display: flex;
		align-items: center;
		gap: 14px;
		padding: 0 12px;
		margin: 0 -12px;
		border-radius: 16px;
		border: 0;
		background: transparent;
		color: var(--tx);
		text-align: left;
		font: inherit;
		text-decoration: none;
		transition: background 0.15s;
	}
	.socio :global(button.f-in:hover),
	.socio :global(a.f-in:hover) {
		background: var(--chip);
	}
	.socio :global(.ini) {
		width: 42px;
		height: 42px;
		flex: none;
		border-radius: 12px;
		display: grid;
		place-items: center;
		font-size: 14px;
		font-weight: 700;
	}
	.socio :global(.quien) {
		flex: 1 1 150px;
		min-width: 0;
	}
	.socio :global(.q-t) {
		display: block;
		font-size: 14px;
		font-weight: 600;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.socio :global(.q-s) {
		display: block;
		font-size: 12px;
		color: var(--fa);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.socio :global(.chip-f) {
		flex: none;
		font-size: 11.5px;
		font-weight: 700;
		padding: 4px 8px;
		border-radius: 999px;
		white-space: nowrap;
	}
	.socio :global(.valor) {
		flex: none;
		min-width: 80px;
		text-align: right;
	}
	.socio :global(.v-v) {
		display: block;
		font-size: 15px;
		font-weight: 600;
		letter-spacing: -0.01em;
		white-space: nowrap;
	}
	.socio :global(.v-s) {
		display: block;
		font-size: 11.5px;
		color: var(--mu);
		white-space: nowrap;
	}
	/* Medidor (MotorLista, stock). */
	.socio :global(.medidor) {
		position: relative;
		display: block;
		height: 8px;
		border-radius: 4px;
		background: var(--sf3);
	}
	.socio :global(.medidor > span) {
		position: absolute;
		left: 0;
		top: 0;
		bottom: 0;
		border-radius: 4px;
		background: #d2582f;
	}
	/* Opciones con ícono (Configuración, seguridad). */
	.socio :global(.opcion) {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 12px 0;
		border-bottom: 1px solid var(--ln);
	}
	.socio :global(.opcion:last-child) {
		border-bottom: 0;
	}
	.socio :global(.o-ico) {
		width: 36px;
		height: 36px;
		flex: none;
		border-radius: 11px;
		background: rgba(47, 143, 99, 0.12);
		color: #237550;
		display: grid;
		place-items: center;
	}
	.socio :global(.o-txt) {
		flex: 1;
		min-width: 0;
	}
	.socio :global(.o-t) {
		display: block;
		font-size: 14.5px;
		font-weight: 600;
	}
	.socio :global(.o-d) {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
	}
	/* Campos (Configuración, negocio). */
	.socio :global(.campos) {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 220px), 1fr));
		gap: 12px;
	}
	.socio :global(.campos label),
	.socio :global(.campo-t) {
		display: flex;
		flex-direction: column;
		gap: 5px;
		font-size: 12.5px;
		font-weight: 600;
		color: var(--mu);
	}
	.socio :global(.campos input),
	.socio :global(.campos textarea) {
		padding: 11px 12px;
		border-radius: 12px;
		border: 1px solid var(--ln2);
		background: var(--bg);
		font:
			500 14.5px 'Instrument Sans',
			sans-serif;
		color: var(--tx);
		outline: none;
		min-width: 0;
	}
	.socio :global(.campos textarea) {
		resize: vertical;
		min-height: 74px;
	}
	.socio :global(.ancho) {
		grid-column: 1 / -1;
	}
	.socio :global(.guardar) {
		grid-column: 1 / -1;
		display: flex;
		justify-content: flex-end;
		gap: 8px;
		flex-wrap: wrap;
	}
	/* Chips de selección (permisos de Configuración). */
	.socio :global(.chips) {
		display: flex;
		gap: 6px;
		flex-wrap: wrap;
	}
	.socio :global(.chips button) {
		font:
			600 12.5px 'Instrument Sans',
			sans-serif;
		padding: 7px 12px;
		border-radius: 999px;
		border: 1.5px solid transparent;
		background: var(--chip);
		color: var(--tx);
	}
	.socio :global(.chips button[aria-pressed='true']) {
		border-color: #d2582f;
		background: rgba(210, 88, 47, 0.06);
		color: #b0441d;
	}
	/* Modal (Configuración, invitar). El modal se monta en el body: su contenido va en un .pm. */
	:global(.pm .roles) {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	:global(.pm .roles button) {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		gap: 2px;
		padding: 10px 12px;
		border-radius: 14px;
		border: 1.5px solid var(--ln2);
		background: transparent;
		color: var(--tx);
		text-align: left;
		font:
			500 13px 'Instrument Sans',
			sans-serif;
	}
	:global(.pm .roles button span) {
		color: var(--mu);
	}
	:global(.pm .roles button[aria-pressed='true']) {
		border-color: #d2582f;
		background: rgba(210, 88, 47, 0.06);
	}
	:global(.pm .m-t) {
		font-size: 21px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}
	:global(.pm .m-d) {
		margin-top: 6px;
		font-size: 14.5px;
		color: var(--mu);
	}
	:global(.pm .m-campo) {
		margin-top: 16px;
	}
	:global(.pm .m-error) {
		margin-top: 10px;
		font-size: 13px;
		font-weight: 600;
		color: #b83a2e;
	}
	:global(.pm .m-acc) {
		margin-top: 20px;
		display: flex;
		gap: 8px;
		justify-content: flex-end;
		flex-wrap: wrap;
	}

	@media (max-width: 1099px) {
		.top {
			padding: 0 24px;
		}
		.contenido {
			padding: 20px 24px 120px;
		}
	}
	@media (max-width: 899px) {
		.grilla {
			grid-template-columns: minmax(0, 1fr);
		}
		.tabs {
			flex-direction: row;
		}
	}
	@media (max-width: 759px) {
		.top {
			padding: 0 12px;
			gap: 8px;
		}
		.contenido {
			padding: 14px 12px 60px;
		}
		.area {
			display: none;
		}
		.miga :global(svg) {
			display: none;
		}
		.socio :global(.kpis) {
			grid-template-columns: 1fr 1fr;
		}
		.socio :global(.kpi) {
			padding: 14px;
		}
		.socio :global(.k-v) {
			font-size: 21px;
		}
		.socio :global(.f-in) {
			gap: 10px;
		}
		/* En el celular el detalle de la fila baja de renglón en vez de cortarse. */
		.socio :global(.q-s) {
			white-space: normal;
		}
	}
</style>
