<script lang="ts">
	import type { Casilla } from '#lib/server/bandeja.ts';
	import { faltaAdjunto, mandar, subir, type Borrador } from '#lib/redactar.svelte.ts';
	import { iniciales, tamano, tipoAdjunto } from '#lib/formato.ts';
	import { COLORES } from './Carpetas.svelte';

	let {
		b,
		casillas,
		contactos,
		forma,
		oncerrar,
		onenviado
	}: {
		b: Borrador;
		casillas: Casilla[];
		contactos: { nombre: string; email: string }[];
		forma: 'ventana' | 'adentro';
		oncerrar: () => void;
		onenviado: (programado: boolean) => void;
	} = $props();

	let escribiendo = $state('');
	let conCc = $state(false);
	let menu = $state(false);
	let manual = $state(false);
	let cuandoManual = $state('');
	let archivo = $state<HTMLInputElement>();
	let subiendo = $state(false);
	let citaVisible = $state(false);

	const casilla = $derived(casillas.find((c) => c.cuenta === b.cuenta));
	// La firma de la casilla desde la que sale: se agrega al final y cambia si se cambia la casilla.
	let firma = '';
	$effect(() => {
		const cuenta = b.cuenta;
		fetch(`/firma?cuenta=${encodeURIComponent(cuenta)}`)
			.then((r) => (r.ok ? r.json() : { firma: '' }))
			.then(({ firma: nueva }: { firma: string }) => {
				const sep = '\n\n-- \n';
				const sin =
					firma && b.texto.endsWith(sep + firma)
						? b.texto.slice(0, -(sep + firma).length)
						: b.texto;
				firma = nueva;
				b.texto = nueva ? sin + sep + nueva : sin;
			})
			.catch(() => {});
	});
	const color = (cuenta: string) =>
		COLORES[
			Math.max(
				0,
				casillas.findIndex((c) => c.cuenta === cuenta)
			) % COLORES.length
		];
	const sugeridos = $derived(
		escribiendo.trim().length < 2
			? []
			: contactos
					.filter(
						(c) =>
							!b.para.includes(c.email) &&
							(c.nombre + ' ' + c.email).toLowerCase().includes(escribiendo.trim().toLowerCase())
					)
					.slice(0, 4)
	);
	const nombreDe = (email: string) => contactos.find((c) => c.email === email)?.nombre || email;

	function agregar(email: string) {
		const e = email.trim().replace(/[,;]$/, '');
		if (e && !b.para.includes(e)) b.para.push(e);
		escribiendo = '';
	}
	function tecla(ev: KeyboardEvent) {
		if (
			(ev.key === 'Enter' || ev.key === ',' || ev.key === ';' || ev.key === 'Tab') &&
			escribiendo.trim()
		) {
			if (ev.key !== 'Tab') ev.preventDefault();
			agregar(sugeridos[0]?.email ?? escribiendo);
		} else if (ev.key === 'Backspace' && !escribiendo && b.para.length) b.para.pop();
	}

	/** Los horarios propuestos (2o): mañana a la apertura y después del mediodía, o el lunes. */
	const propuestos = $derived.by(() => {
		const out: { t: string; d: string; iso: string }[] = [];
		const base = new Date();
		const dia = new Date(base);
		dia.setDate(dia.getDate() + 1);
		while (dia.getDay() === 0 || dia.getDay() === 6) dia.setDate(dia.getDate() + 1);
		const fmt = (d: Date) =>
			d
				.toLocaleDateString('es-AR', { weekday: 'long', day: 'numeric' })
				.replace(/^./, (x) => x.toUpperCase());
		for (const [h, d] of [
			[8, 'a primera hora'],
			[14, 'después del mediodía']
		] as const) {
			const x = new Date(dia);
			x.setHours(h, 0, 0, 0);
			out.push({ t: `${fmt(x)} a las ${h}:00`, d, iso: x.toISOString() });
		}
		const otro = new Date(dia);
		otro.setDate(otro.getDate() + (otro.getDay() === 5 ? 3 : 1));
		otro.setHours(8, 0, 0, 0);
		out.push({ t: `${fmt(otro)} a las 8:00`, d: '', iso: otro.toISOString() });
		return out;
	});

	async function enviar(programado: string | null = null) {
		if (escribiendo.trim()) agregar(escribiendo);
		menu = false;
		if (await mandar(b, programado)) onenviado(!!programado);
	}
	async function adjuntar(ev: Event) {
		const fs = (ev.currentTarget as HTMLInputElement).files;
		if (!fs?.length) return;
		subiendo = true;
		await subir(b, fs);
		subiendo = false;
		(ev.currentTarget as HTMLInputElement).value = '';
	}
	const titulo = $derived(
		b.modo === 'responder' ? 'Responder' : b.modo === 'reenviar' ? 'Reenviar' : 'Mensaje nuevo'
	);
</script>

<div class="ed {forma}" class:min={b.minimizada && forma === 'ventana'}>
	{#if forma === 'ventana'}
		<div class="cab">
			<button class="tit" onclick={() => (b.minimizada = !b.minimizada)}
				>{b.minimizada ? b.asunto || titulo : titulo}</button
			>
			<button
				aria-label={b.minimizada ? 'Abrir' : 'Minimizar'}
				onclick={() => (b.minimizada = !b.minimizada)}
				><i class={b.minimizada ? 'icon-chevron-up' : 'icon-minus'}></i></button
			>
			<button aria-label="Cerrar" onclick={oncerrar}><i class="icon-x"></i></button>
		</div>
	{:else}
		<div class="cab-in">
			<span class="modo"
				><i class={b.modo === 'reenviar' ? 'icon-forward' : 'icon-reply'}></i>{titulo}</span
			>
			{#each b.para as p (p)}
				<span class="chip"
					><span class="av">{iniciales(nombreDe(p))}</span>{nombreDe(p)}<button
						aria-label="Quitar {p}"
						onclick={() => (b.para = b.para.filter((x) => x !== p))}><i class="icon-x"></i></button
					></span
				>
			{/each}
			<input
				class="mas"
				bind:value={escribiendo}
				onkeydown={tecla}
				onblur={() => escribiendo.trim() && agregar(escribiendo)}
				placeholder={b.para.length ? 'Agregar otro…' : 'Para…'}
				aria-label="Para"
			/>
		</div>
	{/if}

	{#if !(b.minimizada && forma === 'ventana')}
		{#if forma === 'ventana'}
			<div class="campo">
				<span class="et">De</span>
				<span class="de">
					<span class="punto" style:background={color(b.cuenta)}></span>
					<select
						bind:value={b.cuenta}
						aria-label="Desde qué casilla"
						disabled={!!b.adjuntos.length}
					>
						{#each casillas as c (c.cuenta)}
							<option value={c.cuenta}>{c.nombre.replace(' (vos)', '')} — {c.email}</option>
						{/each}
					</select>
					<i class="icon-chevron-down"></i>
				</span>
			</div>
			<div class="campo para">
				<span class="et">Para</span>
				{#each b.para as p (p)}
					<span class="chip"
						><span class="av">{iniciales(nombreDe(p))}</span>{nombreDe(p)}<button
							aria-label="Quitar {p}"
							onclick={() => (b.para = b.para.filter((x) => x !== p))}
							><i class="icon-x"></i></button
						></span
					>
				{/each}
				<input
					bind:value={escribiendo}
					onkeydown={tecla}
					onblur={() => setTimeout(() => escribiendo.trim() && agregar(escribiendo), 150)}
					aria-label="Para"
					autocomplete="off"
				/>
				{#if !conCc}<button class="cc" onclick={() => (conCc = true)}>Cc</button>{/if}
				{#if sugeridos.length}
					<div class="sug vidrio" role="listbox">
						{#each sugeridos as s, i (s.email)}
							<button
								class="s"
								class:act={i === 0}
								role="option"
								aria-selected={i === 0}
								onmousedown={(e) => (e.preventDefault(), agregar(s.email))}
							>
								<span class="sav">{iniciales(s.nombre || s.email)}</span>
								<span class="stx"
									><span class="sn">{s.nombre || s.email}</span><span class="sd">{s.email}</span
									></span
								>
							</button>
						{/each}
					</div>
				{/if}
			</div>
			{#if conCc}
				<div class="campo">
					<span class="et">Cc</span>
					<input
						value={b.cc.join(', ')}
						oninput={(e) =>
							(b.cc = e.currentTarget.value
								.split(/[,;]/)
								.map((x) => x.trim())
								.filter(Boolean))}
						aria-label="Con copia"
					/>
				</div>
			{/if}
			<div class="campo">
				<span class="et">Asunto</span>
				<input bind:value={b.asunto} aria-label="Asunto" />
			</div>
		{/if}

		<div class="cuerpo">
			<!-- svelte-ignore a11y_autofocus -->
			<textarea
				bind:value={b.texto}
				aria-label="Mensaje"
				autofocus={forma === 'adentro'}
				placeholder={forma === 'adentro' ? '' : 'Escribí el mensaje'}></textarea>
			{#if b.adjuntos.length}
				<div class="adjs">
					{#each b.adjuntos as a, i (a.blobId)}
						{@const t = tipoAdjunto(a.tipo, a.nombre)}
						<span class="adj"
							><i class={t.ic} style:color={t.c}></i>{a.nombre}<button
								aria-label="Quitar {a.nombre}"
								onclick={() => b.adjuntos.splice(i, 1)}><i class="icon-x"></i></button
							></span
						>
					{/each}
				</div>
			{/if}
			{#if b.cita && b.modo === 'responder'}
				<button class="ocultar" onclick={() => (citaVisible = !citaVisible)}
					><i class={citaVisible ? 'icon-chevron-up' : 'icon-chevron-down'}></i>{citaVisible
						? 'Ocultar el mensaje anterior'
						: 'Ver el mensaje anterior'}</button
				>
				{#if citaVisible}<div class="cita">{b.cita}</div>{/if}
			{:else if b.cita}
				<div class="reenviado">{b.cita}</div>
			{/if}
		</div>

		{#if faltaAdjunto(b)}
			<div class="aviso">
				<i class="icon-paperclip"></i><span>Dice «adjunto», pero no agregaste ningún archivo.</span
				><button onclick={() => archivo?.click()}>Adjuntar</button>
			</div>
		{/if}
		{#if b.error}<div class="error" role="alert">{b.error}</div>{/if}

		<div class="pie">
			<span class="enviar">
				<button class="e" disabled={b.enviando} onclick={() => enviar()}
					>{b.enviando ? 'Enviando…' : 'Enviar'}</button
				>
				<button
					class="f"
					aria-label="Programar el envío"
					aria-expanded={menu}
					onclick={() => (menu = !menu)}><i class="icon-chevron-down"></i></button
				>
				{#if menu}
					<div class="prog vidrio">
						<div class="pt">Programar el envío</div>
						{#each propuestos as p, i (p.iso)}
							<button class="op" class:act={i === 0} onclick={() => enviar(p.iso)}
								><span class="opt">{p.t}</span>{#if p.d}<span class="opd">{p.d}</span>{/if}</button
							>
						{/each}
						{#if manual}
							<div class="manual">
								<input type="datetime-local" bind:value={cuandoManual} aria-label="Día y hora" />
								<button
									disabled={!cuandoManual}
									onclick={() => enviar(new Date(cuandoManual).toISOString())}>Programar</button
								>
							</div>
						{:else}
							<button class="op otro" onclick={() => (manual = true)}
								><i class="icon-calendar"></i>Elegir día y hora</button
							>
						{/if}
					</div>
				{/if}
			</span>
			<button class="her" onclick={() => archivo?.click()} disabled={subiendo}
				><i class="icon-paperclip"></i>{forma === 'ventana'
					? subiendo
						? 'Subiendo…'
						: 'Adjuntar'
					: ''}</button
			>
			<input type="file" multiple hidden bind:this={archivo} onchange={adjuntar} />
			<span class="estado"
				>{b.adjuntos.length
					? `${b.adjuntos.length} ${b.adjuntos.length === 1 ? 'adjunto' : 'adjuntos'} · ${tamano(b.adjuntos.reduce((n, a) => n + a.tamano, 0))}`
					: casilla && forma === 'adentro'
						? `Sale de ${casilla.email}`
						: ''}</span
			>
			<button class="borrar" aria-label="Descartar" onclick={oncerrar}
				><i class="icon-trash-2"></i></button
			>
		</div>
	{/if}
</div>

<style>
	.ed {
		display: flex;
		flex-direction: column;
		background-color: var(--card);
		backdrop-filter: var(--blur);
		-webkit-backdrop-filter: var(--blur);
	}
	.ventana {
		width: 664px;
		height: min(760px, calc(100dvh - 90px));
		border-radius: 14px 14px 0 0;
		box-shadow:
			var(--popsh),
			0 0 0 1px var(--ls);
	}
	.ventana.min {
		width: 300px;
		height: 50px;
	}
	.adentro {
		border-radius: 14px;
		box-shadow:
			0 0 0 1.5px var(--acc),
			0 8px 24px -12px rgb(60 30 10 / 0.3);
		overflow: visible;
	}
	.cab {
		height: 52px;
		flex: none;
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 0 10px 0 20px;
		background: var(--toast);
		color: var(--toastT);
		border-radius: 14px 14px 0 0;
		font-size: 16px;
		font-weight: 600;
	}
	.min .cab {
		height: 50px;
		border-radius: 12px 12px 0 0;
		font-weight: 500;
		padding: 0 14px 0 18px;
	}
	.cab .tit {
		flex: 1;
		text-align: left;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		font-weight: inherit;
	}
	.cab button:not(.tit) {
		padding: 8px;
		font-size: 20px;
	}
	.cab-in {
		display: flex;
		align-items: center;
		flex-wrap: wrap;
		gap: 10px;
		min-height: 52px;
		padding: 8px 14px;
		border-bottom: 1px solid var(--ls);
		font-size: 16px;
	}
	.modo {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		font-weight: 600;
		color: var(--accT);
	}
	.modo i {
		font-size: 18px;
	}
	.mas {
		flex: 1;
		min-width: 120px;
		border: 0;
		outline: 0;
		background: transparent;
		font-size: 16px;
	}
	.mas::placeholder {
		color: var(--t3);
	}
	.campo {
		position: relative;
		display: flex;
		align-items: center;
		flex-wrap: wrap;
		gap: 10px;
		min-height: 54px;
		padding: 6px 20px;
		border-bottom: 1px solid var(--ls);
		font-size: 16px;
	}
	.et {
		color: var(--mut);
		width: 56px;
		flex: none;
	}
	.campo input {
		flex: 1;
		min-width: 80px;
		border: 0;
		outline: 0;
		background: transparent;
		font-size: 16px;
	}
	.de {
		position: relative;
		display: inline-flex;
		align-items: center;
		gap: 10px;
		font-weight: 500;
	}
	.de select {
		appearance: none;
		border: 0;
		background: transparent;
		font-weight: 500;
		font-size: 16px;
		padding-right: 22px;
		margin-right: -22px;
		cursor: pointer;
	}
	.de select option {
		background: var(--pop);
		color: var(--fg);
	}
	.de i {
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
	.chip {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		height: 34px;
		padding: 0 10px 0 4px;
		border-radius: 999px;
		background: var(--mu);
		font-size: 16px;
		max-width: 100%;
	}
	.cab-in .chip {
		height: 32px;
		padding: 0 10px 0 3px;
		font-size: 15px;
	}
	.chip button {
		color: var(--mut);
		font-size: 15px;
		display: inline-flex;
	}
	.av {
		width: 26px;
		height: 26px;
		border-radius: 50%;
		background: var(--avC);
		color: var(--avCT);
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 12px;
		font-weight: 600;
		flex: none;
	}
	.cc {
		margin-left: auto;
		font-size: 15px;
		color: var(--mut);
	}
	.sug {
		position: absolute;
		top: 50px;
		left: 84px;
		width: min(440px, calc(100% - 100px));
		z-index: 3;
		background: var(--pop);
		border-radius: 12px;
		box-shadow:
			var(--popsh),
			0 0 0 1px var(--ls);
		overflow: hidden;
		padding: 6px 0;
	}
	.s {
		width: 100%;
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 10px 16px;
		text-align: left;
	}
	.s.act,
	.s:hover {
		background: var(--sel);
	}
	.sav {
		width: 38px;
		height: 38px;
		flex: none;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 15px;
		font-weight: 600;
		background: var(--mu);
		color: var(--mut);
	}
	.stx {
		flex: 1;
		min-width: 0;
		line-height: 1.35;
	}
	.sn {
		display: block;
		font-size: 16px;
		font-weight: 500;
	}
	.sd {
		display: block;
		font-size: 14px;
		color: var(--mut);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.cuerpo {
		flex: 1;
		min-height: 0;
		display: flex;
		flex-direction: column;
		gap: 12px;
		padding: 18px 20px 10px;
		overflow-y: auto;
	}
	.adentro .cuerpo {
		padding: 14px 16px;
	}
	textarea {
		flex: 1;
		min-height: 160px;
		border: 0;
		outline: 0;
		resize: none;
		background: transparent;
		font-size: 17px;
		line-height: 1.65;
	}
	.adentro textarea {
		min-height: 120px;
		line-height: 1.6;
	}
	.adjs {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}
	.adj {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		height: 36px;
		padding: 0 10px;
		border: 1px solid var(--ls);
		border-radius: 10px;
		font-size: 14px;
		max-width: 100%;
	}
	.adj i {
		font-size: 17px;
	}
	.adj button {
		color: var(--mut);
		display: inline-flex;
		font-size: 14px;
	}
	.ocultar {
		align-self: flex-start;
		display: inline-flex;
		align-items: center;
		gap: 6px;
		height: 30px;
		padding: 0 12px;
		border-radius: 999px;
		background: var(--mu);
		font-size: 14px;
		color: var(--mut);
	}
	.cita {
		border-left: 3px solid var(--ls);
		padding: 2px 0 2px 14px;
		font-size: 15px;
		line-height: 1.6;
		color: var(--mut);
		white-space: pre-wrap;
	}
	.reenviado {
		padding: 12px 14px;
		border-radius: 10px;
		background: var(--hun);
		font-size: 15px;
		line-height: 1.55;
		color: var(--mut);
		white-space: pre-wrap;
		max-height: 220px;
		overflow: auto;
	}
	.aviso {
		margin: 0 20px 12px;
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 12px 16px;
		border-radius: 12px;
		background: var(--espBg);
		color: var(--espT);
		font-size: 16px;
	}
	.aviso span {
		flex: 1;
	}
	.aviso i {
		font-size: 20px;
	}
	.aviso button {
		font-weight: 600;
		color: inherit;
	}
	.error {
		margin: 0 20px 12px;
		padding: 10px 14px;
		border-radius: 10px;
		background: var(--prBg);
		color: var(--prT);
		font-size: 15px;
	}
	.pie {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 12px 16px 16px;
	}
	.adentro .pie {
		padding: 10px 12px;
		border-top: 1px solid var(--ls);
	}
	.enviar {
		position: relative;
		display: inline-flex;
		align-items: stretch;
		height: 48px;
		border-radius: 999px;
		background: var(--acc);
		color: #fff;
		font-size: 17px;
		font-weight: 600;
	}
	.adentro .enviar {
		height: 46px;
		font-size: 16px;
	}
	.enviar .e {
		display: inline-flex;
		align-items: center;
		padding: 0 22px 0 26px;
		color: #fff;
		font-weight: 600;
	}
	.enviar .f {
		display: inline-flex;
		align-items: center;
		padding: 0 14px 0 12px;
		border-left: 1px solid oklch(1 0 0 / 0.3);
		color: #fff;
		font-size: 18px;
	}
	.prog {
		position: absolute;
		bottom: 56px;
		left: 0;
		width: 340px;
		z-index: 4;
		background: var(--pop);
		color: var(--fg);
		border-radius: 12px;
		box-shadow:
			var(--popsh),
			0 0 0 1px var(--ls);
		padding: 6px 0;
		font-weight: 400;
	}
	.pt {
		padding: 10px 16px 6px;
		font-size: 14px;
		font-weight: 600;
		color: var(--mut);
	}
	.op {
		width: 100%;
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		padding: 10px 16px;
		text-align: left;
		font-size: 16px;
	}
	.op.act,
	.op:hover {
		background: var(--sel);
		box-shadow: inset 3px 0 0 var(--acc);
	}
	.opd {
		font-size: 14px;
		color: var(--mut);
	}
	.op.otro {
		flex-direction: row;
		align-items: center;
		gap: 10px;
		color: var(--accT);
		font-weight: 500;
		border-top: 1px solid var(--ls);
	}
	.manual {
		display: flex;
		gap: 8px;
		padding: 10px 16px;
		border-top: 1px solid var(--ls);
	}
	.manual input {
		flex: 1;
		min-width: 0;
		height: 40px;
		padding: 0 10px;
		border-radius: 10px;
		border: 1px solid var(--ln);
		background: transparent;
		color-scheme: dark light;
	}
	.manual button {
		height: 40px;
		padding: 0 14px;
		border-radius: 10px;
		background: var(--acc);
		color: #fff;
		font-weight: 600;
	}
	.manual button:disabled {
		opacity: 0.5;
	}
	.her {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		height: 44px;
		padding: 0 12px;
		margin-left: 8px;
		border-radius: 10px;
		font-size: 15px;
	}
	.her i {
		font-size: 20px;
		color: var(--mut);
	}
	.estado {
		margin-left: auto;
		font-size: 14px;
		color: var(--mut);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		min-width: 0;
	}
	.borrar {
		padding: 10px;
		color: var(--mut);
		font-size: 20px;
	}
	@media (max-width: 760px) {
		.ventana {
			width: 100%;
			height: 100dvh;
			border-radius: 0;
		}
		.ventana.min {
			width: 100%;
		}
		.cab {
			border-radius: 0;
			height: calc(52px + env(safe-area-inset-top));
			padding-top: env(safe-area-inset-top);
		}
		.pie {
			padding-bottom: calc(16px + env(safe-area-inset-bottom));
		}
		.pie {
			flex-wrap: wrap;
		}
	}
</style>
