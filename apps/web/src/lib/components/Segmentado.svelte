<script lang="ts" generics="T extends string">
	/** Control segmentado (Hoy / 7 días / 30 días, Simple / Completo…). */
	let {
		opciones,
		value = $bindable(),
		label,
		onchange
	}: {
		opciones: readonly (readonly [T, string])[];
		value: T;
		label: string;
		onchange?: (v: T) => void;
	} = $props();
</script>

<div class="seg" role="group" aria-label={label}>
	{#each opciones as [v, texto] (v)}
		<button
			type="button"
			aria-pressed={value === v}
			onclick={() => {
				value = v;
				onchange?.(v);
			}}>{texto}</button
		>
	{/each}
</div>

<style>
	.seg {
		display: inline-flex;
		gap: 2px;
		padding: 3px;
		border-radius: 999px;
		background: var(--chip);
	}
	button {
		flex: 1;
		padding: 7px 14px;
		border: 0;
		border-radius: 999px;
		background: transparent;
		color: var(--mu);
		font-size: 12.5px;
		font-weight: 600;
		white-space: nowrap;
	}
	button[aria-pressed='true'] {
		background: var(--sf);
		color: var(--tx);
		box-shadow: var(--sh);
	}
	@media (max-width: 760px) {
		button {
			min-height: 38px;
		}
	}
</style>
