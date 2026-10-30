<script lang="ts">
	import { onMount } from 'svelte';
	import {
		activarAvisos,
		app,
		iniciar,
		instalar,
		marcarVista,
		yaVista
	} from '#lib/avisos.svelte.ts';

	/**
	 * Al entrar: instalar el correo y activar los avisos. Aparece hasta que se completa o se elige
	 * «Ahora no»; con la app ya instalada y los avisos sin decidir, vuelve a aparecer (en iPhone la
	 * app instalada es otro almacenamiento, y es recién ahí donde se pueden activar).
	 */
	let abierta = $state(false);
	onMount(() => {
		iniciar();
		const pendiente = !app.instalada || app.permiso === 'default';
		abierta = pendiente && (!yaVista() || (app.instalada && app.permiso === 'default'));
	});
	const avisosListos = $derived(app.permiso === 'granted');
	const puedeAvisos = $derived(app.permiso !== 'no-soportado' && !(app.ios && !app.instalada));
	function cerrar() {
		marcarVista();
		abierta = false;
	}
	async function avisos() {
		if (await activarAvisos()) setTimeout(cerrar, 900);
	}
</script>

{#if abierta}
	<div class="velo">
		<div class="panel vidrio" role="dialog" aria-modal="true" aria-labelledby="bv-t">
			<img src="/icono-192.png" alt="" width="64" height="64" />
			<h2 id="bv-t">Tené el correo a mano</h2>
			<p class="sub">Instalalo como app y te avisamos cuando llegue algo.</p>

			<div class="paso" class:hecho={app.instalada}>
				<span class="n"
					>{#if app.instalada}<i class="icon-check"></i>{:else}1{/if}</span
				>
				<div class="tx">
					<b>Instalá el correo</b>
					{#if app.instalada}
						<span>Ya está instalado.</span>
					{:else if app.instalable}
						<span>Queda en tu pantalla de inicio, como cualquier app.</span>
						<button class="acc" onclick={instalar}><i class="icon-download"></i>Instalar</button>
					{:else if app.ios}
						<span
							>En Safari, tocá <i class="icon-share" aria-label="Compartir"></i> y después
							<b>Agregar a inicio</b> <i class="icon-square-plus"></i>. Abrilo desde ahí.</span
						>
					{:else}
						<span
							>Desde el menú del navegador, elegí <b>Instalar app</b> o
							<b>Agregar a la pantalla de inicio</b>.</span
						>
					{/if}
				</div>
			</div>

			<div class="paso" class:hecho={avisosListos}>
				<span class="n"
					>{#if avisosListos}<i class="icon-check"></i>{:else}2{/if}</span
				>
				<div class="tx">
					<b>Activá los avisos</b>
					{#if avisosListos}
						<span>Listo: te avisamos cuando llegue un correo.</span>
					{:else if app.permiso === 'denied'}
						<span>Los bloqueaste. Habilitalos en los ajustes del navegador para este sitio.</span>
					{:else if !puedeAvisos}
						<span
							>{app.ios
								? 'En iPhone, los avisos andan con la app instalada: instalala y abrila desde la pantalla de inicio.'
								: 'Este navegador no puede recibir avisos.'}</span
						>
					{:else}
						<span>Remitente y asunto, apenas entra a cualquiera de tus casillas.</span>
						<button class="acc" onclick={avisos} disabled={app.activando}
							><i class="icon-bell"></i>{app.activando ? 'Activando…' : 'Activar avisos'}</button
						>
					{/if}
					{#if app.error}<span class="error">{app.error}</span>{/if}
				</div>
			</div>

			<button class="luego" onclick={cerrar}
				>{app.instalada && avisosListos ? 'Listo' : 'Ahora no'}</button
			>
		</div>
	</div>
{/if}

<style>
	.velo {
		position: fixed;
		inset: 0;
		z-index: 50;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 16px;
		background: rgb(0 0 0 / 0.6);
		backdrop-filter: blur(8px);
		-webkit-backdrop-filter: blur(8px);
	}
	.panel {
		width: min(100%, 460px);
		max-height: calc(100dvh - 32px);
		overflow-y: auto;
		display: flex;
		flex-direction: column;
		gap: 14px;
		padding: 28px;
		border-radius: 16px;
		background: color-mix(in oklch, var(--pop) 55%, var(--side));
		box-shadow:
			var(--popsh),
			0 0 0 1px var(--ls);
	}
	img {
		border-radius: 14px;
	}
	h2 {
		margin: 4px 0 0;
		font-size: 24px;
		font-weight: 600;
		line-height: 1.3;
		letter-spacing: -0.01em;
	}
	.sub {
		margin: -8px 0 4px;
		color: var(--mut);
	}
	.paso {
		display: flex;
		gap: 14px;
		padding: 14px;
		border-radius: 12px;
		background: var(--hun);
	}
	.n {
		width: 32px;
		height: 32px;
		flex: none;
		border-radius: 9px;
		display: flex;
		align-items: center;
		justify-content: center;
		background: var(--accBg);
		color: var(--accT);
		font-weight: 700;
	}
	.hecho .n {
		background: var(--okBg);
		color: var(--okT);
	}
	.tx {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 4px;
		font-size: 15px;
		color: var(--mut);
	}
	.tx b {
		color: var(--fg);
		font-weight: 600;
	}
	.tx > b {
		font-size: 16px;
	}
	.tx i {
		vertical-align: -2px;
		color: var(--accT);
	}
	.acc {
		align-self: flex-start;
		margin-top: 8px;
		display: inline-flex;
		align-items: center;
		gap: 10px;
		height: 46px;
		padding: 0 20px;
		border-radius: 999px;
		background: var(--acc);
		color: #fff;
		font-size: 16px;
		font-weight: 600;
	}
	.acc i {
		color: #fff;
		font-size: 18px;
	}
	.acc:disabled {
		opacity: 0.7;
	}
	.error {
		color: var(--prT);
	}
	.luego {
		align-self: center;
		margin-top: 4px;
		padding: 8px 14px;
		color: var(--mut);
		font-size: 15px;
		font-weight: 500;
	}
</style>
