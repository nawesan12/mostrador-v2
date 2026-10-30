<script lang="ts">
	import { enhance } from '$app/forms';
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import { activarAvisos, app, iniciar } from '#lib/avisos.svelte.ts';
	import type { Regla } from '#lib/server/ajustes.ts';
	import Marca from '#lib/ui/Marca.svelte';
	import { COLORES } from '#lib/ui/Carpetas.svelte';

	let { data, form } = $props();

	const NAV = [
		{ id: 'firma', l: 'Firma', ic: 'icon-pen-line' },
		{ id: 'respuesta', l: 'Respuesta automática', ic: 'icon-plane' },
		{ id: 'reglas', l: 'Reglas', ic: 'icon-filter' },
		{ id: 'reenviar', l: 'Reenviar correos', ic: 'icon-forward' },
		{ id: 'avisos', l: 'Avisos en el teléfono', ic: 'icon-bell' },
		{ id: 'clave', l: 'Clave y sesiones', ic: 'icon-lock' }
	];
	const s = $derived(
		NAV.some((n) => n.id === page.url.searchParams.get('s'))
			? page.url.searchParams.get('s')!
			: 'firma'
	);
	const color = (cuenta: string) =>
		COLORES[
			Math.max(
				0,
				data.casillas.findIndex((c) => c.cuenta === cuenta)
			) % COLORES.length
		];
	const corto = (email: string) => email.replace('@grupovimo.com.ar', '@');

	// Firma
	const EJEMPLO_FIRMA = 'Nahuel Santillán\nGrupo Vimo\ngrupovimo.com.ar';
	let cuentaFirma = $state('');
	const firma = $derived(data.firmas.find((f) => f.cuenta === cuentaFirma) ?? data.firmas[0]);
	let nombre = $derived(firma?.nombre ?? '');
	let textoFirma = $derived(firma?.texto ?? '');

	// Respuesta automática
	let cuentaVac = $state('');
	const vac = $derived(data.vacaciones.find((v) => v.cuenta === cuentaVac) ?? data.vacaciones[0]);
	let activa = $derived(vac?.activa ?? false);
	const dia = (iso: string | null) =>
		iso
			? new Date(iso).toLocaleDateString('en-CA', { timeZone: 'America/Argentina/Buenos_Aires' })
			: '';

	// Reglas y reenvío (se guardan juntos: son el mismo script)
	// Estado (no derivado): se agregan, editan y borran reglas antes de guardar.
	const reglasGuardadas = () => structuredClone(data.filtros.reglas);
	let reglas = $state<Regla[]>(reglasGuardadas());
	let reenviarA = $derived(data.filtros.reenviarA);
	let guardarCopia = $derived(data.filtros.guardarCopia);
	let editando = $state<number | null>(null);
	const filtrosJson = $derived(JSON.stringify({ reglas, reenviarA, guardarCopia }));
	const SI = { de: 'Si llega de', asunto: 'Si el asunto dice' };
	const ACCION = { mover: 'moverlo a', etiquetar: 'ponerle la etiqueta', reenviar: 'mandarlo a' };

	let guardando = $state(false);
	let aviso = $state('');
	$effect(() => {
		if (form?.ok) aviso = form.ok;
		else if (form?.error) aviso = form.error;
		const t = setTimeout(() => (aviso = ''), 3500);
		return () => clearTimeout(t);
	});
	const mejorar = () => {
		guardando = true;
		return async ({ update }: { update: (o?: { reset?: boolean }) => Promise<void> }) => {
			await update({ reset: false });
			guardando = false;
			editando = null;
		};
	};

	onMount(iniciar);
	async function avisos() {
		aviso = (await activarAvisos())
			? 'Avisos activados en este dispositivo'
			: app.error || 'No se activaron.';
	}
</script>

<svelte:head><title>Ajustes · Correo</title></svelte:head>

<div class="app">
	<header class="barra">
		<div class="marca">
			<Marca marca={data.marca} />
			<div class="mt">
				<div class="mc">Correo</div>
				<div class="mm">{data.marca}</div>
			</div>
		</div>
		<a class="volver" href="/"><i class="icon-arrow-left"></i>Volver al correo</a>
		<span class="av">{data.yo.charAt(0).toUpperCase()}</span>
	</header>

	<div class="cuerpo">
		<nav class="nav" aria-label="Ajustes">
			<div class="tit">Ajustes</div>
			{#each NAV as n (n.id)}
				<a
					class="item"
					class:act={s === n.id}
					href="?s={n.id}"
					aria-current={s === n.id ? 'page' : undefined}><i class={n.ic}></i>{n.l}</a
				>
			{/each}
		</nav>

		<main class="tarjeta vidrio">
			{#if s === 'firma' && firma}
				<section>
					<div class="cab">
						<div>
							<h2>Firma</h2>
							<p>Se agrega sola al final de cada correo. Cada casilla tiene la suya.</p>
						</div>
					</div>
					<div class="chips">
						{#each data.firmas as f (f.cuenta)}
							<button
								class="chip"
								class:act={f.cuenta === firma.cuenta}
								onclick={() => (cuentaFirma = f.cuenta)}>{corto(f.email)}</button
							>
						{/each}
					</div>
					<form method="POST" action="?/firma" use:enhance={mejorar} class="dos">
						<input type="hidden" name="cuenta" value={firma.cuenta} />
						<input type="hidden" name="identidad" value={firma.identidad} />
						<div class="campos">
							<label
								><span>Nombre con el que sale</span><input
									class="in"
									name="nombre"
									bind:value={nombre}
									placeholder="Nahuel · Grupo Vimo"
								/></label
							>
							<label
								><span>Firma</span><textarea
									class="in"
									name="texto"
									rows="6"
									bind:value={textoFirma}
									placeholder={EJEMPLO_FIRMA}></textarea></label
							>
							<button class="guardar" disabled={guardando}
								>{guardando ? 'Guardando…' : 'Guardar firma'}</button
							>
						</div>
						<div class="vista">
							<div class="vt">Así la recibe quien lee</div>
							<div class="firma">
								<Marca marca={data.marca} tam={64} />
								<div class="ft">
									{#each (textoFirma || nombre || firma.email).split('\n') as l, i (i)}
										<div class:primera={i === 0}>{l}</div>
									{/each}
								</div>
							</div>
						</div>
					</form>
				</section>
			{:else if s === 'respuesta' && vac}
				<form method="POST" action="?/vacaciones" use:enhance={mejorar}>
					<section>
						<div class="cab">
							<div>
								<h2>Respuesta automática</h2>
								<p>Para cuando no estás: vacaciones, feriados, cierre por inventario.</p>
							</div>
							<label class="tg" aria-label="Activar la respuesta automática"
								><input type="checkbox" name="activa" bind:checked={activa} /><span></span></label
							>
						</div>
						<input type="hidden" name="cuenta" value={vac.cuenta} />
						<div class="tres">
							<label
								><span>Desde</span><input
									class="in"
									type="date"
									name="desde"
									value={dia(vac.desde)}
								/></label
							>
							<label
								><span>Hasta</span><input
									class="in"
									type="date"
									name="hasta"
									value={dia(vac.hasta)}
								/></label
							>
							<label
								><span>Casilla</span><span class="sel"
									><span class="punto" style:background={color(vac.cuenta)}></span><select
										bind:value={cuentaVac}
										aria-label="Casilla"
										>{#each data.vacaciones as v (v.cuenta)}<option value={v.cuenta}
												>{v.email}</option
											>{/each}</select
									><i class="icon-chevron-down"></i></span
								></label
							>
						</div>
						<label
							><span>Asunto (opcional)</span><input
								class="in"
								name="asunto"
								value={vac.asunto}
							/></label
						>
						<label
							><span>Mensaje</span><textarea
								class="in"
								name="texto"
								rows="4"
								value={vac.texto}
								placeholder="Hola, estoy de vacaciones hasta el 6 de enero. Para pedidos escribí a hola@…"
							></textarea></label
						>
						<button class="guardar" disabled={guardando}
							>{guardando ? 'Guardando…' : 'Guardar'}</button
						>
					</section>
				</form>
			{:else if s === 'reglas'}
				<form method="POST" action="?/filtros" use:enhance={mejorar}>
					<input type="hidden" name="filtros" value={filtrosJson} />
					<section>
						<div class="cab">
							<div>
								<h2>Reglas</h2>
								<p>Ordenan solos los correos que llegan a {corto(data.yo)}.</p>
							</div>
							<button
								type="button"
								class="nueva"
								onclick={() => {
									reglas.push({ si: 'de', valor: '', accion: 'mover', destino: '' });
									editando = reglas.length - 1;
								}}><i class="icon-plus"></i>Nueva regla</button
							>
						</div>
						{#each reglas as r, i (i)}
							<div class="regla">
								{#if editando === i}
									<select class="in ch" bind:value={r.si} aria-label="Condición"
										><option value="de">Si llega de</option><option value="asunto"
											>Si el asunto dice</option
										></select
									>
									<input
										class="in ch"
										bind:value={r.valor}
										placeholder={r.si === 'de' ? 'banco.com o proveedor@x.com' : 'factura'}
										aria-label="Valor"
									/>
									<select class="in ch" bind:value={r.accion} aria-label="Acción"
										><option value="mover">moverlo a</option><option value="etiquetar"
											>ponerle la etiqueta</option
										><option value="reenviar">mandarlo a</option></select
									>
									<input
										class="in ch"
										bind:value={r.destino}
										placeholder={r.accion === 'reenviar'
											? 'email'
											: r.accion === 'mover'
												? 'Carpeta'
												: 'Etiqueta'}
										aria-label="Destino"
									/>
									<button
										type="button"
										class="ic"
										aria-label="Listo"
										onclick={() => (editando = null)}><i class="icon-check"></i></button
									>
								{:else}
									<span class="m">{SI[r.si]}</span><span class="v">{r.valor || '…'}</span><span
										class="m">{ACCION[r.accion]}</span
									><span class="v">{r.destino || '…'}</span>
									<button
										type="button"
										class="ic"
										aria-label="Editar"
										onclick={() => (editando = i)}><i class="icon-pencil"></i></button
									>
								{/if}
								<button
									type="button"
									class="ic"
									aria-label="Borrar"
									onclick={() => reglas.splice(i, 1)}><i class="icon-trash-2"></i></button
								>
							</div>
						{:else}
							<div class="vacio">Todavía no hay reglas.</div>
						{/each}
						<button class="guardar" disabled={guardando}
							>{guardando ? 'Guardando…' : 'Guardar reglas'}</button
						>
					</section>
				</form>
			{:else if s === 'reenviar'}
				<form method="POST" action="?/filtros" use:enhance={mejorar}>
					<input type="hidden" name="filtros" value={filtrosJson} />
					<section>
						<div class="cab">
							<div>
								<h2>Reenviar correos</h2>
								<p>Todo lo que llega a {corto(data.yo)} se manda también a otra dirección.</p>
							</div>
						</div>
						<label
							><span>Reenviar a</span><input
								class="in"
								type="email"
								bind:value={reenviarA}
								placeholder="vos@gmail.com"
							/></label
						>
						<label class="check"
							><input type="checkbox" bind:checked={guardarCopia} />Dejar también una copia acá</label
						>
						<button class="guardar" disabled={guardando}
							>{guardando ? 'Guardando…' : 'Guardar'}</button
						>
					</section>
				</form>
			{:else if s === 'avisos'}
				<section>
					<div class="cab">
						<div>
							<h2>Avisos en el teléfono</h2>
							<p>Remitente y asunto, apenas entra un correo a cualquiera de tus casillas.</p>
						</div>
					</div>
					<div class="estado">
						{#if app.permiso === 'granted'}
							<i class="icon-bell-ring"></i><span>Este dispositivo recibe avisos.</span>
						{:else if app.permiso === 'denied'}
							<i class="icon-bell-off"></i><span
								>Los bloqueaste: habilitalos en los ajustes del navegador.</span
							>
						{:else if app.permiso === 'no-soportado' || (app.ios && !app.instalada)}
							<i class="icon-smartphone"></i><span
								>{app.ios
									? 'En iPhone, instalá el correo (Compartir → Agregar a inicio) y abrilo desde ahí.'
									: 'Este navegador no puede recibir avisos.'}</span
							>
						{:else}
							<i class="icon-bell"></i><span>Este dispositivo todavía no recibe avisos.</span>
							<button class="guardar" onclick={avisos} disabled={app.activando}
								>{app.activando ? 'Activando…' : 'Activar avisos'}</button
							>
						{/if}
					</div>
				</section>
			{:else if s === 'clave'}
				<form method="POST" action="?/clave" use:enhance={mejorar}>
					<section>
						<div class="cab">
							<div>
								<h2>Clave</h2>
								<p>
									La de {data.yo}. Cambia también para el correo en el teléfono y otros programas.
								</p>
							</div>
						</div>
						<label
							><span>Clave actual</span><input
								class="in"
								type="password"
								name="actual"
								autocomplete="current-password"
								required
							/></label
						>
						<div class="dos-c">
							<label
								><span>Clave nueva</span><input
									class="in"
									type="password"
									name="nueva"
									autocomplete="new-password"
									minlength="10"
									required
								/></label
							>
							<label
								><span>Repetila</span><input
									class="in"
									type="password"
									name="repetir"
									autocomplete="new-password"
									minlength="10"
									required
								/></label
							>
						</div>
						<button class="guardar" disabled={guardando}
							>{guardando ? 'Cambiando…' : 'Cambiar la clave'}</button
						>
					</section>
				</form>
			{/if}
		</main>
	</div>
	{#if aviso}<div class="toast" role="status">{aviso}</div>{/if}
</div>

<style>
	.app {
		height: 100dvh;
		display: flex;
		flex-direction: column;
		overflow: hidden;
		padding-top: env(safe-area-inset-top);
		padding-left: env(safe-area-inset-left);
		padding-right: env(safe-area-inset-right);
	}
	.barra {
		height: 72px;
		flex: none;
		display: flex;
		align-items: center;
		gap: 16px;
		padding: 0 20px;
	}
	.marca {
		width: 236px;
		flex: none;
		display: flex;
		align-items: center;
		gap: 12px;
		padding-left: 6px;
	}
	.mt {
		line-height: 1.2;
	}
	.mc {
		font-size: 19px;
		font-weight: 600;
	}
	.mm {
		font-size: 14px;
		color: var(--mut);
	}
	.volver {
		display: inline-flex;
		align-items: center;
		gap: 10px;
		height: 44px;
		padding: 0 16px 0 10px;
		border-radius: 10px;
		font-size: 17px;
		font-weight: 500;
		color: var(--fg);
		text-decoration: none;
	}
	.volver i {
		font-size: 20px;
	}
	.av {
		margin-left: auto;
		width: 40px;
		height: 40px;
		border-radius: 50%;
		background: oklch(0.68 calc(0.19 * var(--ac)) calc(var(--h) + 0));
		color: #fff;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 15px;
		font-weight: 600;
	}
	.cuerpo {
		flex: 1;
		min-height: 0;
		display: flex;
	}
	.nav {
		width: 256px;
		flex: none;
		padding: 4px 12px 16px 14px;
		display: flex;
		flex-direction: column;
		gap: 2px;
	}
	.tit {
		padding: 6px 14px 10px;
		font-size: 24px;
		font-weight: 600;
		letter-spacing: -0.01em;
	}
	.item {
		display: flex;
		align-items: center;
		gap: 14px;
		height: 46px;
		padding: 0 14px;
		border-radius: 999px;
		font-size: 16px;
		color: var(--fg);
		text-decoration: none;
		white-space: nowrap;
	}
	.item i {
		font-size: 20px;
	}
	.item.act {
		background: var(--side-act);
		color: var(--accT);
		font-weight: 600;
	}
	.tarjeta {
		flex: 1;
		min-width: 0;
		margin: 0 16px 16px 0;
		background-color: var(--card);
		background-image: var(--texC);
		border-radius: 14px;
		box-shadow: var(--sh);
		overflow-y: auto;
		padding: 36px 48px;
	}
	section {
		max-width: 880px;
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
	.cab {
		display: flex;
		align-items: center;
		gap: 16px;
	}
	.cab > div {
		flex: 1;
	}
	h2 {
		margin: 0;
		font-size: 22px;
		font-weight: 600;
	}
	.cab p {
		margin: 0;
		font-size: 16px;
		color: var(--mut);
	}
	label {
		display: flex;
		flex-direction: column;
		gap: 6px;
		font-size: 15px;
		color: var(--mut);
	}
	.in {
		min-height: 50px;
		padding: 0 14px;
		border: 1px solid var(--ln);
		border-radius: 12px;
		background: transparent;
		color: var(--fg);
		font-size: 17px;
		color-scheme: dark light;
	}
	textarea.in {
		padding: 14px 16px;
		line-height: 1.6;
		resize: vertical;
	}
	.in:focus {
		outline: 0;
		border-color: var(--acc);
		box-shadow: 0 0 0 1px var(--acc);
	}
	.chips {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}
	.chip {
		height: 42px;
		display: inline-flex;
		align-items: center;
		padding: 0 18px;
		border-radius: 999px;
		font-size: 16px;
		border: 1px solid var(--ln);
		font-weight: 500;
	}
	.chip.act {
		background: var(--acc);
		color: #fff;
		border-color: transparent;
		font-weight: 600;
	}
	.dos {
		display: grid;
		grid-template-columns: minmax(0, 1fr) minmax(0, 1.1fr);
		gap: 24px;
		align-items: start;
	}
	.campos {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}
	.vista {
		border: 1px solid var(--ln);
		border-radius: 14px;
		overflow: hidden;
	}
	.vt {
		padding: 10px 16px;
		border-bottom: 1px solid var(--ls);
		font-size: 14px;
		color: var(--mut);
	}
	.firma {
		padding: 20px 22px;
		display: flex;
		gap: 16px;
		align-items: flex-start;
	}
	.ft {
		font-size: 16px;
		line-height: 1.55;
		color: var(--mut);
		min-width: 0;
		overflow-wrap: break-word;
	}
	.ft .primera {
		font-size: 18px;
		font-weight: 600;
		color: var(--fg);
	}
	.tres {
		display: grid;
		grid-template-columns: 1fr 1fr 1.4fr;
		gap: 14px;
	}
	.dos-c {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 14px;
	}
	.sel {
		position: relative;
		height: 50px;
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 0 14px;
		border: 1px solid var(--ln);
		border-radius: 12px;
		color: var(--fg);
		font-size: 17px;
	}
	.sel select {
		flex: 1;
		min-width: 0;
		appearance: none;
		border: 0;
		background: transparent;
		font-size: 17px;
		cursor: pointer;
	}
	.sel option {
		background: var(--pop);
		color: var(--fg);
	}
	.sel i {
		font-size: 18px;
		color: var(--mut);
		pointer-events: none;
	}
	.punto {
		width: 10px;
		height: 10px;
		border-radius: 50%;
		flex: none;
	}
	.tg {
		position: relative;
		flex: none;
		width: 56px;
		height: 32px;
		cursor: pointer;
	}
	.tg input {
		position: absolute;
		opacity: 0;
		inset: 0;
		margin: 0;
		cursor: pointer;
	}
	.tg span {
		position: absolute;
		inset: 0;
		border-radius: 999px;
		background: var(--mu);
		box-shadow: inset 0 0 0 1px var(--ln);
		transition: background 0.15s;
	}
	.tg span::after {
		content: '';
		position: absolute;
		left: 3px;
		top: 3px;
		width: 26px;
		height: 26px;
		border-radius: 50%;
		background: #fff;
		transition: transform 0.15s;
	}
	.tg input:checked + span {
		background: var(--acc);
		box-shadow: none;
	}
	.tg input:checked + span::after {
		transform: translateX(24px);
	}
	.tg input:focus-visible + span {
		outline: 2px solid var(--acc);
		outline-offset: 2px;
	}
	.nueva {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		height: 44px;
		padding: 0 18px;
		border: 1px solid var(--ln);
		border-radius: 999px;
		font-size: 16px;
		font-weight: 500;
		white-space: nowrap;
	}
	.regla {
		display: flex;
		align-items: center;
		gap: 8px;
		flex-wrap: wrap;
		padding: 12px 16px;
		border-radius: 12px;
		background: var(--hun);
		font-size: 17px;
	}
	.regla .m {
		color: var(--mut);
	}
	.regla .v {
		padding: 2px 10px;
		border-radius: 8px;
		background: var(--card);
		box-shadow: inset 0 0 0 1px var(--ls);
		font-weight: 500;
	}
	.regla .ch {
		min-height: 40px;
		font-size: 16px;
		padding: 0 10px;
	}
	.regla .ic:first-of-type {
		margin-left: auto;
	}
	.ic {
		padding: 6px;
		color: var(--mut);
		font-size: 18px;
	}
	.vacio {
		padding: 16px;
		border-radius: 12px;
		background: var(--hun);
		color: var(--mut);
	}
	.check {
		flex-direction: row;
		align-items: center;
		gap: 10px;
		color: var(--fg);
		font-size: 16px;
	}
	.check input {
		width: 18px;
		height: 18px;
		accent-color: var(--acc);
	}
	.estado {
		display: flex;
		align-items: center;
		flex-wrap: wrap;
		gap: 12px;
		padding: 16px;
		border-radius: 12px;
		background: var(--hun);
		font-size: 16px;
	}
	.estado i {
		font-size: 22px;
		color: var(--accT);
	}
	.estado .guardar {
		margin-left: auto;
	}
	.guardar {
		align-self: flex-start;
		display: inline-flex;
		align-items: center;
		height: 48px;
		padding: 0 24px;
		border-radius: 999px;
		background: var(--acc);
		color: #fff;
		font-size: 16px;
		font-weight: 600;
	}
	.guardar:disabled {
		opacity: 0.7;
	}
	.toast {
		position: fixed;
		left: 50%;
		bottom: calc(28px + env(safe-area-inset-bottom));
		transform: translateX(-50%);
		z-index: 40;
		padding: 12px 18px;
		border-radius: 12px;
		background: var(--toast);
		color: var(--toastT);
		font-size: 15px;
		font-weight: 500;
		box-shadow: var(--popsh);
	}
	@media (max-width: 900px) {
		.barra {
			height: 64px;
			padding: 0 16px;
			gap: 8px;
		}
		.mt {
			display: none;
		}
		.marca {
			width: auto;
		}
		.cuerpo {
			flex-direction: column;
		}
		.nav {
			width: auto;
			flex-direction: row;
			overflow-x: auto;
			scrollbar-width: none;
			padding: 0 16px 10px;
			gap: 6px;
		}
		.tit {
			display: none;
		}
		.item {
			height: 40px;
			font-size: 15px;
			border: 1px solid var(--ls);
		}
		.tarjeta {
			margin: 0;
			border-radius: 0;
			padding: 22px 16px 32px;
		}
		.dos,
		.tres,
		.dos-c {
			grid-template-columns: minmax(0, 1fr);
		}
	}
</style>
