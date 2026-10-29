<script lang="ts">
	import MarcaIcono from '#lib/components/MarcaIcono.svelte';
	import type { components } from '@mostrador/contrato/esquema';
	import { rutaDe } from './navegacion.ts';
	import { getShell, hace } from './shell.svelte.ts';

	/** Panel de notificaciones (handoff): Todo / Ventas / Sistema y "Marcar todo como leído". */
	let {
		notificaciones,
		marcarLeidas
	}: { notificaciones: components['schemas']['Notificacion'][]; marcarLeidas: () => void } =
		$props();
	const shell = getShell();
	let grupo = $state<'todo' | 'ventas' | 'sistema'>('todo');
	const lista = $derived(notificaciones.filter((n) => grupo === 'todo' || n.grupo === grupo));

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

{#if shell.notificaciones}
	<div class="fuera" onclick={() => (shell.notificaciones = false)} aria-hidden="true"></div>
	<div class="caja" role="dialog" aria-label="Notificaciones" use:entra>
		<div class="cab">
			<span class="t">Notificaciones</span>
			<button type="button" class="leer" onclick={marcarLeidas}>Marcar todo como leído</button>
		</div>
		<div class="tabs" role="group" aria-label="Filtrar">
			{#each [['todo', 'Todo'], ['ventas', 'Ventas'], ['sistema', 'Sistema']] as const as [k, t] (k)}
				<button type="button" aria-pressed={grupo === k} onclick={() => (grupo = k)}>{t}</button>
			{/each}
		</div>
		<div class="lista">
			{#each lista as n (n.id)}
				<div class="n">
					<MarcaIcono k={n.tipo} size={36} radius={11} />
					<span class="n-txt">
						<span class="n-t">{n.titulo}</span>
						<span class="n-d">{n.detalle}</span>
						<span class="n-h">{hace(n.creada)}</span>
					</span>
					{#if !n.leida}<span class="sin-leer" aria-label="Sin leer"></span>{/if}
				</div>
			{:else}
				<div class="vacio">No hay notificaciones por ahora.</div>
			{/each}
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
		max-height: 400px;
		overflow: auto;
		border-top: 1px solid var(--ln);
	}
	.n {
		display: flex;
		gap: 12px;
		padding: 12px 18px;
		border-bottom: 1px solid var(--ln);
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
		color: var(--mu);
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
