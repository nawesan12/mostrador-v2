<script lang="ts">
	import type { Snippet } from 'svelte';

	/** Estado vacío con Tincho y la primera acción. */
	let {
		hornero = 'horno',
		titulo,
		texto,
		children
	}: { hornero?: string; titulo: string; texto?: string; children?: Snippet } = $props();
</script>

<div class="vacio">
	<div class="ave"><img src="/hornero/{hornero}.webp" alt="" width="96" height="96" /></div>
	<div class="titulo">{titulo}</div>
	{#if texto}<div class="texto">{texto}</div>{/if}
	{#if children}<div class="acciones">{@render children()}</div>{/if}
</div>

<style>
	.vacio {
		padding: 24px 18px;
		border-radius: 20px;
		background: var(--sf);
		text-align: center;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.05);
	}
	.ave {
		width: 96px;
		height: 96px;
		margin: 0 auto;
		border-radius: 50%;
		overflow: hidden;
		background: var(--durazno);
	}
	.ave img {
		display: block;
		width: 100%;
		height: 100%;
		object-fit: contain;
	}
	.titulo {
		margin-top: 12px;
		font-size: 15px;
		font-weight: 600;
	}
	.texto {
		margin-top: 4px;
		font-size: 13px;
		line-height: 1.4;
		color: var(--fa);
	}
	.acciones {
		margin-top: 12px;
		display: flex;
		justify-content: center;
		gap: 8px;
		flex-wrap: wrap;
	}
</style>
