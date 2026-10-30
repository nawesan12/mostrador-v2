<script lang="ts">
	import { onMount } from 'svelte';
	import type { Mensaje } from '#lib/server/bandeja.ts';
	import { tamano, tipoAdjunto } from '#lib/formato.ts';

	/**
	 * Imprimir (handoff 2m): la hoja con membrete a la izquierda, las opciones a la derecha. La hoja
	 * es un documento aparte (iframe sin scripts); «Imprimir» abre el diálogo del sistema sobre ella,
	 * que también deja guardarla como PDF y elegir impresora y copias.
	 */
	let { m, marca, oncerrar }: { m: Mensaje; marca: string; oncerrar: () => void } = $props();

	let todo = $state(false);
	const fotos = $derived(m.adjuntos.filter((a) => /^image\/(png|jpeg|gif|webp)$/.test(a.tipo)));
	// Las fotos van marcadas: se imprimen a continuación, una por hoja.
	let conFotos = $state<Record<string, boolean>>({});
	let hoja = $state<HTMLIFrameElement>();

	const esc = (s: string) =>
		s.replace(/[&<>"]/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;' })[c]!);
	const dir = (d: { name: string | null; email: string }) =>
		d.name ? `${esc(d.name)} &lt;${esc(d.email)}&gt;` : esc(d.email);
	const largo = (iso: string) =>
		new Date(iso).toLocaleString('es-AR', {
			weekday: 'long',
			day: 'numeric',
			month: 'long',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit',
			timeZone: 'America/Argentina/Buenos_Aires'
		});
	const urlAdj = (a: Mensaje['adjuntos'][number]) =>
		`/adjunto/${encodeURIComponent(m.cuenta)}/${encodeURIComponent(a.blob)}?n=${encodeURIComponent(a.nombre)}&t=${encodeURIComponent(a.tipo)}&ver=1`;
	const marcada = (a: Mensaje['adjuntos'][number]) => conFotos[a.blob + a.parte] ?? true;

	const doc = $derived.by(() => {
		const hoy = new Date().toLocaleDateString('es-AR', {
			timeZone: 'America/Argentina/Buenos_Aires'
		});
		const cuerpo = esc(m.texto.replace(/\r/g, '').trim());
		const anteriores = todo
			? m.anteriores
					.map(
						(a) =>
							`<div class="ant"><div class="meta">${dir(a.de)} · ${esc(largo(a.fecha))}</div><div>${esc(a.vista)}</div></div>`
					)
					.join('')
			: '';
		const otros = m.adjuntos.map(
			(a) => esc(a.nombre) + (fotos.includes(a) && marcada(a) ? ' (se imprime a continuación)' : '')
		);
		const paginas = fotos
			.filter(marcada)
			.map(
				(a) =>
					`<section class="foto"><div class="meta">${esc(a.nombre)}</div><img src="${urlAdj(a)}" alt=""></section>`
			)
			.join('');
		return `<!doctype html><html><head><meta charset="utf-8"><style>
@page{size:A4;margin:16mm}
html,body{margin:0;background:#fff;color:#1a1a1a;font:13px/1.5 Geist,ui-sans-serif,system-ui,sans-serif}
@media screen{body{padding:34px 36px}}
.mem{display:flex;align-items:center;gap:10px;padding-bottom:12px;border-bottom:2px solid #1a1a1a}
.logo{width:30px;height:30px;border-radius:6px;background:#2d5fe2;-webkit-print-color-adjust:exact;print-color-adjust:exact;display:flex;align-items:center;justify-content:center}
.logo img{width:19px;height:19px;filter:brightness(0) invert(1)}
.mm{flex:1;line-height:1.25}.mn{font-weight:700;font-size:14px}.gris{color:#666;font-size:12px}
h1{font-size:16px;line-height:1.3;margin:14px 0 6px}
.meta{color:#555;font-size:12px;line-height:1.5}
.texto{white-space:pre-wrap;overflow-wrap:anywhere;margin-top:12px}
.ant{margin-top:14px;padding-top:10px;border-top:1px solid #ddd}
.adj{margin-top:18px;padding-top:10px;border-top:1px solid #ddd;color:#555;font-size:12px}
.foto{break-before:page;display:flex;flex-direction:column;gap:8px}
.foto img{max-width:100%;max-height:240mm;object-fit:contain}
</style></head><body>
<div class="mem"><div class="logo"><img src="/vimo.svg" alt=""></div><div class="mm"><div class="mn">${esc(marca)}</div><div class="gris">${esc(m.casilla)}</div></div><div class="gris">${hoy}</div></div>
<h1>${esc(m.asunto || '(sin asunto)')}</h1>
<div class="meta">De: ${dir(m.de)}<br>Para: ${m.para.map(dir).join(', ')}${m.cc.length ? `<br>CC: ${m.cc.map(dir).join(', ')}` : ''}<br>Fecha: ${esc(largo(m.fecha))}</div>
<div class="texto">${cuerpo}</div>
${anteriores}
${otros.length ? `<div class="adj">Adjuntos: ${otros.join(' · ')}</div>` : ''}
${paginas}
</body></html>`;
	});

	function imprimir() {
		hoja?.contentWindow?.print();
	}
	onMount(() => {
		const tecla = (e: KeyboardEvent) => e.key === 'Escape' && oncerrar();
		addEventListener('keydown', tecla);
		return () => removeEventListener('keydown', tecla);
	});
</script>

<div class="fondo" role="presentation" onclick={oncerrar}></div>
<div class="dialogo" role="dialog" aria-modal="true" aria-label="Imprimir">
	<div class="previa">
		<!-- Sin scripts: sólo se deja abrir el diálogo de impresión. -->
		<iframe
			bind:this={hoja}
			title="Vista previa"
			sandbox="allow-same-origin allow-modals"
			srcdoc={doc}
		></iframe>
	</div>
	<div class="opciones">
		<div class="tit">
			<span>Imprimir</span>
			<button aria-label="Cerrar" onclick={oncerrar}><i class="icon-x"></i></button>
		</div>
		{#if m.anteriores.length}
			<div class="grupo">
				<div class="et">Qué imprimir</div>
				<div class="dos">
					<button class="op" class:act={!todo} onclick={() => (todo = false)}
						><span class="on">Solo este correo</span></button
					>
					<button class="op" class:act={todo} onclick={() => (todo = true)}
						><span class="on">Toda la conversación</span><span class="od"
							>{m.anteriores.length + 1} correos</span
						></button
					>
				</div>
			</div>
		{/if}
		{#if fotos.length}
			<div class="grupo">
				<div class="et">Imprimir también las fotos</div>
				{#each fotos as a (a.blob + a.parte)}
					{@const t = tipoAdjunto(a.tipo, a.nombre)}
					<label class="fila">
						<input
							type="checkbox"
							checked={marcada(a)}
							onchange={(e) => (conFotos[a.blob + a.parte] = e.currentTarget.checked)}
						/>
						<i class={t.ic} style:color={t.c}></i>
						<span class="fn"><span>{a.nombre}</span><span class="od">{tamano(a.tamano)}</span></span
						>
					</label>
				{/each}
			</div>
		{/if}
		<p class="nota">
			La impresora, las copias y «Guardar como PDF» se eligen en el paso siguiente.
		</p>
		<div class="pie">
			<button class="cancelar" onclick={oncerrar}>Cancelar</button>
			<button class="primario" onclick={imprimir}><i class="icon-printer"></i>Imprimir</button>
		</div>
	</div>
</div>

<style>
	.fondo {
		position: fixed;
		inset: 0;
		z-index: 50;
		background: oklch(0.12 calc(0.01 * var(--nc)) var(--nh) / 0.7);
		backdrop-filter: blur(6px);
		-webkit-backdrop-filter: blur(6px);
	}
	.dialogo {
		position: fixed;
		left: 50%;
		top: 50%;
		transform: translate(-50%, -50%);
		z-index: 51;
		width: min(1080px, calc(100vw - 32px));
		height: min(760px, calc(100dvh - 32px));
		display: grid;
		grid-template-columns: 540px 1fr;
		/* El tono de los menús, sobre el fondo sólido: no se transparenta el correo de atrás. */
		background: linear-gradient(var(--pop), var(--pop)), var(--bg);
		border-radius: 18px;
		overflow: hidden;
		box-shadow: var(--popsh);
		color: var(--fg);
	}
	.previa {
		background: var(--mu);
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 24px;
		min-height: 0;
	}
	iframe {
		width: 440px;
		height: 100%;
		max-height: 640px;
		border: 0;
		background: #fff;
		box-shadow: 0 10px 30px -10px rgb(0 0 0 / 0.35);
	}
	.opciones {
		padding: 30px 32px;
		display: flex;
		flex-direction: column;
		gap: 22px;
		min-height: 0;
		overflow-y: auto;
	}
	.tit {
		display: flex;
		align-items: center;
	}
	.tit span {
		flex: 1;
		font-size: 24px;
		font-weight: 600;
	}
	.tit button {
		font-size: 22px;
		color: var(--mut);
		padding: 6px;
	}
	.grupo {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.et {
		font-size: 15px;
		color: var(--mut);
	}
	.dos {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 10px;
	}
	.op {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		padding: 12px 14px;
		border-radius: 12px;
		box-shadow: inset 0 0 0 1px var(--ln);
		text-align: left;
	}
	.op.act {
		box-shadow: 0 0 0 2px var(--acc);
		background: var(--sel);
	}
	.on {
		font-size: 16px;
		font-weight: 600;
	}
	.od {
		font-size: 14px;
		color: var(--mut);
	}
	.fila {
		display: flex;
		align-items: center;
		gap: 12px;
		min-height: 46px;
		cursor: pointer;
	}
	.fila input {
		width: 20px;
		height: 20px;
		accent-color: var(--acc);
	}
	.fila i {
		font-size: 20px;
	}
	.fn {
		display: flex;
		flex-direction: column;
		line-height: 1.3;
		font-size: 16px;
		min-width: 0;
	}
	.nota {
		margin: 0;
		font-size: 14px;
		color: var(--mut);
	}
	.pie {
		margin-top: auto;
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.cancelar {
		font-size: 16px;
		color: var(--mut);
		font-weight: 500;
		padding: 0 6px;
	}
	.primario {
		margin-left: auto;
		display: inline-flex;
		align-items: center;
		gap: 8px;
		height: 50px;
		padding: 0 26px;
		border-radius: 999px;
		background: var(--acc);
		color: #fff;
		font-size: 16px;
		font-weight: 600;
	}
	@media (max-width: 900px) {
		.dialogo {
			grid-template-columns: 1fr;
			grid-template-rows: minmax(0, 1fr) auto;
		}
		.previa {
			padding: 16px;
		}
		iframe {
			width: 100%;
		}
		.opciones {
			padding: 20px;
			gap: 16px;
		}
	}
</style>
