<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { enhance } from '$app/forms';
	import Grano from '#lib/components/Grano.svelte';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { CONSEJOS, ILUSTRACION, NOTA_PIE, PASOS, PORCENTAJE } from '#lib/onboarding/datos.ts';
	import { Onboarding, setOnboarding } from '#lib/onboarding/estado.svelte.ts';
	import PasoCanales from '#lib/onboarding/PasoCanales.svelte';
	import PasoCobros from '#lib/onboarding/PasoCobros.svelte';
	import PasoCuenta from '#lib/onboarding/PasoCuenta.svelte';
	import PasoHornero from '#lib/onboarding/PasoHornero.svelte';
	import PasoImportacion from '#lib/onboarding/PasoImportacion.svelte';
	import PasoListo from '#lib/onboarding/PasoListo.svelte';
	import PasoNegocio from '#lib/onboarding/PasoNegocio.svelte';
	import PasoTienda from '#lib/onboarding/PasoTienda.svelte';
	import type { PageProps } from './$types';

	/** Onboarding de 8 pasos (design/Onboarding Mostrador.dc.html). */
	let { data, form }: PageProps = $props();
	const o = setOnboarding(new Onboarding());
	const paso = $derived(o.e.paso);

	let listo = $state(false);
	onMount(() => {
		o.cargar();
		// Con la API, sin sesión no se puede pasar del paso 1; al volver del registro, ?paso= manda.
		if (data.paso !== null) o.ir(data.paso);
		else if (!data.demo && !data.conSesion) o.ir(0);
		listo = true;
		return () => o.detener();
	});

	// Con la API, cada cambio de paso se guarda en la tienda.
	let formPaso = $state<HTMLFormElement>();
	let formFin = $state<HTMLFormElement>();
	$effect(() => {
		void paso;
		if (listo && !data.demo && data.conSesion && paso > 0)
			tick().then(() => formPaso?.requestSubmit());
	});
	const estado = $derived(
		JSON.stringify({
			plantilla: o.e.plantilla,
			elecciones: {
				objetivos: o.e.objetivos,
				conexiones: o.e.conexiones,
				conflicto: o.e.conflicto,
				instagram: o.e.instagram,
				secciones: o.e.secciones,
				config: o.e.config,
				tono: o.e.tono,
				caja: o.e.caja,
				equipo: o.e.equipo
			}
		})
	);
	const terminar = () => {
		try {
			localStorage.setItem('ob-step', '0');
		} catch {
			// Sin almacenamiento: no importa.
		}
		formFin?.requestSubmit();
	};

	const siguiente = () => o.ir(paso + 1);
	const etiquetaSiguiente = $derived(
		paso === 3 && !o.importacionLista
			? 'Seguir mientras importa'
			: paso === 6
				? 'Terminar'
				: 'Continuar'
	);
	const ilustracion = $derived(`/hornero/${ILUSTRACION[paso]}.webp`);

	// Se anima cada vez que cambia (como en el handoff).
	let caja = $state<HTMLElement>();
	$effect(() => {
		void paso;
		caja?.animate(
			[
				{ opacity: 0, transform: 'translateY(16px)' },
				{ opacity: 1, transform: 'none' }
			],
			{ duration: 520, easing: 'cubic-bezier(.2,.7,.1,1)' }
		);
	});
</script>

<Seo
	title="Creá tu tienda en Mostrador"
	description="Importamos tus productos y te dejamos la tienda lista en 10 minutos."
	path="/onboarding"
/>

<!-- eslint-disable svelte/no-navigation-without-resolve -- inicio y ayuda son rutas públicas fijas -->
<div class="ob">
	<Grano />
	<aside>
		<div class="arcos"></div>
		<div class="luz"></div>
		<a class="marca" href="/" aria-label="Mostrador, inicio">
			<svg viewBox="0 0 100 76" aria-hidden="true"
				><path
					fill="#E8875F"
					d="M3 76a3 3 0 0 1-3-3V5a5 5 0 0 1 5-5h90a5 5 0 0 1 5 5v68a3 3 0 0 1-3 3H82a3 3 0 0 1-3-3V21H21v52a3 3 0 0 1-3 3Z"
				></path></svg
			>
			<img src="/wordmark-cream.webp" alt="mostrador" width="86" height="18" />
		</a>
		<div class="avance">
			<div class="a-t">Tu cuenta está</div>
			<div class="a-v"><span>{PORCENTAJE[paso]}%</span><span class="a-l">lista</span></div>
			<div class="a-barra"><div style:width="{PORCENTAJE[paso]}%"></div></div>
			<div class="a-d">Sin cargar nada a mano: lo traemos de lo que ya usás.</div>
		</div>
		<nav class="pasos" aria-label="Pasos">
			{#each PASOS as label, i (label)}
				<button
					type="button"
					class:cur={i === paso}
					class:prox={i > paso}
					aria-current={i === paso ? 'step' : undefined}
					disabled={i > paso + 1 || (!data.demo && !data.conSesion && i > 0)}
					onclick={() => o.ir(i)}
				>
					{#if i < paso}
						<span class="n hecho"><Icon icon={I.phCheckBold} size={12} /></span>
					{:else}
						<span class="n">{i + 1}</span>
					{/if}
					<span class="p-l">{label}</span>
				</button>
			{/each}
		</nav>
		<div class="esp"></div>
		<div class="consejo">
			<div class="c-av"><img src={ilustracion} alt="Hornero" width="52" height="52" /></div>
			<div>
				<div class="c-n">Hornero</div>
				<div class="c-t">{CONSEJOS[paso]}</div>
			</div>
		</div>
	</aside>

	<main>
		<div class="movil">
			<div class="m-fila">
				<svg viewBox="0 0 100 76" aria-hidden="true"
					><path
						fill="#E8875F"
						d="M3 76a3 3 0 0 1-3-3V5a5 5 0 0 1 5-5h90a5 5 0 0 1 5 5v68a3 3 0 0 1-3 3H82a3 3 0 0 1-3-3V21H21v52a3 3 0 0 1-3 3Z"
					></path></svg
				>
				<span class="m-paso"
					><span class="m-n">Paso {paso + 1} de 8</span><span class="m-l">{PASOS[paso]}</span></span
				>
				<span class="m-pct">{PORCENTAJE[paso]}%</span>
				<button
					type="button"
					class="m-av"
					aria-label="Consejo de Hornero"
					aria-expanded={o.consejo}
					onclick={() => (o.consejo = !o.consejo)}><img src={ilustracion} alt="" /></button
				>
			</div>
			<div class="m-barras">
				{#each PASOS as label, i (label)}
					<button
						type="button"
						aria-label={label}
						class:hecho={i < paso}
						class:cur={i === paso}
						disabled={i > paso + 1 || (!data.demo && !data.conSesion && i > 0)}
						onclick={() => o.ir(i)}
					></button>
				{/each}
			</div>
			{#if o.consejo}<div class="m-consejo"><b>Hornero:</b> {CONSEJOS[paso]}</div>{/if}
		</div>
		<div class="arriba">
			<span class="ar-paso">Paso {paso + 1} de 8</span>
			<span class="ar-ok"><Icon icon={I.phCloudCheckFill} size={16} />Guardado automáticamente</span
			>
			<span class="esp"></span>
			<a href="/ayuda">¿Necesitás ayuda?</a>
			<a href="/">Salir</a>
		</div>

		<div class="contenido">
			<div class="paso" bind:this={caja}>
				{#if paso === 0}
					<PasoCuenta demo={data.demo} seguir={siguiente} />
				{:else if paso === 1}
					<PasoNegocio />
				{:else if paso === 2}
					<PasoCanales />
				{:else if paso === 3}
					<PasoImportacion />
				{:else if paso === 4}
					<PasoTienda />
				{:else if paso === 5}
					<PasoCobros />
				{:else if paso === 6}
					<PasoHornero />
				{:else}
					<PasoListo {terminar} />
					{#if form && 'error' in form}<div class="error" role="alert">{form.error}</div>{/if}
				{/if}
			</div>
		</div>

		{#if paso > 0 && paso < 7}
			<div class="pie">
				{#if paso > 1}
					<button type="button" class="atras" onclick={() => o.ir(paso - 1)}
						><Icon icon={I.phArrowLeftBold} size={14} /><span class="atras-t">Atrás</span></button
					>
				{/if}
				<div class="nota">{NOTA_PIE[paso] ?? 'Todo se guarda solo.'}</div>
				<button type="button" class="saltar" onclick={siguiente}
					><span class="largo">Hacerlo después</span><span class="corto">Después</span></button
				>
				<button type="button" class="seguir" onclick={siguiente}
					>{etiquetaSiguiente}<Icon icon={I.phArrowRightBold} size={15} /></button
				>
			</div>
		{/if}
	</main>
</div>

<form method="POST" action="?/paso" bind:this={formPaso} use:enhance={() => async () => {}} hidden>
	<input type="hidden" name="estado" value={estado} />
</form>
<form method="POST" action="?/terminar" bind:this={formFin} use:enhance hidden>
	<input type="hidden" name="estado" value={estado} />
</form>

<style>
	.ob {
		font-variant-numeric: tabular-nums;
		color: #2a140c;
		background: #fbf7f2;
		min-height: 100vh;
		display: flex;
	}
	aside {
		width: 340px;
		flex: none;
		position: sticky;
		top: 0;
		height: 100vh;
		background: #2a140c;
		color: #fbf7f2;
		display: flex;
		flex-direction: column;
		padding: 28px 24px;
		overflow: hidden;
	}
	.arcos {
		position: absolute;
		inset: 0;
		background-image: url('/textura/arch-pattern-light.svg');
	}
	.luz {
		position: absolute;
		inset: 0;
		background: radial-gradient(420px 320px at 0% 100%, rgba(210, 88, 47, 0.35), transparent 70%);
	}
	.marca {
		position: relative;
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.marca svg {
		width: 24px;
		height: 18px;
		display: block;
	}
	.marca img {
		height: 18px;
		width: auto;
		display: block;
	}
	.avance {
		position: relative;
		margin-top: 36px;
	}
	.a-t {
		font-size: 13.5px;
		color: #d9c8ba;
	}
	.a-v {
		display: flex;
		align-items: baseline;
		gap: 8px;
		margin-top: 2px;
	}
	.a-v > span:first-child {
		font-size: 52px;
		font-weight: 600;
		letter-spacing: -0.05em;
		line-height: 1;
	}
	.a-l {
		font-size: 16px;
		font-weight: 600;
		color: #f5b796;
	}
	.a-barra {
		margin-top: 12px;
		height: 8px;
		border-radius: 8px;
		background: rgba(251, 247, 242, 0.12);
		overflow: hidden;
	}
	.a-barra div {
		height: 100%;
		border-radius: 8px;
		background: linear-gradient(90deg, #d2582f, #f5b796);
		transition: width 0.9s cubic-bezier(0.2, 0.7, 0.1, 1);
	}
	.a-d {
		margin-top: 8px;
		font-size: 12.5px;
		color: #c9b4a4;
	}
	.pasos {
		position: relative;
		margin-top: 28px;
		display: flex;
		flex-direction: column;
		gap: 2px;
	}
	.pasos button {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 9px 10px;
		border-radius: 12px;
		border: 0;
		background: transparent;
		color: #fbf7f2;
		text-align: left;
	}
	.pasos button.cur {
		background: rgba(251, 247, 242, 0.1);
	}
	.pasos button.prox {
		color: #c9b4a4;
	}
	.pasos button:disabled {
		cursor: default;
	}
	.n {
		width: 24px;
		height: 24px;
		flex: none;
		border-radius: 50%;
		display: grid;
		place-items: center;
		font-size: 12px;
		font-weight: 600;
		border: 1.5px solid rgba(251, 247, 242, 0.25);
		color: #c9b4a4;
	}
	.cur .n {
		border: 0;
		background: #b0441d;
		color: #fff;
		font-weight: 700;
		box-shadow: 0 0 0 4px rgba(210, 88, 47, 0.25);
	}
	.n.hecho {
		border: 0;
		background: #2f8f63;
		color: #fff;
	}
	.p-l {
		flex: 1;
		font-size: 14.5px;
		font-weight: 500;
	}
	.cur .p-l {
		font-weight: 600;
	}
	.esp {
		flex: 1;
	}
	.consejo {
		position: relative;
		display: flex;
		gap: 12px;
		align-items: flex-start;
		padding: 14px;
		border-radius: 18px;
		background: rgba(251, 247, 242, 0.07);
		border: 1px solid rgba(251, 247, 242, 0.12);
	}
	.c-av {
		width: 56px;
		height: 56px;
		flex: none;
		border-radius: 50%;
		overflow: hidden;
		background: #fbf7f2;
		box-shadow: 0 0 0 2px #e8875f;
		display: grid;
		place-items: center;
	}
	.c-av img {
		display: block;
		object-fit: contain;
	}
	.c-n {
		font-size: 13.5px;
		font-weight: 600;
	}
	.c-t {
		font-size: 13px;
		line-height: 1.45;
		color: #eadbcf;
		margin-top: 2px;
	}
	main {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		min-height: 100vh;
	}
	.movil {
		display: none;
	}
	.arriba {
		height: 68px;
		display: flex;
		align-items: center;
		gap: 14px;
		padding: 0 40px;
		border-bottom: 1px solid rgba(58, 29, 18, 0.08);
	}
	.ar-paso {
		font-size: 14px;
		color: #7a6252;
	}
	.ar-ok {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 13px;
		color: #237550;
		font-weight: 600;
	}
	.arriba a {
		font-size: 14px;
		font-weight: 600;
		color: #5c4536;
	}
	.arriba a:hover {
		color: #2a140c;
	}
	.contenido {
		flex: 1;
		padding: 44px 40px 40px;
		display: flex;
		justify-content: center;
	}
	.paso {
		width: 100%;
		max-width: 880px;
	}
	.error {
		margin-top: 14px;
		text-align: center;
		font-size: 14px;
		color: #b83a2e;
	}
	.pie {
		position: sticky;
		bottom: 0;
		z-index: 20;
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 16px 40px;
		background: rgba(251, 247, 242, 0.9);
		backdrop-filter: blur(12px);
		-webkit-backdrop-filter: blur(12px);
		border-top: 1px solid rgba(58, 29, 18, 0.08);
	}
	.atras {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 12px 18px;
		border-radius: 999px;
		border: 1px solid rgba(58, 29, 18, 0.16);
		background: #fff;
		color: #2a140c;
		font-size: 14.5px;
		font-weight: 600;
	}
	.nota {
		flex: 1;
		min-width: 0;
		font-size: 13.5px;
		color: #7a6252;
	}
	.saltar {
		border: 0;
		background: transparent;
		font-size: 14px;
		font-weight: 600;
		color: #7a6252;
		padding: 12px 6px;
		white-space: nowrap;
	}
	.corto {
		display: none;
	}
	.seguir {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 13px 24px;
		border-radius: 999px;
		border: 0;
		background: linear-gradient(180deg, #c24e25, #b0441d);
		color: #fff;
		font-size: 15px;
		font-weight: 600;
		white-space: nowrap;
		box-shadow:
			inset 0 1px 0 rgba(255, 255, 255, 0.3),
			0 12px 26px -12px rgba(210, 88, 47, 0.7);
	}
	@media (max-width: 1099px) {
		aside {
			width: 280px;
		}
	}
	/* Angosto (handoff: < 900 px): sin lateral; arriba, el paso y las barras de avance. */
	@media (max-width: 899px) {
		aside,
		.arriba,
		.nota {
			display: none;
		}
		.movil {
			display: block;
			position: sticky;
			top: 0;
			z-index: 20;
			background: #2a140c;
			color: #fbf7f2;
		}
		.m-fila {
			height: 58px;
			display: flex;
			align-items: center;
			gap: 10px;
			padding: 0 14px;
		}
		.m-fila svg {
			width: 22px;
			height: 17px;
			flex: none;
		}
		.m-paso {
			flex: 1;
			min-width: 0;
		}
		.m-n {
			display: block;
			font-size: 12px;
			color: #c9b4a4;
		}
		.m-l {
			display: block;
			font-size: 15px;
			font-weight: 600;
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
		}
		.m-pct {
			font-size: 20px;
			font-weight: 600;
			letter-spacing: -0.03em;
		}
		.m-av {
			width: 40px;
			height: 40px;
			flex: none;
			border-radius: 50%;
			border: 0;
			padding: 0;
			overflow: hidden;
			background: #fbf7f2;
			box-shadow: 0 0 0 2px #e8875f;
		}
		.m-av img {
			width: 100%;
			height: 100%;
			object-fit: contain;
		}
		.m-barras {
			display: flex;
			gap: 3px;
			padding: 0 14px 10px;
		}
		.m-barras button {
			flex: 1;
			height: 5px;
			padding: 0;
			border: 0;
			border-radius: 3px;
			background: rgba(251, 247, 242, 0.18);
		}
		.m-barras button.hecho {
			background: #2f8f63;
		}
		.m-barras button.cur {
			background: #e8875f;
		}
		.m-consejo {
			margin: 0 12px 12px;
			padding: 12px 14px;
			border-radius: 14px;
			background: rgba(251, 247, 242, 0.08);
			border: 1px solid rgba(251, 247, 242, 0.14);
			font-size: 13.5px;
			line-height: 1.45;
			color: #eadbcf;
		}
		.m-consejo b {
			color: #fff;
		}
		.contenido {
			padding: 28px 22px 32px;
		}
		.pie {
			padding: 12px 18px;
		}
		.saltar {
			margin-left: auto;
		}
	}
	@media (max-width: 559px) {
		.contenido {
			padding: 22px 14px 28px;
		}
		.pie {
			gap: 6px;
			padding: 10px 12px calc(10px + env(safe-area-inset-bottom));
		}
		.atras {
			padding: 12px 14px;
		}
		.atras-t,
		.largo {
			display: none;
		}
		.corto {
			display: inline;
		}
	}
</style>
