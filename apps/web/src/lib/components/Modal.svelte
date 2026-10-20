<script lang="ts">
	import { Dialog } from 'bits-ui';
	import type { Snippet } from 'svelte';

	/**
	 * Modal centrado del handoff: radio 26 px, fondo con blur, acciones a la derecha.
	 * Foco atrapado, Esc y clic afuera cierran (Bits UI).
	 */
	let {
		open = $bindable(false),
		titulo,
		ancho = 460,
		padding = 26,
		children
	}: {
		open?: boolean;
		/** Título accesible (se lee aunque el diseño lo muestre dentro de `children`). */
		titulo: string;
		ancho?: number;
		padding?: number;
		children: Snippet;
	} = $props();
</script>

<Dialog.Root bind:open>
	<Dialog.Portal>
		<Dialog.Overlay class="mo-fondo" />
		<Dialog.Content class="mo-caja" style="--ancho:{ancho}px;--pad:{padding}px">
			<Dialog.Title class="sr-only">{titulo}</Dialog.Title>
			{@render children()}
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

<style>
	:global(.mo-fondo) {
		position: fixed;
		inset: 0;
		z-index: 70;
		background: var(--fondo-modal);
		backdrop-filter: blur(3px);
	}
	:global(.mo-fondo[data-state='open']) {
		animation: mo-fade 260ms both;
	}
	:global(.mo-fondo[data-state='closed']) {
		animation: mo-fade 200ms reverse both;
	}
	:global(.mo-caja) {
		position: fixed;
		left: 50%;
		top: 50%;
		z-index: 71;
		width: var(--ancho);
		max-width: calc(100vw - 32px);
		max-height: calc(100dvh - 32px);
		overflow: auto;
		translate: -50% -50%;
		padding: var(--pad);
		border-radius: 26px;
		background: var(--sf);
		color: var(--tx);
		box-shadow: var(--sombra-modal);
		outline: none;
	}
	:global(.mo-caja[data-state='open']) {
		animation: mo-entra 340ms var(--ease) both;
	}
	:global(.mo-caja[data-state='closed']) {
		animation: mo-entra 200ms var(--ease) reverse both;
	}
	@keyframes mo-fade {
		from {
			opacity: 0;
		}
	}
	@keyframes mo-entra {
		from {
			opacity: 0;
			transform: translateY(10px) scale(0.96);
		}
	}
</style>
