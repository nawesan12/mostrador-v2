<script lang="ts">
	import { PLANTILLAS } from '#lib/plantillas.ts';
	import { page } from '$app/state';
	import Marca from '#lib/landing/Marca.svelte';
	import { onMount, tick } from 'svelte';
	import { deserialize, enhance } from '$app/forms';
	import { invalidateAll, pushState, replaceState } from '$app/navigation';
	import { resolve } from '$app/paths';
	import Grano from '#lib/components/Grano.svelte';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import {
		CONSEJOS,
		CONSEJOS_REALES,
		ILUSTRACION,
		NOTA_PIE,
		NOTA_PIE_REAL,
		PORCENTAJE,
		ROTULO
	} from '#lib/onboarding/datos.ts';
	import { Onboarding, setOnboarding } from '#lib/onboarding/estado.svelte.ts';
	import { slugFinal } from '#lib/tienda/direccion.ts';
	import PasoCanales from '#lib/onboarding/PasoCanales.svelte';
	import PasoCobros from '#lib/onboarding/PasoCobros.svelte';
	import PasoCuenta from '#lib/onboarding/PasoCuenta.svelte';
	import PasoHornero from '#lib/onboarding/PasoHornero.svelte';
	import PasoImportacion from '#lib/onboarding/PasoImportacion.svelte';
	import PasoListo from '#lib/onboarding/PasoListo.svelte';
	import PasoNegocio from '#lib/onboarding/PasoNegocio.svelte';
	import PasoTienda from '#lib/onboarding/PasoTienda.svelte';
	import type { PageProps } from './$types';

	/**
	 * Onboarding (design/Onboarding Mostrador.dc.html): los 8 pasos del handoff en la demo; con una
	 * cuenta real, sólo los que sirven para tener la tienda andando y compartida (datos.ts).
	 */
	let { data, form }: PageProps = $props();
	const o = setOnboarding(new Onboarding());
	const paso = $derived(o.e.paso);
	const id = $derived(o.id);
	const total = $derived(o.pasos.length);
	// El avance: el del handoff en la demo; con la API, los pasos hechos de los que hay (honesto).
	const porcentaje = $derived(
		data.demo ? PORCENTAJE[paso] : Math.round(((paso + 1) / total) * 100)
	);
	const consejo = $derived(data.demo ? CONSEJOS[paso] : (CONSEJOS_REALES[id] ?? ''));

	let listo = $state(false);
	onMount(() => {
		o.demo = data.demo;
		o.fijarTienda(data.tienda);
		o.cargar(data.demo || !data.conSesion ? undefined : data.guardado);
		o.alCambiar = () => formPaso?.requestSubmit();
		// Desde «Usar esta plantilla» de la landing: ?plantilla= manda sobre lo guardado.
		const tpl = page.url.searchParams.get('plantilla');
		if (tpl && PLANTILLAS.some((p) => p.id === tpl)) {
			o.e.plantilla = tpl;
			o.guardar();
		}
		// Desde /migrar: la plataforma (?desde=) y el link (?origen=) ya elegidos para el paso 4.
		const desde = page.url.searchParams.get('desde');
		if (!data.demo && (desde === 'tiendanube' || desde === 'shopify' || desde === 'woocommerce')) {
			o.e.importar = { plataforma: desde, url: page.url.searchParams.get('origen') ?? '' };
			o.guardar();
		}
		// Con la API, sin sesión no se puede pasar del paso 1 y con sesión no se vuelve a él; si no, al
		// volver del registro o de conectar algo, ?paso= manda.
		if (!data.demo && !data.conSesion) o.ir(0);
		else if (data.paso !== null) o.ir(!data.demo && data.paso === 0 ? 1 : data.paso);
		listo = true;
		return () => o.detener();
	});

	// El paso queda en la dirección (?paso=N): al recargar se vuelve al mismo, no al que trajo el
	// registro (antes quedaba fijo en ?paso=1 y recargar en el 6 mandaba al 2).
	// Cada paso nuevo es una entrada del historial: el «atrás» del celular (o del navegador) vuelve
	// al paso anterior en vez de salir del onboarding. El primero sólo se reemplaza.
	let primero = true;
	$effect(() => {
		const n = paso;
		if (!listo) return;
		const u = new URL(location.href);
		if (u.searchParams.get('paso') === String(n)) return;
		u.searchParams.set('paso', String(n));
		const a = `${resolve('/onboarding')}${u.search}`;
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- la ruta pasa por resolve(); sólo cambia la consulta
		if (primero) replaceState(a, { paso: n });
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- la ruta pasa por resolve(); sólo cambia la consulta
		else pushState(a, { paso: n });
		primero = false;
	});
	$effect(() => {
		const s = page.state as { paso?: number };
		if (listo && typeof s.paso === 'number' && s.paso !== o.e.paso) o.ir(s.paso);
	});

	// Con la API, cada cambio de paso se guarda en la tienda (y lo que se elige, un momento después).
	let formPaso = $state<HTMLFormElement>();
	let formFin = $state<HTMLFormElement>();
	let sinGuardar = $state(false);
	$effect(() => {
		void paso;
		if (listo && !data.demo && data.conSesion && paso > 0)
			tick().then(() => formPaso?.requestSubmit());
	});
	const estado = $derived(
		JSON.stringify({
			plantilla: o.e.plantilla,
			elecciones: {
				// Con la API, el paso en que quedó: se retoma ahí (también desde otro dispositivo).
				...(data.demo ? {} : { paso: id }),
				objetivos: o.e.objetivos,
				rubro: o.e.rubro,
				conexiones: o.e.conexiones,
				conflicto: o.e.conflicto,
				instagram: o.e.instagram,
				secciones: o.e.secciones,
				config: o.e.config,
				tono: o.e.tono,
				equipo: o.e.equipo,
				importar: o.e.importar
			}
		})
	);
	let destino = $state<'panel' | 'producto'>('panel');
	let terminando = $state(false);
	const terminar = (a: 'panel' | 'producto' = 'panel') => {
		try {
			localStorage.setItem('ob-step', '0');
		} catch {
			// Sin almacenamiento: no importa.
		}
		destino = a;
		tick().then(() => formFin?.requestSubmit());
	};

	/**
	 * Paso 2 con la API: el nombre y la dirección se guardan al seguir (decisiones §38). Si la
	 * dirección no se puede, queda en el paso con el motivo. «Hacerlo después» no frena por eso.
	 */
	async function guardarTienda(): Promise<boolean> {
		if (!o.tienda) return true;
		const nombre = o.nombre.trim();
		const slug = slugFinal(o.slug) || o.tienda.slug;
		if (!nombre) {
			o.errorTienda = 'Poné el nombre de tu tienda.';
			return false;
		}
		if (nombre === o.tienda.nombre && slug === o.tienda.slug) return true;
		if (o.estadoDireccion === 'mal') {
			o.errorTienda = 'Elegí otra dirección: esa no se puede usar.';
			return false;
		}
		o.guardandoTienda = true;
		try {
			const f = new FormData();
			f.set('nombre', nombre);
			f.set('slug', slug);
			const r = await fetch('?/tienda', {
				method: 'POST',
				body: f,
				headers: { 'x-sveltekit-action': 'true' }
			});
			const res = deserialize(await r.text());
			if (res.type === 'failure') {
				o.errorTienda = String(res.data?.error ?? 'No pudimos guardar. Probá de nuevo.');
				return false;
			}
			if (res.type !== 'success') {
				o.errorTienda = 'No pudimos guardar. Probá de nuevo.';
				return false;
			}
			o.fijarTienda({ nombre, slug });
			await invalidateAll();
			return true;
		} catch {
			o.errorTienda = 'No pudimos guardar. Revisá tu conexión y probá de nuevo.';
			return false;
		} finally {
			o.guardandoTienda = false;
		}
	}

	const siguiente = async () => {
		if (id === 'negocio' && !data.demo && !(await guardarTienda())) return;
		if (id === 'importacion' && !o.importacionValida) {
			o.errorImportacion = 'Pegá el link de tu tienda, o elegí «Empiezo de cero».';
			return;
		}
		o.ir(paso + 1);
	};
	const despues = async () => {
		if (id === 'negocio' && !data.demo) await guardarTienda();
		// «Después» en la importación es no traer nada ahora (se hace desde Productos).
		if (id === 'importacion' && !data.demo) o.e.importar = null;
		o.ir(paso + 1);
	};
	const etiquetaSiguiente = $derived(
		id === 'importacion' && !o.importacionLista
			? 'Seguir mientras importa'
			: paso === total - 2
				? 'Terminar'
				: 'Continuar'
	);
	const ilustracion = $derived(`/hornero/${ILUSTRACION[id]}.webp`);
	const nota = $derived((data.demo ? NOTA_PIE[paso] : NOTA_PIE_REAL[id]) ?? 'Todo se guarda solo.');

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
	title="Creá tu tienda en Tu Local Web"
	description="Importamos tus productos y te dejamos la tienda lista en 10 minutos."
	path="/onboarding"
/>

<!-- eslint-disable svelte/no-navigation-without-resolve -- inicio y ayuda son rutas públicas fijas -->
<div class="ob">
	<Grano />
	<aside>
		<div class="luz"></div>
		<a class="marca" href="/" aria-label="Tu Local Web, inicio">
			<Marca color="#E8875F" alto={24} />
			<img src="/wordmark-cream.webp" alt="Tu Local Web" width="157" height="18" />
		</a>
		<div class="avance">
			<div class="a-t">Tu cuenta está</div>
			<div class="a-v"><span>{porcentaje}%</span><span class="a-l">lista</span></div>
			<div class="a-barra"><div style:width="{porcentaje}%"></div></div>
			<div class="a-d">
				{data.demo
					? 'Sin cargar nada a mano: lo traemos de lo que ya usás.'
					: 'Lo justo para empezar a vender. Lo demás, cuando quieras.'}
			</div>
		</div>
		<nav class="pasos" aria-label="Pasos">
			{#each o.pasos as p, i (p)}
				<button
					type="button"
					class:cur={i === paso}
					class:prox={i > paso}
					aria-current={i === paso ? 'step' : undefined}
					disabled={i > paso + 1 || (!data.demo && (!data.conSesion ? i > 0 : i === 0))}
					onclick={() => o.ir(i)}
				>
					{#if i < paso}
						<span class="n hecho"><Icon icon={I.phCheckBold} size={12} /></span>
					{:else}
						<span class="n">{i + 1}</span>
					{/if}
					<span class="p-l">{ROTULO[p]}</span>
				</button>
			{/each}
		</nav>
		<div class="esp"></div>
		<div class="consejo">
			<div class="c-av"><img src={ilustracion} alt="Tincho" width="52" height="52" /></div>
			<div>
				<div class="c-n">Tincho</div>
				<div class="c-t">{consejo}</div>
			</div>
		</div>
	</aside>

	<main>
		<div class="movil">
			<div class="m-fila">
				<Marca color="#E8875F" alto={22} />
				<span class="m-paso"
					><span class="m-n">Paso {paso + 1} de {total}</span><span class="m-l">{ROTULO[id]}</span
					></span
				>
				<span class="m-pct">{porcentaje}%</span>
				<button
					type="button"
					class="m-av"
					aria-label="Consejo de Tincho"
					aria-expanded={o.consejo}
					onclick={() => (o.consejo = !o.consejo)}><img src={ilustracion} alt="" /></button
				>
			</div>
			<!-- El avance, sin tocarse: en el celular se vuelve con «Atrás» (unas barras de 5 px no son
			     un botón que se pueda tocar). -->
			<div class="m-barras" aria-hidden="true">
				{#each o.pasos as p, i (p)}
					<span class:hecho={i < paso} class:cur={i === paso}></span>
				{/each}
			</div>
			{#if o.consejo}<div class="m-consejo">
					<b>Tincho:</b>
					{consejo}
				</div>{/if}
		</div>
		<div class="arriba">
			<span class="ar-paso">Paso {paso + 1} de {total}</span>
			{#if sinGuardar}
				<span class="ar-ok mal"
					><Icon icon={I.phWarningCircleFill} size={16} />No se pudo guardar: revisá tu conexión</span
				>
			{:else}
				<span class="ar-ok"
					><Icon icon={I.phCloudCheckFill} size={16} />Guardado automáticamente</span
				>
			{/if}
			<span class="esp"></span>
			<a href="/ayuda">¿Necesitás ayuda?</a>
			<a href="/">Salir</a>
		</div>

		<div class="contenido">
			<div class="paso" bind:this={caja}>
				{#if id === 'cuenta'}
					<PasoCuenta demo={data.demo} clerk={data.clerk} seguir={siguiente} />
				{:else if id === 'negocio'}
					<PasoNegocio />
				{:else if id === 'canales'}
					<PasoCanales />
				{:else if id === 'importacion'}
					<PasoImportacion />
				{:else if id === 'tienda'}
					<PasoTienda />
				{:else if id === 'cobros'}
					<PasoCobros cobros={data.cobros} mp={data.mp} volver={paso} />
				{:else if id === 'tincho'}
					<PasoHornero />
				{:else}
					<PasoListo {terminar} {terminando} />
					{#if form && 'error' in form}<div class="error" role="alert">{form.error}</div>{/if}
				{/if}
			</div>
		</div>

		{#if paso > 0 && id !== 'listo'}
			<div class="pie">
				{#if paso > 1}
					<button type="button" class="atras" onclick={() => o.ir(paso - 1)}
						><Icon icon={I.phArrowLeftBold} size={14} /><span class="atras-t">Atrás</span></button
					>
				{/if}
				<div class="nota">{nota}</div>
				<button type="button" class="saltar" onclick={despues} disabled={o.guardandoTienda}
					><span class="largo">Hacerlo después</span><span class="corto">Después</span></button
				>
				<button type="button" class="seguir" onclick={siguiente} disabled={o.guardandoTienda}
					>{o.guardandoTienda ? 'Guardando…' : etiquetaSiguiente}<Icon
						icon={I.phArrowRightBold}
						size={15}
					/></button
				>
			</div>
		{/if}
	</main>
</div>

<form
	method="POST"
	action="?/paso"
	bind:this={formPaso}
	use:enhance={() =>
		async ({ result }) => {
			sinGuardar = result.type !== 'success';
		}}
	hidden
>
	<input type="hidden" name="estado" value={estado} />
	<input type="hidden" name="listo" value={id === 'listo' ? '1' : ''} />
</form>
<form
	method="POST"
	action="?/terminar"
	bind:this={formFin}
	use:enhance={() => {
		terminando = true;
		return async ({ update }) => {
			await update();
			terminando = false;
		};
	}}
	hidden
>
	<input type="hidden" name="estado" value={estado} />
	<input type="hidden" name="destino" value={destino} />
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
		/* El confeti de «¡Listo!» no abre scroll horizontal (clip no corta el pie pegajoso). */
		overflow-x: clip;
	}
	/* Un campo enfocado no queda tapado por el pie (ni por el teclado que lo empuja). */
	.contenido :global(input),
	.contenido :global(textarea) {
		scroll-margin-bottom: 110px;
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
	.ar-ok.mal {
		color: #b83a2e;
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
		min-height: 44px;
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
		min-height: 44px;
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
			width: 44px;
			height: 44px;
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
		.m-barras span {
			flex: 1;
			height: 5px;
			border-radius: 3px;
			background: rgba(251, 247, 242, 0.18);
		}
		.m-barras span.hecho {
			background: #2f8f63;
		}
		.m-barras span.cur {
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
