<script lang="ts">
	import { Dialog } from 'bits-ui';
	import type { Snippet } from 'svelte';

	/**
	 * Hoja (sheet) del handoff: lateral derecha flotante en desktop (radio 26, margen 12) y
	 * bottom-sheet con manija en mobile. `cabecera` va fija arriba; `children` scrollea.
	 */
	let {
		open = $bindable(false),
		titulo,
		ancho = 520,
		cabecera,
		pie,
		children
	}: {
		open?: boolean;
		titulo: string;
		ancho?: number;
		cabecera?: Snippet;
		pie?: Snippet;
		children: Snippet;
	} = $props();
</script>

<Dialog.Root bind:open>
	<Dialog.Portal>
		<Dialog.Overlay class="mo-fondo" />
		<Dialog.Content class="hoja" style="--ancho:{ancho}px">
			<span class="hoja-manija" aria-hidden="true"></span>
			<Dialog.Title class="sr-only">{titulo}</Dialog.Title>
			{#if cabecera}<div class="hoja-cab">{@render cabecera()}</div>{/if}
			<div class="hoja-cuerpo" data-lenis-prevent>{@render children()}</div>
			{#if pie}<div class="hoja-pie">{@render pie()}</div>{/if}
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

<style>
	:global(.hoja) {
		position: fixed;
		top: 12px;
		right: 12px;
		bottom: 12px;
		z-index: 71;
		width: var(--ancho);
		max-width: calc(100vw - 24px);
		display: flex;
		flex-direction: column;
		overflow: hidden;
		border-radius: 26px;
		background: var(--bg);
		color: var(--tx);
		box-shadow: var(--sombra-hoja);
		outline: none;
	}
	:global(.hoja[data-state='open']) {
		animation: hoja-entra 480ms var(--ease-hoja) both;
	}
	:global(.hoja[data-state='closed']) {
		animation: hoja-entra 280ms var(--ease) reverse both;
	}
	:global(.hoja-manija) {
		display: none;
	}
	:global(.hoja-cab) {
		padding: 18px 20px;
		display: flex;
		align-items: center;
		gap: 12px;
		background: var(--sf);
		border-bottom: 1px solid var(--ln);
	}
	/* Una cabecera armada en un solo bloque ocupa todo el ancho (el «Cerrar» queda a la derecha). */
	:global(.hoja-cab > :only-child) {
		flex: 1;
		min-width: 0;
	}
	:global(.hoja-cuerpo) {
		flex: 1;
		overflow: auto;
		padding: 16px;
		display: flex;
		flex-direction: column;
		gap: 12px;
		overscroll-behavior: contain;
	}
	:global(.hoja-pie) {
		padding: 14px 16px calc(14px + env(safe-area-inset-bottom));
		display: flex;
		gap: 8px;
		background: var(--sf);
		border-top: 1px solid var(--ln);
	}
	@keyframes hoja-entra {
		from {
			transform: translateX(105%);
		}
	}
	@keyframes hoja-sube {
		from {
			transform: translateY(105%);
		}
	}
	@media (max-width: 760px) {
		:global(.hoja) {
			top: auto;
			left: 0;
			right: 0;
			bottom: 0;
			width: 100%;
			max-width: none;
			max-height: calc(100dvh - 40px);
			border-radius: 26px 26px 0 0;
		}
		:global(.hoja[data-state='open']) {
			animation-name: hoja-sube;
		}
		:global(.hoja[data-state='closed']) {
			animation-name: hoja-sube;
		}
		:global(.hoja-manija) {
			display: block;
			position: absolute;
			top: 7px;
			left: 50%;
			translate: -50% 0;
			width: 40px;
			height: 5px;
			border-radius: 999px;
			background: var(--ln2);
			z-index: 1;
		}
		:global(.hoja-cab) {
			padding-top: 22px;
		}
	}
</style>
