<script lang="ts">
	import { onMount } from 'svelte';
	import type { Mensaje } from '#lib/server/bandeja.ts';
	import { cuando, tamano, tipoAdjunto } from '#lib/formato.ts';

	/**
	 * Visor de adjuntos (handoff 2h foto, 2i PDF): oscuro siempre, sobre todo. Las fotos con zoom y
	 * giro; los PDF se leen sin descargarlos; lo demás (planillas, documentos) se descarga.
	 */
	let { m, inicio, oncerrar }: { m: Mensaje; inicio: number; oncerrar: () => void } = $props();

	// Escribible: arranca en el que se tocó y se mueve con las flechas.
	let i = $derived(inicio);
	let zoom = $state(1);
	let giro = $state(0);
	const a = $derived(m.adjuntos[i]);
	const t = $derived(tipoAdjunto(a.tipo, a.nombre));
	const foto = $derived(/^image\/(png|jpeg|gif|webp|avif)$/.test(a.tipo));
	const pdf = $derived(a.tipo === 'application/pdf' || a.nombre.toLowerCase().endsWith('.pdf'));
	const url = (ver: boolean) =>
		`/adjunto/${encodeURIComponent(m.cuenta)}/${encodeURIComponent(a.blob)}?n=${encodeURIComponent(a.nombre)}&t=${encodeURIComponent(pdf ? 'application/pdf' : a.tipo)}${ver ? '&ver=1' : ''}`;
	const de = $derived(m.de.name || m.de.email);

	function ir(d: number) {
		i = (i + d + m.adjuntos.length) % m.adjuntos.length;
		zoom = 1;
		giro = 0;
	}
	function imprimir() {
		const w = window.open(url(true), '_blank', 'noopener');
		if (!w) return;
	}
	onMount(() => {
		const tecla = (e: KeyboardEvent) => {
			if (e.key === 'Escape') oncerrar();
			else if (e.key === 'ArrowRight') ir(1);
			else if (e.key === 'ArrowLeft') ir(-1);
		};
		addEventListener('keydown', tecla);
		document.body.style.overflow = 'hidden';
		return () => {
			removeEventListener('keydown', tecla);
			document.body.style.overflow = '';
		};
	});
</script>

<div class="visor" role="dialog" aria-modal="true" aria-label="Adjunto {a.nombre}">
	<div class="cab">
		<button class="redondo" aria-label="Cerrar" onclick={oncerrar}><i class="icon-x"></i></button>
		<i class="{t.ic} ti"></i>
		<div class="nombre">
			<div class="n">{a.nombre}</div>
			<div class="d">{tamano(a.tamano)} · de {de} · {cuando(m.fecha).toLowerCase()}</div>
		</div>
		<div class="acciones">
			{#if foto}
				<span class="zoom">
					<button aria-label="Alejar" onclick={() => (zoom = Math.max(0.25, zoom - 0.25))}
						><i class="icon-zoom-out"></i></button
					>
					<span class="pct">{Math.round(zoom * 100)}%</span>
					<button aria-label="Acercar" onclick={() => (zoom = Math.min(4, zoom + 0.25))}
						><i class="icon-zoom-in"></i></button
					>
				</span>
				<button class="redondo chico" aria-label="Girar" onclick={() => (giro = (giro + 90) % 360)}
					><i class="icon-rotate-cw"></i></button
				>
			{/if}
			{#if foto || pdf}
				<button class="pill" onclick={imprimir}
					><i class="icon-printer"></i><span>Imprimir</span></button
				>
			{/if}
			<a class="pill claro" href={url(false)} download={a.nombre}
				><i class="icon-download"></i><span>Descargar</span></a
			>
		</div>
	</div>

	<div class="cuerpo">
		{#if foto}
			<img src={url(true)} alt={a.nombre} style:transform="scale({zoom}) rotate({giro}deg)" />
		{:else if pdf}
			<iframe src={url(true)} title={a.nombre}></iframe>
		{:else}
			<div class="otro">
				<i class={t.ic} style:color={t.c}></i>
				<p>
					Este archivo ({t.l.toLowerCase()}) no se puede ver acá. Descargalo para abrirlo con su
					programa.
				</p>
				<a class="pill claro" href={url(false)} download={a.nombre}
					><i class="icon-download"></i>Descargar</a
				>
			</div>
		{/if}
		{#if m.adjuntos.length > 1}
			<button class="flecha izq" aria-label="Anterior" onclick={() => ir(-1)}
				><i class="icon-chevron-left"></i></button
			>
			<button class="flecha der" aria-label="Siguiente" onclick={() => ir(1)}
				><i class="icon-chevron-right"></i></button
			>
		{/if}
	</div>

	{#if m.adjuntos.length > 1}
		<div class="tira">
			{#each m.adjuntos as x, j (x.blob + x.parte)}
				{@const tx = tipoAdjunto(x.tipo, x.nombre)}
				<button
					class="miniatura"
					class:act={j === i}
					onclick={() => ((i = j), (zoom = 1), (giro = 0))}
				>
					<span class="mini"><i class={tx.ic}></i></span>
					<span class="mn">{x.nombre}</span>
				</button>
			{/each}
		</div>
		<div class="ayuda">← → para pasar · Esc para cerrar</div>
	{/if}
</div>

<style>
	.visor {
		position: fixed;
		inset: 0;
		z-index: 60;
		display: flex;
		flex-direction: column;
		background: oklch(0.14 calc(0.008 * var(--nc)) var(--nh));
		color: #fff;
		padding: env(safe-area-inset-top) env(safe-area-inset-right) env(safe-area-inset-bottom)
			env(safe-area-inset-left);
	}
	.cab {
		min-height: 68px;
		flex: none;
		display: flex;
		align-items: center;
		flex-wrap: wrap;
		gap: 14px;
		padding: 8px 16px 8px 12px;
	}
	.redondo {
		width: 46px;
		height: 46px;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		background: oklch(1 0 0 / 0.1);
		color: #fff;
		font-size: 22px;
		flex: none;
	}
	.redondo.chico {
		width: 44px;
		height: 44px;
		font-size: 18px;
	}
	.ti {
		font-size: 24px;
		color: oklch(0.78 0.11 250);
	}
	.nombre {
		line-height: 1.3;
		min-width: 0;
		flex: 1;
	}
	.n {
		font-size: 17px;
		font-weight: 600;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.d {
		font-size: 14px;
		color: oklch(1 0 0 / 0.65);
	}
	.acciones {
		margin-left: auto;
		display: flex;
		align-items: center;
		gap: 8px;
	}
	.zoom {
		display: inline-flex;
		align-items: center;
		gap: 2px;
		height: 44px;
		padding: 0 6px;
		border-radius: 999px;
		background: oklch(1 0 0 / 0.1);
	}
	.zoom button {
		color: #fff;
		padding: 8px;
		font-size: 18px;
	}
	.pct {
		font-family: 'Geist Mono', monospace;
		font-size: 14px;
		width: 46px;
		text-align: center;
	}
	.pill {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		height: 44px;
		padding: 0 16px;
		border-radius: 999px;
		background: oklch(1 0 0 / 0.1);
		color: #fff;
		font-size: 15px;
		font-weight: 500;
		text-decoration: none;
	}
	.pill.claro {
		background: oklch(0.97 calc(0.005 * var(--nc)) var(--nh));
		color: oklch(0.15 calc(0.01 * var(--nc)) var(--nh));
		font-weight: 600;
	}
	.cuerpo {
		position: relative;
		flex: 1;
		min-height: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden;
		padding: 12px 88px;
	}
	img {
		max-width: 100%;
		max-height: 100%;
		object-fit: contain;
		border-radius: 4px;
		box-shadow: 0 30px 80px -20px rgb(0 0 0 / 0.7);
		transition: transform 0.15s ease;
	}
	iframe {
		width: min(100%, 980px);
		height: 100%;
		border: 0;
		border-radius: 6px;
		background: #fff;
	}
	.otro {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 14px;
		max-width: 420px;
		text-align: center;
		color: oklch(1 0 0 / 0.75);
	}
	.otro i {
		font-size: 48px;
	}
	.flecha {
		position: absolute;
		top: 50%;
		width: 56px;
		height: 56px;
		margin-top: -28px;
		border-radius: 50%;
		background: oklch(1 0 0 / 0.12);
		color: #fff;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 26px;
	}
	.izq {
		left: 20px;
	}
	.der {
		right: 20px;
	}
	.tira {
		height: 112px;
		flex: none;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 12px;
		overflow-x: auto;
		padding: 0 16px;
	}
	.miniatura {
		width: 120px;
		flex: none;
		display: flex;
		flex-direction: column;
		gap: 6px;
		align-items: center;
		opacity: 0.55;
		color: #fff;
	}
	.miniatura.act {
		opacity: 1;
	}
	.mini {
		width: 120px;
		height: 68px;
		border-radius: 8px;
		background: repeating-linear-gradient(
			135deg,
			oklch(0.32 calc(0.01 * var(--nc)) var(--nh)) 0 6px,
			oklch(0.28 calc(0.01 * var(--nc)) var(--nh)) 6px 12px
		);
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 22px;
	}
	.miniatura.act .mini {
		box-shadow: 0 0 0 3px oklch(0.68 calc(0.19 * var(--ac)) calc(var(--h) + 0));
	}
	.mn {
		font-size: 13px;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		max-width: 120px;
	}
	.ayuda {
		position: absolute;
		left: 50%;
		bottom: calc(122px + env(safe-area-inset-bottom));
		transform: translateX(-50%);
		font-size: 14px;
		color: oklch(1 0 0 / 0.6);
	}
	@media (max-width: 760px) {
		.cuerpo {
			padding: 8px;
		}
		.flecha {
			width: 44px;
			height: 44px;
			font-size: 22px;
		}
		.izq {
			left: 8px;
		}
		.der {
			right: 8px;
		}
		.pill span,
		.ayuda {
			display: none;
		}
		.acciones {
			width: 100%;
			justify-content: flex-end;
		}
	}
</style>
