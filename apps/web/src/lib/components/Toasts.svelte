<script lang="ts">
	import Icon from './Icon.svelte';
	import MarcaIcono from './MarcaIcono.svelte';
	import { phX } from '#lib/icons.ts';
	import { BARRA } from '#lib/marcas.ts';
	import { getUi, VIDA_TOAST } from '#lib/ui.svelte.ts';

	/** Pila de toasts abajo a la derecha (arriba de la barra inferior en mobile). */
	let { bottom = 20 }: { bottom?: number } = $props();
	const ui = getUi();
</script>

<div class="pila" style:--abajo="{bottom}px" aria-live="polite">
	{#each ui.toasts as t (t.id)}
		<div class="toast" role="status">
			<MarcaIcono k={t.k} />
			<div class="cuerpo">
				<div class="titulo">{t.title}</div>
				{#if t.desc}<div class="desc">{t.desc}</div>{/if}
				{#if t.action}
					<button type="button" class="accion" onclick={() => ui.act(t.id)}>{t.action}</button>
				{/if}
			</div>
			<button
				type="button"
				class="cerrar"
				aria-label="Cerrar aviso"
				onclick={() => ui.dismiss(t.id)}
			>
				<Icon icon={phX} size={14} />
			</button>
			<span
				class="vida"
				style:background={BARRA[t.k] ?? 'var(--terracota)'}
				style:animation-duration="{VIDA_TOAST}ms"
			></span>
		</div>
	{/each}
</div>

<style>
	.pila {
		position: fixed;
		right: 20px;
		bottom: var(--abajo);
		z-index: 80;
		display: flex;
		flex-direction: column;
		gap: 10px;
		align-items: flex-end;
		pointer-events: none;
	}
	.toast {
		pointer-events: auto;
		position: relative;
		overflow: hidden;
		width: min(370px, calc(100vw - 24px));
		display: flex;
		gap: 12px;
		padding: 14px 12px 16px 14px;
		border-radius: 18px;
		background: var(--sf);
		color: var(--tx);
		box-shadow: var(--sombra-toast);
		animation: entra 440ms var(--ease) both;
	}
	.cuerpo {
		flex: 1;
		min-width: 0;
	}
	.titulo {
		font-size: 14px;
		font-weight: 600;
	}
	.desc {
		margin-top: 2px;
		font-size: 13px;
		line-height: 1.4;
		color: var(--mu);
	}
	.accion {
		margin-top: 10px;
		padding: 6px 12px;
		border: 0;
		border-radius: 999px;
		background: var(--chip);
		color: var(--tx);
		font-size: 12.5px;
		font-weight: 600;
		line-height: 1.2;
		white-space: nowrap;
	}
	.cerrar {
		width: 26px;
		height: 26px;
		flex: none;
		display: grid;
		place-items: center;
		border: 0;
		border-radius: 50%;
		background: transparent;
		color: var(--fa);
	}
	.vida {
		position: absolute;
		left: 0;
		bottom: 0;
		height: 3px;
		width: 100%;
		transform-origin: left;
		animation: vida linear both;
	}
	@keyframes entra {
		from {
			opacity: 0;
			transform: translateX(30px) scale(0.97);
		}
	}
	@keyframes vida {
		to {
			transform: scaleX(0);
		}
	}
	@media (max-width: 760px) {
		.pila {
			right: 12px;
			left: 12px;
			align-items: stretch;
		}
		.toast {
			width: auto;
		}
		.cerrar {
			width: 44px;
			height: 44px;
			margin: -10px -8px 0 0;
		}
	}
</style>
