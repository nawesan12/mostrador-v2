<script lang="ts">
	import { SvelteSet } from 'svelte/reactivity';
	import MarcaIcono from '#lib/components/MarcaIcono.svelte';
	import type { components } from '@mostrador/contrato/esquema';
	import { rutaDe } from './navegacion.ts';
	import { getShell, hace } from './shell.svelte.ts';
	import { esRutaPanel } from './notificaciones.ts';

	type Notificacion = components['schemas']['Notificacion'];
	type Pagina = components['schemas']['PaginaNotificaciones'];

	/**
	 * La campanita (handoff; §45): Todo / Ventas / Sistema, «Marcar todo como leído» y, al tocar una,
	 * se marca leída y abre su página del panel.
	 */
	let {
		pagina,
		marcarLeidas,
		tocar
	}: {
		pagina: Pagina;
		marcarLeidas: () => void;
		/** Marca leída una (la recarga la hace quien llama). */
		tocar: (n: Notificacion) => void;
	} = $props();
	const shell = getShell();
	let grupo = $state<'todo' | 'ventas' | 'sistema'>('todo');
	// «Ver anteriores»: las páginas que se suman abajo. Si llega una lista nueva (tiempo real, marcar
	// todo), se vuelve a la primera.
	let anteriores = $state<Notificacion[]>([]);
	let siguiente = $state<string | null>(null);
	let cargando = $state(false);
	let falla = $state(false);
	$effect.pre(() => {
		siguiente = pagina.siguiente;
		anteriores = [];
		falla = false;
	});
	// Lo que se tocó se ve leído en el acto, antes de que vuelva la lista.
	const tocadas = new SvelteSet<string>();
	const todas = $derived([...pagina.notificaciones, ...anteriores]);
	const lista = $derived(todas.filter((n) => grupo === 'todo' || n.grupo === grupo));
	const leida = (n: Notificacion) => n.leida || tocadas.has(n.id);

	async function verAnteriores() {
		if (!siguiente || cargando) return;
		cargando = true;
		falla = false;
		try {
			const r = await fetch(`/panel/notificaciones/lista?cursor=${encodeURIComponent(siguiente)}`);
			if (!r.ok) throw new Error(String(r.status));
			const p = (await r.json()) as Pagina;
			anteriores = [...anteriores, ...p.notificaciones];
			siguiente = p.siguiente;
		} catch {
			falla = true;
		} finally {
			cargando = false;
		}
	}

	function abrir(n: Notificacion) {
		if (!leida(n)) {
			tocadas.add(n.id);
			tocar(n);
		}
		shell.notificaciones = false;
	}

	function entra(el: HTMLElement) {
		el.animate(
			[
				{ opacity: 0, transform: 'scale(.96) translateY(-6px)' },
				{ opacity: 1, transform: 'none' }
			],
			{ duration: 260, easing: 'cubic-bezier(.2,.7,.1,1)' }
		);
	}
</script>

{#snippet contenido(n: Notificacion)}
	<MarcaIcono k={n.tipo} size={36} radius={11} />
	<span class="n-txt">
		<span class="n-t">{n.titulo}</span>
		{#if n.detalle}<span class="n-d">{n.detalle}</span>{/if}
		<span class="n-h">{hace(n.creada)}</span>
	</span>
	{#if !leida(n)}<span class="sin-leer" aria-label="Sin leer"></span>{/if}
{/snippet}

{#if shell.notificaciones}
	<div class="fuera" onclick={() => (shell.notificaciones = false)} aria-hidden="true"></div>
	<div class="caja" role="dialog" aria-label="Notificaciones" use:entra>
		<div class="cab">
			<span class="t">Notificaciones</span>
			{#if pagina.sinLeer > 0 || todas.some((n) => !leida(n))}
				<button type="button" class="leer" onclick={marcarLeidas}>Marcar todo como leído</button>
			{/if}
		</div>
		<div class="tabs" role="group" aria-label="Filtrar">
			{#each [['todo', 'Todo'], ['ventas', 'Ventas'], ['sistema', 'Sistema']] as const as [k, t] (k)}
				<button type="button" aria-pressed={grupo === k} onclick={() => (grupo = k)}>{t}</button>
			{/each}
		</div>
		<div class="lista">
			{#each lista as n (n.id)}
				{#if n.ruta && esRutaPanel(n.ruta)}
					<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- la ruta del panel la arma la API (§45) -->
					<a class="n" class:leida={leida(n)} href={n.ruta} onclick={() => abrir(n)}
						>{@render contenido(n)}</a
					>
				{:else}
					<button type="button" class="n" class:leida={leida(n)} onclick={() => abrir(n)}
						>{@render contenido(n)}</button
					>
				{/if}
			{:else}
				<div class="vacio">
					{#if todas.length === 0}
						No tenés notificaciones por ahora. Acá te avisamos de ventas, pagos y lo que necesite tu
						atención.
					{:else}
						Nada en {grupo === 'ventas' ? 'Ventas' : 'Sistema'} por ahora.
					{/if}
				</div>
			{/each}
			{#if siguiente && grupo === 'todo'}
				<div class="mas">
					<button
						type="button"
						class="btn btn-soft btn-sm"
						disabled={cargando}
						onclick={verAnteriores}>{cargando ? 'Cargando…' : 'Ver anteriores'}</button
					>
					{#if falla}<span class="falla" role="status"
							>No pudimos traer las anteriores. Probá de nuevo.</span
						>{/if}
				</div>
			{/if}
		</div>
		<div class="pie">
			<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- ruta del panel armada con rutaDe -->
			<a href={rutaDe('Notificaciones')} onclick={() => (shell.notificaciones = false)}
				>Configurar qué me avisa Tu Local Web</a
			>
		</div>
	</div>
{/if}

<style>
	.fuera {
		position: fixed;
		inset: 0;
		z-index: 65;
	}
	.caja {
		position: fixed;
		top: 58px;
		right: 12px;
		z-index: 66;
		width: min(400px, calc(100vw - 24px));
		border-radius: 22px;
		background: var(--sf);
		color: var(--tx);
		box-shadow:
			0 30px 70px -20px rgba(42, 20, 12, 0.5),
			0 0 0 1px var(--ln);
		overflow: hidden;
		transform-origin: top right;
	}
	.cab {
		display: flex;
		align-items: center;
		justify-content: space-between;
		min-height: 52px;
		padding: 16px 18px 10px;
	}
	.t {
		font-size: 16px;
		font-weight: 600;
	}
	.leer {
		border: 0;
		background: transparent;
		font-size: 13px;
		font-weight: 600;
		color: #b0441d;
		padding: 0;
	}
	.tabs {
		display: flex;
		gap: 6px;
		padding: 0 18px 12px;
	}
	.tabs button {
		font-size: 13px;
		font-weight: 500;
		padding: 6px 12px;
		border-radius: 999px;
		border: 0;
		background: var(--chip);
		color: var(--mu);
	}
	.tabs button[aria-pressed='true'] {
		font-weight: 600;
		background: var(--tx);
		color: var(--sf);
	}
	.lista {
		max-height: min(400px, calc(100dvh - 220px));
		overflow: auto;
		border-top: 1px solid var(--ln);
	}
	.n {
		display: flex;
		gap: 12px;
		width: 100%;
		padding: 12px 18px;
		border: 0;
		border-bottom: 1px solid var(--ln);
		background: transparent;
		color: inherit;
		font: inherit;
		text-align: left;
		text-decoration: none;
	}
	.n:hover {
		background: var(--chip);
	}
	.n-txt {
		flex: 1;
		min-width: 0;
	}
	.n-t {
		display: block;
		font-size: 14px;
		font-weight: 600;
	}
	.leida .n-t {
		font-weight: 500;
	}
	.n-d {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
		line-height: 1.4;
		margin-top: 1px;
	}
	.n-h {
		display: block;
		font-size: 11.5px;
		color: var(--fa);
		margin-top: 3px;
	}
	.sin-leer {
		width: 8px;
		height: 8px;
		margin-top: 6px;
		border-radius: 50%;
		background: #d2582f;
		flex: none;
	}
	.vacio {
		padding: 28px 18px;
		text-align: center;
		font-size: 14px;
		line-height: 1.45;
		color: var(--mu);
		text-wrap: balance;
	}
	.mas {
		display: grid;
		justify-items: center;
		gap: 6px;
		padding: 12px 18px;
	}
	.falla {
		font-size: 12.5px;
		color: var(--mu);
		text-align: center;
	}
	.pie {
		padding: 12px 18px;
		text-align: center;
	}
	.pie a {
		font-size: 13px;
		font-weight: 600;
		color: #b0441d;
	}
</style>
