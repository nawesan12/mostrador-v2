<script lang="ts">
	import { Dialog } from 'bits-ui';
	import { tick } from 'svelte';
	import { goto, invalidateAll } from '$app/navigation';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import type { IdeaHornero } from '#lib/server/hornero/ideas.ts';
	import { getUi } from '#lib/ui.svelte.ts';
	import { pesos, saludo, variacion } from './formato.ts';
	import { rutaDe } from './navegacion.ts';
	import { getShell } from './shell.svelte.ts';

	/**
	 * Hornero (handoff: panel lateral `ovAi`): cómo viene el día, sus ideas para aprobar y la charla.
	 * Responde con Workers AI sobre los datos de la tienda.
	 */
	const shell = getShell();
	const ui = getUi();

	type Inicio = {
		hola: string | null;
		ventasHoy: number;
		ventasAyer: number;
		ideas: IdeaHornero[];
	};
	let inicio = $state<Inicio | null>(null);
	let escribiendo = $state(false);
	let texto = $state('');
	let lista = $state<HTMLElement>();
	const SUGERENCIAS = ['¿Cuánto vendí ayer?', 'Subí precios 8%', 'Clientes que no vuelven'];

	async function cargar() {
		const r = await fetch('/panel/hornero');
		if (r.ok) inicio = await r.json();
	}

	$effect(() => {
		if (!shell.ia) return;
		void cargar();
		const p = shell.preguntaPendiente;
		if (p) {
			shell.preguntaPendiente = null;
			void enviar(p);
		}
	});

	const bajar = async () => {
		await tick();
		lista?.scrollTo({ top: lista.scrollHeight, behavior: 'smooth' });
	};

	async function enviar(pregunta: string) {
		const q = pregunta.trim();
		if (!q || escribiendo) return;
		texto = '';
		shell.charla.push({ rol: 'persona', texto: q.slice(0, 1000) });
		escribiendo = true;
		void bajar();
		try {
			const r = await fetch('/panel/hornero', {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({ mensajes: shell.charla.slice(-12) })
			});
			if (!r.ok || !r.body) throw new Error(String(r.status));
			shell.charla.push({ rol: 'hornero', texto: '' });
			const yo = shell.charla[shell.charla.length - 1];
			const lector = r.body.pipeThrough(new TextDecoderStream()).getReader();
			for (;;) {
				const { done, value } = await lector.read();
				if (done) break;
				yo.texto += value;
				void bajar();
			}
			if (!yo.texto.trim()) yo.texto = 'No me salió la respuesta. ¿Me lo preguntás de otra forma?';
		} catch {
			shell.charla.push({
				rol: 'hornero',
				texto: 'Uy, se me cortó la conexión. ¿Me lo preguntás de nuevo?'
			});
		} finally {
			escribiendo = false;
			void bajar();
		}
	}

	async function marcar(idea: IdeaHornero, hecha: boolean) {
		const r = await fetch(`/panel/hornero/ideas/${idea.id}`, {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ hecha })
		});
		if (!r.ok) {
			ui.toast({ k: 'err', title: 'No pudimos guardarlo', desc: 'Probá de nuevo en un rato.' });
			return;
		}
		await Promise.all([cargar(), invalidateAll()]);
		if (hecha)
			ui.toast({
				k: 'ai',
				title: idea.listo.titulo,
				desc: idea.listo.detalle,
				action: 'Deshacer',
				onAction: () => marcar(idea, false)
			});
	}

	/** "**dato**" en negrita, sin interpretar HTML. */
	const partes = (t: string) => t.split(/\*\*(.+?)\*\*/g).map((x, i) => ({ t: x, b: i % 2 === 1 }));

	const encabezado = $derived.by(() => {
		if (!inicio) return null;
		const nombre = inicio.hola ? `, ${inicio.hola}` : '';
		const v = variacion(inicio.ventasHoy, inicio.ventasAyer);
		const pend = inicio.ideas.filter((i) => !i.hecha).length;
		const comp = !v
			? ''
			: v.startsWith('−')
				? `, un ${v.slice(1)} menos que ayer a esta hora`
				: v.startsWith('+')
					? `, un ${v.slice(1)} más que ayer a esta hora`
					: ', lo mismo que ayer a esta hora';
		const cierre = !inicio.ideas.length
			? 'No encontré nada urgente. Preguntame lo que quieras de tu tienda.'
			: pend === 0
				? 'Ya aprobaste todo lo que te propuse. Preguntame lo que quieras.'
				: `Encontré ${pend === 1 ? 'una cosa' : `${pend} cosas`} para hacer:`;
		return {
			saludo: `${saludo(new Date())}${nombre}.`,
			ventas: pesos(inicio.ventasHoy),
			comp,
			cierre
		};
	});
</script>

<Dialog.Root bind:open={shell.ia}>
	<Dialog.Portal>
		<Dialog.Overlay class="ho-fondo" />
		<Dialog.Content class="ho-panel" aria-describedby={undefined}>
			<div class="cab">
				<div class="arcos"></div>
				<div class="cab-in">
					<div class="av"><img src="/hornero/avatar.png" alt="" /></div>
					<div class="quien">
						<Dialog.Title class="nombre">Hornero</Dialog.Title>
						<div class="estado"><span class="punto"></span>Tu asistente · en línea</div>
					</div>
					<Dialog.Close class="cerrar" aria-label="Cerrar"
						><Icon icon={I.phXBold} size={15} /></Dialog.Close
					>
				</div>
			</div>

			<div class="lista" bind:this={lista} aria-live="polite">
				{#if encabezado}
					<div class="globo hornero">
						{encabezado.saludo} Hoy vendiste <b>{encabezado.ventas}</b>{encabezado.comp}. {encabezado.cierre}
					</div>
					{#each inicio?.ideas ?? [] as idea (idea.id)}
						<div class="idea">
							<div class="i-fila">
								<span class="i-ico"
									><Icon
										icon={idea.id === 'carritos'
											? I.phShoppingCartSimpleFill
											: idea.id === 'stock'
												? I.phPackageFill
												: I.phMegaphoneFill}
										size={19}
									/></span
								>
								<div class="i-txt">
									<div class="i-t">{idea.titulo}</div>
									<div class="i-d">{idea.detalle}</div>
									<div class="i-n">{idea.nota}</div>
								</div>
							</div>
							{#if idea.hecha}
								<div class="hecho"><Icon icon={I.phCheckCircleFill} size={16} />Hecho</div>
							{:else}
								<div class="i-acc">
									<button type="button" class="aprobar" onclick={() => marcar(idea, true)}
										>Aprobar</button
									>
									<button
										type="button"
										class="editar"
										onclick={() => {
											shell.ia = false;
											// eslint-disable-next-line svelte/no-navigation-without-resolve -- ruta del panel armada con rutaDe
											goto(rutaDe(idea.seccion));
										}}>Editar</button
									>
								</div>
							{/if}
						</div>
					{/each}
				{:else}
					<div class="globo hornero cargando"><span></span><span></span><span></span></div>
				{/if}

				{#each shell.charla as m, i (i)}
					{#if m.texto}
						<div class="globo {m.rol}">
							{#each partes(m.texto) as p, j (j)}{#if p.b}<b>{p.t}</b>{:else}{p.t}{/if}{/each}
						</div>
					{/if}
				{/each}
				{#if escribiendo && !shell.charla.at(-1)?.texto}
					<div class="globo hornero cargando" aria-label="Hornero está escribiendo">
						<span></span><span></span><span></span>
					</div>
				{/if}
			</div>

			<div class="pie">
				<div class="chips">
					{#each SUGERENCIAS as s (s)}
						<button type="button" onclick={() => enviar(s)} disabled={escribiendo}>{s}</button>
					{/each}
				</div>
				<form
					class="escribir"
					onsubmit={(e) => {
						e.preventDefault();
						void enviar(texto);
					}}
				>
					<input
						bind:value={texto}
						placeholder="Escribile a Hornero…"
						aria-label="Escribile a Hornero"
						maxlength="1000"
						autocomplete="off"
					/>
					<button class="mandar" aria-label="Mandar" disabled={escribiendo || !texto.trim()}
						><Icon icon={I.phArrowUpBold} size={15} /></button
					>
				</form>
			</div>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

<style>
	:global(.ho-fondo) {
		position: fixed;
		inset: 0;
		z-index: 70;
		background: rgba(42, 20, 12, 0.3);
		backdrop-filter: blur(2px);
	}
	:global(.ho-fondo[data-state='open']) {
		animation: ho-fade 260ms both;
	}
	:global(.ho-panel) {
		position: fixed;
		top: 12px;
		right: 12px;
		bottom: 12px;
		z-index: 71;
		width: 440px;
		max-width: calc(100vw - 24px);
		border-radius: 26px;
		background: var(--bg);
		color: var(--tx);
		box-shadow: 0 40px 90px -20px rgba(42, 20, 12, 0.6);
		display: flex;
		flex-direction: column;
		overflow: hidden;
		outline: none;
	}
	:global(.ho-panel[data-state='open']) {
		animation: ho-entra 480ms cubic-bezier(0.16, 1, 0.3, 1) both;
	}
	@keyframes ho-fade {
		from {
			opacity: 0;
		}
	}
	@keyframes ho-entra {
		from {
			transform: translateX(105%);
		}
	}
	.cab {
		position: relative;
		padding: 18px 18px 20px;
		background: linear-gradient(140deg, #d2582f, #b9481f);
		color: #fff;
		overflow: hidden;
	}
	.arcos {
		position: absolute;
		inset: 0;
		background-image: url('/textura/arch-pattern-light.svg');
	}
	.cab-in {
		position: relative;
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.av {
		width: 56px;
		height: 56px;
		flex: none;
		border-radius: 50%;
		overflow: hidden;
		background: #fbf7f2;
		box-shadow: 0 0 0 3px rgba(255, 255, 255, 0.35);
	}
	.av img {
		display: block;
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.quien {
		flex: 1;
	}
	.quien :global(.nombre) {
		margin: 0;
		font-size: 18px;
		font-weight: 600;
	}
	.estado {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 13px;
		color: #fce3d6;
	}
	.punto {
		width: 7px;
		height: 7px;
		border-radius: 50%;
		background: #8fe0b5;
	}
	.cab :global(.cerrar) {
		width: 36px;
		height: 36px;
		border: 0;
		border-radius: 50%;
		background: rgba(255, 255, 255, 0.16);
		color: #fff;
		display: grid;
		place-items: center;
	}
	.lista {
		flex: 1;
		overflow: auto;
		padding: 16px;
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	.globo {
		max-width: 90%;
		padding: 12px 14px;
		font-size: 14.5px;
		line-height: 1.5;
		white-space: pre-wrap;
		overflow-wrap: anywhere;
	}
	.globo.hornero {
		align-self: flex-start;
		border-radius: 4px 18px 18px 18px;
		background: var(--sf);
		box-shadow: var(--sh);
	}
	.globo.persona {
		align-self: flex-end;
		max-width: 85%;
		padding: 11px 14px;
		border-radius: 18px 18px 4px 18px;
		background: var(--tx);
		color: var(--sf);
		font-size: 14px;
		line-height: 1.45;
	}
	.cargando {
		display: flex;
		gap: 5px;
		padding: 16px;
	}
	.cargando span {
		width: 7px;
		height: 7px;
		border-radius: 50%;
		background: var(--fa);
		animation: puntos 1.1s infinite;
	}
	.cargando span:nth-child(2) {
		animation-delay: 0.15s;
	}
	.cargando span:nth-child(3) {
		animation-delay: 0.3s;
	}
	@keyframes puntos {
		0%,
		100% {
			opacity: 0.25;
		}
		50% {
			opacity: 1;
		}
	}
	.idea {
		padding: 14px;
		border-radius: 18px;
		background: var(--sf);
		box-shadow: var(--sh);
		border: 1px solid var(--ln);
	}
	.i-fila {
		display: flex;
		gap: 12px;
	}
	.i-ico {
		width: 38px;
		height: 38px;
		flex: none;
		border-radius: 12px;
		background: rgba(210, 88, 47, 0.12);
		color: #b0441d;
		display: grid;
		place-items: center;
	}
	.i-txt {
		flex: 1;
		min-width: 0;
	}
	.i-t {
		font-size: 14.5px;
		font-weight: 600;
	}
	.i-d {
		font-size: 13px;
		color: var(--mu);
		margin-top: 2px;
		line-height: 1.4;
	}
	.i-n {
		font-size: 12.5px;
		font-weight: 600;
		color: #8a5f0e;
		margin-top: 6px;
	}
	.i-acc {
		margin-top: 12px;
		display: flex;
		gap: 8px;
	}
	.aprobar,
	.editar {
		padding: 9px 14px;
		border-radius: 999px;
		font-size: 13px;
		font-weight: 600;
	}
	.aprobar {
		flex: 1;
		border: 0;
		background: var(--tx);
		color: var(--sf);
	}
	.editar {
		border: 1px solid var(--ln2);
		background: transparent;
		color: var(--tx);
	}
	.hecho {
		margin-top: 12px;
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 13px;
		font-weight: 600;
		color: #237550;
	}
	.pie {
		padding: 12px 14px 14px;
		padding-bottom: calc(14px + env(safe-area-inset-bottom));
		background: var(--sf);
		border-top: 1px solid var(--ln);
	}
	.chips {
		display: flex;
		gap: 6px;
		overflow-x: auto;
		scrollbar-width: none;
		margin-bottom: 10px;
	}
	.chips button {
		font-size: 12.5px;
		font-weight: 500;
		padding: 6px 11px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: transparent;
		color: var(--mu);
		white-space: nowrap;
	}
	.chips button:hover:not(:disabled) {
		color: var(--tx);
		border-color: var(--tx);
	}
	.escribir {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 6px 6px 6px 14px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: var(--bg);
	}
	.escribir:focus-within {
		border-color: #d2582f;
		box-shadow: 0 0 0 3px rgba(210, 88, 47, 0.18);
	}
	.escribir input {
		flex: 1;
		min-width: 0;
		border: 0;
		outline: 0;
		background: transparent;
		font-size: 14px;
		color: var(--tx);
	}
	.escribir input::placeholder {
		color: var(--fa);
	}
	.mandar {
		width: 34px;
		height: 34px;
		flex: none;
		border: 0;
		border-radius: 50%;
		background: #b0441d;
		color: #fff;
		display: grid;
		place-items: center;
	}
	.mandar:disabled {
		opacity: 0.5;
	}
	@media (prefers-reduced-motion: reduce) {
		:global(.ho-panel[data-state='open']),
		.cargando span {
			animation: none;
		}
	}
</style>
