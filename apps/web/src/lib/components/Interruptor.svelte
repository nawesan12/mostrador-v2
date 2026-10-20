<script lang="ts">
	/** Toggle del handoff: 42×24, perilla blanca, terracota encendido. */
	let {
		checked = $bindable(false),
		label,
		disabled = false,
		onchange
	}: {
		checked?: boolean;
		label: string;
		disabled?: boolean;
		onchange?: (v: boolean) => void;
	} = $props();
</script>

<button
	type="button"
	role="switch"
	aria-checked={checked}
	aria-label={label}
	{disabled}
	class="interruptor"
	class:on={checked}
	onclick={() => {
		checked = !checked;
		onchange?.(checked);
	}}
>
	<span class="perilla"></span>
</button>

<style>
	.interruptor {
		position: relative;
		width: 42px;
		height: 24px;
		flex: none;
		padding: 0;
		border: 0;
		border-radius: 999px;
		background: var(--ln2);
		transition: background 0.25s;
	}
	.interruptor.on {
		background: var(--terracota);
	}
	.interruptor:disabled {
		opacity: 0.5;
	}
	.perilla {
		position: absolute;
		top: 2px;
		left: 2px;
		width: 20px;
		height: 20px;
		border-radius: 50%;
		background: #fff;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
		transition: left 0.25s var(--ease);
	}
	.on .perilla {
		left: 20px;
	}
	/* Área de toque de 44 px en mobile sin agrandar el dibujo. */
	@media (max-width: 760px) {
		.interruptor::after {
			content: '';
			position: absolute;
			inset: -10px -1px;
		}
	}
</style>
