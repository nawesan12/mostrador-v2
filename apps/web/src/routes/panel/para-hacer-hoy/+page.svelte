<script lang="ts">
	import { goto, invalidateAll } from '$app/navigation';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { pesos } from '#lib/panel/formato.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/** Para hacer hoy (design/PanelCrecer.dc.html, sTodo): tareas ordenadas por la plata que mueven. */
	let { data } = $props();
	const ui = getUi();
	type T = (typeof data.grupos)[number]['tareas'][number];

	const COLOR = { Urgente: '#B83A2E', Hoy: '#9A6A10', 'Esta semana': '#2F8F63' } as const;
	const ESTILO: Record<T['tipo'], [I.IconData, string, string, string]> = {
		preparar: [
			I.phPackageFill,
			'rgba(210,88,47,.13)',
			'#C24E25',
			'/panel/pedidos?estado=para-preparar'
		],
		chats: [I.phChatCircleDotsFill, 'rgba(47,143,99,.13)', '#2F8F63', '/panel/whatsapp'],
		carritos: [
			I.phShoppingCartSimpleFill,
			'rgba(227,169,59,.2)',
			'#9A6A10',
			'/panel/carritos-abandonados'
		],
		transferencias: [
			I.phBankFill,
			'var(--chip)',
			'var(--mu)',
			'/panel/pedidos?estado=sin-pagar&pago=transf'
		],
		certificado: [I.phCertificateFill, '#E6ECF7', '#1F3A6E', '/panel/facturacion'],
		reponer: [I.phPackageFill, 'var(--chip)', 'var(--mu)', '/panel/inventario?seg=out,crit,low']
	};
	const todas = $derived(data.grupos.flatMap((g) => g.tareas));
	const hechas = $derived(todas.filter((t) => t.hecha).length);
	const titulo = $derived(
		!todas.length || hechas === todas.length
			? '¡Día terminado! 🎉'
			: `Te ${todas.length - hechas === 1 ? 'queda 1 cosa' : `quedan ${todas.length - hechas} cosas`} para hoy`
	);

	async function marcar(t: T, hecha: boolean) {
		await fetch(`/panel/para-hacer-hoy/${t.id}`, {
			method: 'PUT',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ hecha })
		});
		await invalidateAll();
	}
	async function hacer(t: T) {
		// Recuperar carritos manda los recordatorios de verdad; el resto lleva a la pantalla justa.
		if (t.tipo === 'carritos') {
			const r = await fetch('/panel/carritos-abandonados/todos', {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: '{}'
			});
			if (!r.ok) return ui.toast({ k: 'err', title: 'No se mandaron', desc: 'Probá de nuevo.' });
		}
		await marcar(t, true);
		ui.toast({
			k: 'ok',
			title: t.titulo,
			desc: 'Listo, lo marcamos como hecho',
			action: 'Deshacer',
			onAction: () => marcar(t, false)
		});
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- rutas del panel con filtros
		if (t.tipo !== 'carritos') await goto(ESTILO[t.tipo][3]);
	}
	const C = 2 * Math.PI * 52;
</script>

<Seo
	title="Para hacer hoy · Panel"
	description="Tus tareas de hoy"
	path="/panel/para-hacer-hoy"
	noindex
/>

<div class="pagina">
	<div class="banda" use:anim={{ t: 'rise' }}>
		<div class="arcos"></div>
		<div class="anillo">
			<svg viewBox="0 0 120 120" aria-hidden="true"
				><circle cx="60" cy="60" r="52" class="fondo-a"></circle><circle
					cx="60"
					cy="60"
					r="52"
					class="lleno"
					stroke-dasharray={C}
					stroke-dashoffset={C * (1 - (todas.length ? hechas / todas.length : 1))}
				></circle></svg
			>
			<div class="cuenta">{hechas}/{todas.length}</div>
		</div>
		<div class="b-txt">
			<h1>{titulo}</h1>
			<div class="b-s">Hornero ordenó tus tareas por lo que más plata te mueve.</div>
		</div>
		<img src="/hornero/portapapeles.png" alt="" width="96" height="96" />
	</div>

	{#each data.grupos as g (g.titulo)}
		<div class="grupo" use:anim={{ t: 'rise', delay: 60 }}>
			<div class="g-t" style:color={COLOR[g.titulo]}>
				<span class="punto" style:background={COLOR[g.titulo]}></span>{g.titulo}
			</div>
			{#each g.tareas as t (t.id)}
				{@const [icono, bg, fg] = ESTILO[t.tipo]}
				<div class="tarea" class:hecha={t.hecha}>
					<button
						type="button"
						class="check"
						role="checkbox"
						aria-checked={t.hecha}
						aria-label="Marcar {t.titulo}"
						onclick={() => marcar(t, !t.hecha)}
						>{#if t.hecha}<Icon icon={I.phCheckBold} size={14} />{/if}</button
					>
					<span class="ico" style:background={bg} style:color={fg}
						><Icon icon={icono} size={19} /></span
					>
					<span class="t-txt"
						><span class="t-t">{t.titulo}</span><span class="t-d">{t.detalle}</span></span
					>
					{#if t.valor}<span class="valor">{t.aproximado ? '~' : ''}{pesos(t.valor)}</span>{/if}
					<button type="button" class="hacer" onclick={() => hacer(t)}>{t.accion}</button>
				</div>
			{/each}
		</div>
	{/each}
</div>

<style>
	.pagina {
		display: flex;
		flex-direction: column;
		gap: 18px;
	}
	.banda {
		display: flex;
		align-items: center;
		gap: 18px;
		flex-wrap: wrap;
		padding: 20px;
		border-radius: 24px;
		background: #2a140c;
		color: #fbf7f2;
		position: relative;
		overflow: hidden;
	}
	.banda > :not(.arcos) {
		position: relative;
	}
	.arcos {
		position: absolute;
		inset: 0;
		background-image: url('/textura/arch-pattern-light.svg');
	}
	.anillo {
		width: 96px;
		height: 96px;
		flex: none;
	}
	.anillo svg {
		width: 100%;
		height: 100%;
		transform: rotate(-90deg);
	}
	.anillo circle {
		fill: none;
		stroke-width: 12;
	}
	.fondo-a {
		stroke: rgba(251, 247, 242, 0.12);
	}
	.lleno {
		stroke: #e8875f;
		stroke-linecap: round;
		transition: stroke-dashoffset 0.6s cubic-bezier(0.2, 0.7, 0.1, 1);
	}
	.cuenta {
		position: absolute;
		inset: 0;
		display: grid;
		place-items: center;
		font-size: 22px;
		font-weight: 600;
	}
	.b-txt {
		flex: 1 1 240px;
	}
	h1 {
		margin: 0;
		font-size: 28px;
		letter-spacing: -0.035em;
		font-weight: 600;
	}
	.b-s {
		margin-top: 4px;
		font-size: 14.5px;
		color: #d9c8ba;
	}
	.banda img {
		object-fit: contain;
	}
	.grupo {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.g-t {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 13px;
		font-weight: 600;
	}
	.punto {
		width: 8px;
		height: 8px;
		border-radius: 50%;
	}
	.tarea {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 14px;
		border-radius: 20px;
		background: var(--sf);
		box-shadow: var(--shd);
		transition: opacity 0.3s;
		flex-wrap: wrap;
	}
	.tarea.hecha {
		opacity: 0.55;
	}
	.check {
		width: 26px;
		height: 26px;
		flex: none;
		border-radius: 8px;
		border: 2px solid var(--ln2);
		background: transparent;
		color: #fff;
		display: grid;
		place-items: center;
		padding: 0;
	}
	.hecha .check {
		border-color: #2f8f63;
		background: #2f8f63;
	}
	.ico {
		width: 38px;
		height: 38px;
		flex: none;
		border-radius: 12px;
		display: grid;
		place-items: center;
	}
	.t-txt {
		flex: 1 1 180px;
		min-width: 0;
	}
	.t-t {
		display: block;
		font-size: 14.5px;
		font-weight: 600;
	}
	.hecha .t-t {
		text-decoration: line-through;
	}
	.t-d {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
	}
	.valor {
		font-size: 13px;
		font-weight: 600;
		color: #2f8f63;
		white-space: nowrap;
	}
	.hacer {
		padding: 8px 14px;
		border-radius: 999px;
		border: 0;
		background: var(--tx);
		color: var(--sf);
		font:
			600 13px 'Instrument Sans',
			sans-serif;
		white-space: nowrap;
	}
</style>
