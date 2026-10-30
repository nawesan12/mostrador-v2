<script lang="ts">
	import type { Casilla, Mensaje } from '#lib/server/bandeja.ts';
	import { cuando, iniciales, tamano, tipoAdjunto } from '#lib/formato.ts';
	import { nuevo, type Borrador } from '#lib/redactar.svelte.ts';
	import Editor from './Editor.svelte';

	let {
		m,
		casillas,
		contactos,
		hrefImagenes,
		onaccion,
		onvolver,
		onenviado
	}: {
		m: Mensaje;
		casillas: Casilla[];
		contactos: { nombre: string; email: string }[];
		hrefImagenes: string;
		onaccion: (que: string) => void;
		onvolver: () => void;
		onenviado: (programado: boolean) => void;
	} = $props();

	let escribiendo = $state<Borrador | null>(null);
	let verHilo = $state(false);
	let marco = $state<HTMLIFrameElement>();
	let alto = $state(200);
	let menu = $state(false);

	const nombre = (d: { name: string | null; email: string }) => d.name || d.email;
	const urlAdj = (a: Mensaje['adjuntos'][number], ver = false) =>
		`/adjunto/${encodeURIComponent(m.cuenta)}/${encodeURIComponent(a.blob)}?n=${encodeURIComponent(a.nombre)}&t=${encodeURIComponent(a.tipo)}${ver ? '&ver=1' : ''}`;
	const pesoTotal = $derived(m.adjuntos.reduce((n, a) => n + a.tamano, 0));
	const parrafos = $derived(
		m.texto
			.replace(/\r/g, '')
			.split(/\n{2,}/)
			.map((p) => p.trim())
			.filter(Boolean)
	);

	/** El HTML del correo en un documento aparte, con la letra y los colores de la app. */
	const doc = $derived(
		m.html
			? `<!doctype html><html><head><meta charset="utf-8"><base target="_blank"><style>
html,body{margin:0;background:#fff;color:#1d1d1f;font:17px/1.65 Geist,ui-sans-serif,system-ui,sans-serif;overflow-wrap:anywhere}
body{padding:20px 22px}a{color:#1d5fd1}img{max-width:100%;height:auto}table{max-width:100%}blockquote{margin:0 0 0 4px;padding-left:14px;border-left:3px solid rgba(127,127,127,.35)}
</style></head><body>${m.html}</body></html>`
			: ''
	);
	function medir() {
		const d = marco?.contentDocument;
		if (d) alto = Math.max(80, d.documentElement.scrollHeight);
	}

	const cita = () =>
		`El ${cuando(m.fecha).toLowerCase()}, ${nombre(m.de)} escribió:\n` +
		m.texto
			.trim()
			.split('\n')
			.map((l) => `> ${l}`)
			.join('\n');
	const asuntoCon = (pre: string) =>
		new RegExp(`^${pre}:`, 'i').test(m.asunto) ? m.asunto : `${pre}: ${m.asunto}`;

	function responder() {
		escribiendo = nuevo({
			modo: 'responder',
			cuenta: m.cuenta,
			para: [m.de.email],
			asunto: asuntoCon('Re'),
			responde: m.id,
			cita: cita()
		});
	}
	function reenviar() {
		escribiendo = nuevo({
			modo: 'reenviar',
			cuenta: m.cuenta,
			asunto: asuntoCon('Fwd'),
			adjuntos: m.adjuntos.map((a) => ({
				blobId: a.blob,
				nombre: a.nombre,
				tipo: a.tipo,
				tamano: a.tamano
			})),
			cita: `Mensaje reenviado\nDe: ${nombre(m.de)} · ${m.de.email}\nFecha: ${cuando(m.fecha)}\nAsunto: ${m.asunto}\n\n${m.texto.trim()}`
		});
	}
	const ACCIONES = [
		{ l: 'Responder', ic: 'icon-reply', f: responder },
		{ l: 'Reenviar', ic: 'icon-forward', f: reenviar },
		{ l: 'Archivar', ic: 'icon-archive', f: () => onaccion('archivar') },
		{ l: 'Borrar', ic: 'icon-trash-2', f: () => onaccion('borrar') }
	];
</script>

<section class="leer">
	<div class="barra">
		<button class="volver" aria-label="Volver a la lista" onclick={onvolver}
			><i class="icon-arrow-left"></i></button
		>
		{#each ACCIONES as a (a.l)}
			<button class="acc" onclick={a.f}><i class={a.ic}></i><span>{a.l}</span></button>
		{/each}
		<span class="mas-w">
			<button
				class="mas"
				aria-label="Más acciones"
				aria-expanded={menu}
				onclick={() => (menu = !menu)}><i class="icon-ellipsis-vertical"></i></button
			>
			{#if menu}
				<div class="menu vidrio" role="menu">
					<button role="menuitem" onclick={() => ((menu = false), onaccion('no-leido'))}
						><i class="icon-mail"></i>Marcar como no leído</button
					>
					<button
						role="menuitem"
						onclick={() => (
							(menu = false),
							onaccion(m.destacado ? 'quitar-destacado' : 'destacar')
						)}><i class="icon-star"></i>{m.destacado ? 'Quitar de destacados' : 'Destacar'}</button
					>
					<button role="menuitem" onclick={() => ((menu = false), onaccion('spam'))}
						><i class="icon-shield-alert"></i>Es spam</button
					>
					<button role="menuitem" onclick={() => ((menu = false), window.print())}
						><i class="icon-printer"></i>Imprimir</button
					>
				</div>
			{/if}
		</span>
	</div>

	<div class="cont">
		<h1>{m.asunto}</h1>
		{#if m.anteriores.length}
			<div class="hilo">
				<span class="avs"
					>{#each m.anteriores.slice(0, 2) as a (a.id)}<span class="mini"
							>{iniciales(nombre(a.de))}</span
						>{/each}</span
				>
				<span class="ht"
					>{m.anteriores.length}
					{m.anteriores.length === 1 ? 'mensaje anterior' : 'mensajes anteriores'} en la conversación</span
				>
				<button class="ver" onclick={() => (verHilo = !verHilo)}
					>{verHilo ? 'Ocultar' : 'Ver todos'}</button
				>
			</div>
			{#if verHilo}
				<div class="anteriores">
					{#each m.anteriores as a (a.id)}
						<a class="ant" href="?m={encodeURIComponent(a.id)}" data-sveltekit-reset="false">
							<b>{nombre(a.de)}</b><span class="af">{cuando(a.fecha)}</span>
							<span class="av2">{a.vista}</span>
						</a>
					{/each}
				</div>
			{/if}
		{/if}

		<div class="remite">
			<div class="avg">{iniciales(nombre(m.de))}</div>
			<div class="rt">
				<div class="rn">{nombre(m.de)}</div>
				<div class="rd">
					{m.de.email} · para {m.para.map((p) => p.email).join(', ') || m.casilla}{#if m.cc.length}
						· cc {m.cc.map((p) => p.email).join(', ')}{/if}
				</div>
			</div>
			<span class="rf">{cuando(m.fecha)}</span>
			<button
				class="estrella"
				class:on={m.destacado}
				aria-label={m.destacado ? 'Quitar de destacados' : 'Destacar'}
				onclick={() => onaccion(m.destacado ? 'quitar-destacado' : 'destacar')}
				><i class="icon-star"></i></button
			>
		</div>

		{#if m.remotas}
			<div class="ocultas">
				<i class="icon-image-off"></i><span>Las imágenes de este correo están ocultas.</span><a
					href={hrefImagenes}
					data-sveltekit-reset="false">Mostrar imágenes</a
				>
			</div>
		{/if}

		{#if m.html}
			<iframe
				bind:this={marco}
				title="Contenido del correo"
				sandbox="allow-same-origin allow-popups allow-popups-to-escape-sandbox"
				srcdoc={doc}
				style:height="{alto}px"
				onload={medir}
			></iframe>
		{:else}
			<div class="texto">
				{#each parrafos as p, i (i)}<div>{p}</div>{/each}
			</div>
		{/if}

		{#if m.adjuntos.length}
			<div class="adjs">
				<div class="adjc">
					<i class="icon-paperclip"></i>
					<b>{m.adjuntos.length} {m.adjuntos.length === 1 ? 'adjunto' : 'adjuntos'}</b><span
						>{tamano(pesoTotal)}</span
					>
				</div>
				<div class="grilla">
					{#each m.adjuntos as a (a.blob + a.parte)}
						{@const t = tipoAdjunto(a.tipo, a.nombre)}
						<div class="adj">
							<div class="vp">
								{#if a.tipo.startsWith('image/')}
									<img src={urlAdj(a)} alt="" loading="lazy" />
								{:else}
									<i class={t.ic} style:color={t.c}></i>
								{/if}
								<div class="hov">
									<a class="b-ver" href={urlAdj(a, true)} target="_blank" rel="noopener"
										><i class="icon-eye"></i>Ver</a
									>
									<a
										class="b-dl"
										href={urlAdj(a)}
										download={a.nombre}
										aria-label="Descargar {a.nombre}"><i class="icon-download"></i></a
									>
								</div>
							</div>
							<div class="pie">
								<i class={t.ic} style:color={t.c}></i>
								<div class="pt">
									<div class="pn">{a.nombre}</div>
									<div class="pm">{t.l} · {tamano(a.tamano)}</div>
								</div>
							</div>
						</div>
					{/each}
				</div>
			</div>
		{/if}

		{#if escribiendo}
			<Editor
				b={escribiendo}
				{casillas}
				{contactos}
				forma="adentro"
				oncerrar={() => (escribiendo = null)}
				onenviado={(p) => ((escribiendo = null), onenviado(p))}
			/>
		{:else}
			<div class="botones">
				<button class="b" onclick={responder}><i class="icon-reply"></i>Responder</button>
				<button class="b" onclick={reenviar}><i class="icon-forward"></i>Reenviar</button>
			</div>
		{/if}
	</div>
</section>

<style>
	.leer {
		display: flex;
		flex-direction: column;
		min-height: 0;
		min-width: 0;
	}
	.barra {
		display: flex;
		align-items: center;
		gap: 4px;
		height: 56px;
		flex: none;
		padding: 0 12px;
		border-bottom: 1px solid var(--ls);
	}
	.volver {
		display: none;
		width: 42px;
		height: 42px;
		align-items: center;
		justify-content: center;
		font-size: 22px;
	}
	.acc {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		height: 42px;
		padding: 0 12px;
		border-radius: 10px;
		font-size: 16px;
	}
	.acc:hover,
	.mas:hover {
		background: var(--sel);
	}
	.acc i {
		font-size: 19px;
		color: var(--mut);
	}
	.mas-w {
		position: relative;
		margin-left: auto;
	}
	.mas {
		width: 42px;
		height: 42px;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		border-radius: 10px;
		color: var(--mut);
		font-size: 20px;
	}
	.menu {
		position: absolute;
		right: 0;
		top: 48px;
		z-index: 5;
		width: 280px;
		padding: 6px 0;
		border-radius: 12px;
		background: var(--pop);
		box-shadow:
			var(--popsh),
			0 0 0 1px var(--ls);
	}
	.menu button {
		width: 100%;
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 10px 16px;
		font-size: 16px;
		text-align: left;
	}
	.menu button:hover {
		background: var(--sel);
	}
	.menu i {
		font-size: 19px;
		color: var(--mut);
	}
	.cont {
		flex: 1;
		min-height: 0;
		overflow-y: auto;
		padding: 24px 32px;
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
	h1 {
		margin: 0;
		font-size: 26px;
		font-weight: 600;
		line-height: 1.3;
		letter-spacing: -0.01em;
		text-wrap: pretty;
	}
	.hilo {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 12px 0;
		border-top: 1px solid var(--ls);
		border-bottom: 1px solid var(--ls);
		color: var(--mut);
		font-size: 16px;
	}
	.avs {
		display: flex;
	}
	.mini {
		width: 28px;
		height: 28px;
		border-radius: 50%;
		background: var(--avC);
		color: var(--avCT);
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 13px;
		font-weight: 600;
		box-shadow: 0 0 0 2px var(--card);
	}
	.mini + .mini {
		margin-left: -6px;
		background: oklch(0.68 calc(0.19 * var(--ac)) calc(var(--h) + 0));
		color: #fff;
	}
	.ht {
		flex: 1;
	}
	.ver {
		color: var(--accT);
		font-weight: 500;
	}
	.anteriores {
		display: flex;
		flex-direction: column;
		border-radius: 10px;
		background: var(--hun);
		margin-top: -8px;
	}
	.ant {
		display: grid;
		grid-template-columns: 1fr auto;
		gap: 2px 12px;
		padding: 10px 14px;
		color: var(--fg);
		text-decoration: none;
		font-size: 15px;
	}
	.ant + .ant {
		border-top: 1px solid var(--ls);
	}
	.af {
		color: var(--mut);
	}
	.av2 {
		grid-column: 1 / -1;
		color: var(--mut);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.remite {
		display: flex;
		gap: 14px;
		align-items: flex-start;
	}
	.avg {
		width: 48px;
		height: 48px;
		flex: none;
		border-radius: 50%;
		background: var(--avC);
		color: var(--avCT);
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 18px;
		font-weight: 600;
	}
	.rt {
		flex: 1;
		min-width: 0;
		line-height: 1.4;
	}
	.rn {
		font-size: 18px;
		font-weight: 600;
	}
	.rd {
		font-size: 15px;
		color: var(--mut);
		overflow-wrap: break-word;
	}
	.rf {
		font-size: 15px;
		color: var(--mut);
		white-space: nowrap;
	}
	.estrella {
		font-size: 20px;
		color: var(--mut);
		padding: 4px;
	}
	.estrella.on {
		color: var(--espT);
	}
	.ocultas {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 14px;
		border-radius: 10px;
		background: var(--hun);
		font-size: 15px;
		color: var(--mut);
	}
	.ocultas i {
		font-size: 18px;
	}
	.ocultas span {
		flex: 1;
	}
	.ocultas a {
		color: var(--accT);
		font-weight: 600;
		text-decoration: none;
	}
	/* Los correos HTML vienen pensados para fondo blanco: van en una hoja clara en los dos temas. */
	iframe {
		width: 100%;
		max-width: 760px;
		border: 0;
		display: block;
		flex: none;
		border-radius: 12px;
		background: #fff;
		box-shadow: 0 0 0 1px var(--ls);
	}
	.texto {
		font-size: 17px;
		line-height: 1.65;
		max-width: 700px;
		display: flex;
		flex-direction: column;
		gap: 12px;
		white-space: pre-wrap;
		overflow-wrap: anywhere;
	}
	.adjs {
		display: flex;
		flex-direction: column;
		gap: 10px;
		margin-top: 4px;
	}
	.adjc {
		display: flex;
		align-items: center;
		gap: 10px;
		font-size: 16px;
	}
	.adjc i {
		font-size: 18px;
		color: var(--mut);
	}
	.adjc span {
		font-size: 15px;
		color: var(--mut);
	}
	.grilla {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(min(100%, 180px), 1fr));
		gap: 10px;
	}
	.adj {
		position: relative;
		min-width: 0;
		border: 1px solid var(--ls);
		border-radius: 10px;
		overflow: hidden;
	}
	.adj:hover {
		border-color: var(--ln);
		box-shadow: 0 6px 18px -8px rgb(40 30 20 / 0.35);
	}
	.vp {
		position: relative;
		height: 96px;
		background: repeating-linear-gradient(135deg, var(--hun) 0 8px, var(--mu) 8px 16px);
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 30px;
	}
	.vp img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.hov {
		position: absolute;
		inset: 0;
		background: oklch(0.15 calc(0.01 * var(--nc)) var(--nh) / 0.55);
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 10px;
		opacity: 0;
		transition: opacity 0.15s;
	}
	.adj:hover .hov,
	.adj:focus-within .hov {
		opacity: 1;
	}
	.b-ver {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		height: 36px;
		padding: 0 14px;
		border-radius: 999px;
		background: #fff;
		color: oklch(0.15 calc(0.01 * var(--nc)) var(--nh));
		font-size: 14px;
		font-weight: 600;
		text-decoration: none;
	}
	.b-dl {
		width: 36px;
		height: 36px;
		border-radius: 50%;
		background: #fff;
		color: oklch(0.15 calc(0.01 * var(--nc)) var(--nh));
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 16px;
	}
	.pie {
		padding: 9px 10px;
		display: flex;
		align-items: center;
		gap: 8px;
		border-top: 1px solid var(--ls);
	}
	.pie > i {
		font-size: 20px;
	}
	.pt {
		flex: 1;
		min-width: 0;
		line-height: 1.3;
	}
	.pn {
		font-size: 14px;
		font-weight: 500;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.pm {
		font-size: 13px;
		color: var(--mut);
	}
	.botones {
		display: flex;
		gap: 12px;
		margin-top: 8px;
		flex-wrap: wrap;
	}
	.b {
		display: inline-flex;
		align-items: center;
		gap: 10px;
		height: 48px;
		padding: 0 22px;
		border: 1px solid var(--ln);
		border-radius: 999px;
		font-size: 17px;
		font-weight: 500;
	}
	.b i {
		font-size: 20px;
	}
	@media (max-width: 1100px) {
		.acc span {
			display: none;
		}
	}
	@media (max-width: 900px) {
		.volver {
			display: inline-flex;
		}
		.cont {
			padding: 18px 16px calc(24px + env(safe-area-inset-bottom));
		}
		h1 {
			font-size: 22px;
		}
	}
</style>
