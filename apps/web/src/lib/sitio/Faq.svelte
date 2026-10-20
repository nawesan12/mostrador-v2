<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import { phMinusBold, phPlusBold } from '#lib/icons.ts';

	/** Preguntas frecuentes del sitio: una abierta por vez, la primera de entrada (handoff). */
	let { items }: { items: readonly (readonly [string, string])[] } = $props();
	let abierta = $state(0);
	const uid = $props.id();
</script>

<div class="faq">
	{#each items as [p, r], i (p)}
		{@const on = abierta === i}
		<div class="item">
			<button
				type="button"
				aria-expanded={on}
				aria-controls="{uid}-{i}"
				onclick={() => (abierta = on ? -1 : i)}
			>
				<span>{p}</span><Icon icon={on ? phMinusBold : phPlusBold} size={15} />
			</button>
			{#if on}<div class="resp" id="{uid}-{i}">{r}</div>{/if}
		</div>
	{/each}
</div>

<style>
	.faq {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.item {
		border-radius: 18px;
		background: #fff;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.06);
	}
	button {
		width: 100%;
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 16px 18px;
		border: 0;
		background: transparent;
		text-align: left;
		font-size: 15.5px;
		font-weight: 600;
		color: #2a140c;
	}
	button span {
		flex: 1;
	}
	.resp {
		padding: 0 18px 16px;
		font-size: 14.5px;
		line-height: 1.6;
		color: #5c4536;
	}
</style>
