<script lang="ts">
	import { untrack } from 'svelte';
	import { page } from '$app/state';
	import { invalidateAll } from '$app/navigation';
	import BarraAbajo from '#lib/panel/BarraAbajo.svelte';
	import BarraLateral from '#lib/panel/BarraLateral.svelte';
	import Cmdk from '#lib/panel/Cmdk.svelte';
	import Hornero from '#lib/panel/Hornero.svelte';
	import LinkPago from '#lib/panel/LinkPago.svelte';
	import Notificaciones from '#lib/panel/Notificaciones.svelte';
	import Rail from '#lib/panel/Rail.svelte';
	import Topbar from '#lib/panel/Topbar.svelte';
	import { AREAS, ubicar } from '#lib/panel/navegacion.ts';
	import { Shell, setShell } from '#lib/panel/shell.svelte.ts';
	import { SITE } from '#lib/site.ts';

	/** Armazón del panel (design/Mostrador v4.dc.html, vista Panel). */
	let { data, children } = $props();
	// El tema se toma al entrar; después lo maneja el armazón (y lo guarda en una cookie).
	const shell = setShell(new Shell(untrack(() => data.tema)));
	// La sección Turnos (también en el primer render del servidor, donde los efectos no corren).
	shell.turnos = untrack(() => data.turnos);
	// Lo que el rol puede ver y las tiendas de la persona (también en el primer render del servidor).
	const acceso = () => {
		shell.permisos = data.acceso.permisos;
		shell.tiendas = data.acceso.tiendas;
		shell.tiendaId = data.acceso.tiendaId;
	};
	untrack(acceso);
	$effect.pre(() => {
		shell.turnos = data.turnos;
		acceso();
	});
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
	// Al cambiar de sección se cierra lo que haya abierto. Depende del texto de la ruta: al recargar
	// los datos (invalidateAll) `page.url` es otro objeto con la misma ruta y no tiene que cerrar nada.
	const ruta = $derived(page.url.pathname);
	$effect(() => {
		void ruta;
		untrack(() => shell.cerrarTodo());
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

{#if data.mirada}
	<form class="mirada" method="POST" action="/staff/mirada/salir">
		<span
			>Estás viendo {data.tienda.nombre} como la ve la dueña, en modo lectura: no se puede cambiar nada.</span
		>
		<button type="submit">Volver al equipo</button>
	</form>
{/if}
<div class="panel" data-theme={shell.tema}>
	<Rail area={donde.area} {iniciales} />
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
	<Hornero />
	<LinkPago />
</div>

<style>
	.mirada {
		position: sticky;
		top: 0;
		z-index: 60;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-wrap: wrap;
		gap: 8px 14px;
		padding: 9px 16px;
		background: #2a140c;
		color: #fbf7f2;
		font-size: 13.5px;
		text-align: center;
	}
	.mirada button {
		padding: 6px 14px;
		border-radius: 999px;
		border: 1px solid rgba(251, 247, 242, 0.3);
		background: transparent;
		color: inherit;
		font: inherit;
		font-weight: 600;
	}
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
