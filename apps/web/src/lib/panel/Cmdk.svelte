<script lang="ts">
	import { Dialog } from 'bits-ui';
	import { goto } from '$app/navigation';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { AREAS, rutaDe } from './navegacion.ts';
	import { getShell } from './shell.svelte.ts';

	/**
	 * ⌘K (handoff): buscar y saltar a cualquier sección. Las acciones del handoff (crear pedido, link
	 * de pago, cargar con foto, abrir caja) y "Preguntale a Hornero" se suman con cada función.
	 */
	const shell = getShell();
	let q = $state('');
	let sel = $state(0);
	const items = $derived(
		AREAS.flatMap((a) =>
			a.items.map((i) => ({ t: i.label, s: a.label, icon: i.icon, href: rutaDe(i.label) }))
		).filter((i) => !q.trim() || `${i.t} ${i.s}`.toLowerCase().includes(q.trim().toLowerCase()))
	);
	$effect(() => {
		void q;
		sel = 0;
	});
	$effect(() => {
		if (!shell.cmdk) q = '';
	});

	function ir(href: string) {
		shell.cmdk = false;
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- rutas del panel armadas con rutaDe
		goto(href);
	}
	function teclas(e: KeyboardEvent) {
		if (e.key === 'ArrowDown') {
			e.preventDefault();
			sel = Math.min(items.length - 1, sel + 1);
			document.getElementById(`cmdk-${sel}`)?.scrollIntoView({ block: 'nearest' });
		} else if (e.key === 'ArrowUp') {
			e.preventDefault();
			sel = Math.max(0, sel - 1);
			document.getElementById(`cmdk-${sel}`)?.scrollIntoView({ block: 'nearest' });
		} else if (e.key === 'Enter' && items[sel]) {
			e.preventDefault();
			ir(items[sel].href);
		}
	}
</script>

<svelte:window
	onkeydown={(e) => {
		if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
			e.preventDefault();
			shell.cmdk = !shell.cmdk;
		}
	}}
/>

<Dialog.Root bind:open={shell.cmdk}>
	<Dialog.Portal>
		<Dialog.Overlay class="cmdk-fondo" />
		<Dialog.Content class="cmdk" aria-describedby={undefined}>
			<Dialog.Title class="sr-only">Buscar</Dialog.Title>
			<div class="cab">
				<Icon icon={I.phMagnifyingGlass} size={20} style="color:var(--mu)" />
				<input
					bind:value={q}
					onkeydown={teclas}
					placeholder="Buscá una sección del panel…"
					aria-label="Buscar"
					aria-controls="cmdk-lista"
					aria-activedescendant={items[sel] ? `cmdk-${sel}` : undefined}
				/>
				<kbd>esc</kbd>
			</div>
			<div class="lista" id="cmdk-lista" role="listbox" aria-label="Resultados">
				{#if items.length}<div class="g">Ir a</div>{/if}
				{#each items as it, i (it.href)}
					<button
						type="button"
						id="cmdk-{i}"
						role="option"
						aria-selected={i === sel}
						class:sel={i === sel}
						onmouseenter={() => (sel = i)}
						onclick={() => ir(it.href)}
					>
						<span class="ico"><Icon icon={it.icon} size={17} /></span>
						<span class="txt"><span class="t">{it.t}</span><span class="s">{it.s}</span></span>
						<kbd>↵</kbd>
					</button>
				{:else}
					<div class="nada">Nada con «{q}».</div>
				{/each}
			</div>
			<div class="pie">
				<span>↑↓ navegar</span><span>↵ abrir</span><span>esc cerrar</span><span class="esp"
				></span><span>⌘K desde cualquier pantalla</span>
			</div>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

<style>
	:global(.cmdk-fondo) {
		position: fixed;
		inset: 0;
		z-index: 70;
		background: rgba(42, 20, 12, 0.34);
		backdrop-filter: blur(3px);
	}
	:global(.cmdk) {
		position: fixed;
		left: 0;
		right: 0;
		margin: 0 auto;
		top: 11vh;
		z-index: 71;
		width: min(640px, calc(100vw - 24px));
		border-radius: 24px;
		background: var(--sf);
		color: var(--tx);
		box-shadow:
			0 40px 100px -20px rgba(42, 20, 12, 0.6),
			0 0 0 1px var(--ln);
		overflow: hidden;
		outline: none;
	}
	:global(.cmdk[data-state='open']) {
		animation: cmdk-entra 0.28s cubic-bezier(0.2, 0.7, 0.1, 1) both;
	}
	@keyframes cmdk-entra {
		from {
			opacity: 0;
			transform: translateY(8px) scale(0.98);
		}
	}
	.cab {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 16px 18px;
		border-bottom: 1px solid var(--ln);
	}
	.cab input {
		flex: 1;
		min-width: 0;
		border: 0;
		outline: 0;
		background: transparent;
		font-size: 17px;
		font-weight: 500;
		color: var(--tx);
	}
	kbd {
		font-family: inherit;
		font-size: 12px;
		font-weight: 600;
		padding: 3px 8px;
		border-radius: 6px;
		background: var(--chip);
		color: var(--mu);
		white-space: nowrap;
	}
	.lista {
		max-height: 440px;
		overflow: auto;
		padding: 8px;
	}
	.g {
		padding: 12px 12px 4px;
		font-size: 12px;
		font-weight: 600;
		color: var(--fa);
	}
	.lista button {
		width: 100%;
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 10px 12px;
		border-radius: 12px;
		border: 0;
		background: transparent;
		color: var(--tx);
		text-align: left;
	}
	.lista button.sel {
		background: var(--chip);
	}
	.ico {
		width: 32px;
		height: 32px;
		flex: none;
		border-radius: 10px;
		display: grid;
		place-items: center;
		background: rgba(210, 88, 47, 0.1);
		color: #c24e25;
	}
	.txt {
		flex: 1;
		min-width: 0;
	}
	.t {
		display: block;
		font-size: 14px;
		font-weight: 600;
	}
	.s {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
	}
	.nada {
		padding: 28px;
		text-align: center;
		font-size: 14px;
		color: var(--mu);
	}
	.pie {
		display: flex;
		gap: 16px;
		padding: 10px 18px;
		border-top: 1px solid var(--ln);
		font-size: 12px;
		color: var(--fa);
	}
	.esp {
		flex: 1;
	}
	@media (max-width: 759px) {
		:global(.cmdk) {
			top: 12px;
		}
		.pie {
			display: none;
		}
	}
</style>
