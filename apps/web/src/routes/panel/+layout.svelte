<script lang="ts">
	import { untrack } from 'svelte';
	import { page } from '$app/state';
	import { invalidateAll } from '$app/navigation';
	import BarraAbajo from '#lib/panel/BarraAbajo.svelte';
	import BarraLateral from '#lib/panel/BarraLateral.svelte';
	import Cmdk from '#lib/panel/Cmdk.svelte';
	import Notificaciones from '#lib/panel/Notificaciones.svelte';
	import Rail from '#lib/panel/Rail.svelte';
	import Topbar from '#lib/panel/Topbar.svelte';
	import { AREAS, ubicar } from '#lib/panel/navegacion.ts';
	import { Shell, setShell } from '#lib/panel/shell.svelte.ts';
	import { SITE } from '#lib/site.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/** Armazón del panel (design/Mostrador v4.dc.html, vista Panel). */
	let { data, children } = $props();
	// Tema y modo se toman al entrar; después los maneja el armazón (y los guarda en cookies).
	const shell = setShell(
		new Shell(
			untrack(() => data.tema),
			untrack(() => data.modo)
		)
	);
	const ui = getUi();
	shell.hornero = () =>
		ui.toast({ k: 'ai', title: 'Hornero', desc: 'Llega con la próxima versión del panel.' });
	const donde = $derived(ubicar(page.url.pathname) ?? { seccion: 'Resumen', area: AREAS[0] });
	const sinLeer = $derived(data.resumen.notificaciones.some((n) => !n.leida));
	const dominio = $derived(`${SITE.replace(/^https?:\/\//, '')}/t/${data.tienda.slug}`);
	const iniciales = $derived(
		data.tienda.nombre
			.split(/\s+/)
			.slice(0, 2)
			.map((p) => p[0])
			.join('')
			.toUpperCase()
	);

	async function marcarLeidas() {
		await fetch('/panel/notificaciones', { method: 'POST' });
		await invalidateAll();
	}
	// Al cambiar de sección se cierra lo que haya abierto.
	$effect(() => {
		void page.url.pathname;
		shell.cerrarTodo();
	});
</script>

<!-- Escape cierra la barra lateral y las notificaciones (⌘K lo maneja su diálogo). -->
<svelte:window
	onkeydown={(e) => {
		if (e.key === 'Escape' && !shell.cmdk) {
			shell.barra = false;
			shell.notificaciones = false;
		}
	}}
/>

<svelte:head>
	<meta name="robots" content="noindex" />
	<meta name="theme-color" content={shell.tema === 'dark' ? '#22140d' : '#f7f1ea'} />
</svelte:head>

<div class="panel" data-theme={shell.tema}>
	<Rail area={donde.area} foto={data.demo ? '1545262722-9e0d80a0bc01' : null} {iniciales} />
	<BarraLateral
		area={donde.area}
		seccion={donde.seccion}
		tienda={data.tienda.nombre}
		foto={data.fotoTienda}
		{dominio}
		resumen={data.resumen}
	/>
	<main>
		<div class="luz" aria-hidden="true"></div>
		<Topbar
			area={shell.modo === 'simple' ? 'Día a día' : donde.area.label}
			seccion={donde.seccion === 'Resumen' && shell.modo === 'simple' ? 'Inicio' : donde.seccion}
			foto={data.fotoTienda}
			{sinLeer}
		/>
		<div class="contenido">
			{@render children()}
		</div>
	</main>
	<BarraAbajo seccion={donde.seccion} pedidos={data.resumen.insignias.pedidos} />
	<Notificaciones notificaciones={data.resumen.notificaciones} {marcarLeidas} />
	<Cmdk />
</div>

<style>
	.panel {
		display: flex;
		min-height: 100vh;
		min-height: 100dvh;
		background: var(--bg);
		color: var(--tx);
		transition: background 0.3s;
	}
	main {
		flex: 1;
		min-width: 0;
		position: relative;
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
	.contenido {
		position: relative;
		padding: 20px 36px 120px;
		max-width: 1440px;
	}
	@media (max-width: 1099px) {
		.contenido {
			padding: 20px 24px 120px;
		}
	}
	@media (max-width: 759px) {
		.contenido {
			padding: 14px 12px 150px;
		}
	}
</style>
